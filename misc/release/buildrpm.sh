#! /bin/bash

echo
echo ">>> Sanity checks"
echo
if [ -z "${ERIS_VERSION}" -o -z "${ERIS_RELEASE}" ]
then
    echo "The ERIS_VERSION or ERIS_RELEASE environment variables are not set, aborting"
    echo
    echo "Please start this container from the 'release.sh' script"
    exit 1
fi

export ERIS_BRANCH=master
if [ ! -z "$1" ]; then
  export ERIS_BRANCH="$1"
fi

echo
echo ">>> Importing GPG keys"
echo
gpg2 --import linux-public-key.asc
expect <<EOF
spawn gpg2 --import linux-private-key.asc
expect "Really update the preferences?"
send "y\r"
interact
EOF

echo
echo ">>> Building and signing the RPM package (#${ERIS_BRANCH})"
echo
rpmdev-setuptree

cat > $HOME/.rpmmacros <<EOF
%_signature gpg
%_gpg_path $HOME/.gnupg
%_gpg_name ${KEY_NAME}
%_gpgbin %{_bindir}/gpg2
EOF

expect <<EOF
set timeout 300
spawn rpmbuild -ba --sign rpmbuild/SPECS/eris-cli.spec
expect {
    timeout              { send_error "Failed to submit password"; exit 1 }
    "Enter pass phrase:" { send -- "${KEY_PASSWORD}\r";
                           send_user "********";
                           exp_continue
                         }
}
wait
exit 0
EOF

echo
echo ">>> Copying RPM packages to Amazon S3"
echo
cat > $HOME/.s3cfg <<EOF
[default]
access_key = ${AWS_ACCESS_KEY}
secret_key = ${AWS_SECRET_ACCESS_KEY}
EOF

s3cmd put rpmbuild/RPMS/x86_64/* s3://${AWS_S3_RPM_PACKAGES}
s3cmd put rpmbuild/SRPMS/* s3://${AWS_S3_RPM_PACKAGES}

if [ "$ERIS_BRANCH" != "master" ]
then
   echo
   echo ">>> Not recreating a repo for #${ERIS_BRANCH} branch"
   echo
   exit 0
fi

echo
echo ">>> Creating repos"
echo
mkdir eris eris/x86_64 eris/source
gpg2 --armor --export "${KEY_NAME}" > eris/RPM-GPG-KEY
cp rpmbuild/RPMS/x86_64/* eris/x86_64
cp rpmbuild/SRPMS/* eris/source
createrepo eris/x86_64
createrepo eris/source

echo
echo ">>> Verifying the signature"
echo
rpm --import eris/RPM-GPG-KEY
rpm --checksig eris/x86_64/*rpm eris/source/*rpm

echo
echo ">>> Syncing repos to Amazon S3"
echo
s3cmd sync eris s3://${AWS_S3_RPM_REPO}

echo
echo ">>> Installation instructions"
echo
echo "Create a file named /etc/yum.repos.d/eris.repo with the following contents"
cat <<EOF

[eris]
name=Eris
baseurl=https://${AWS_S3_RPM_REPO}.s3.amazonaws.com/eris/x86_64/
metadata_expire=1d
enabled=1
gpgkey=http://${AWS_S3_RPM_REPO}.s3.amazonaws.com/eris/RPM-GPG-KEY
gpgcheck=1

[eris-source]
name=Eris Source
baseurl=http://${AWS_S3_RPM_REPO}.s3.amazonaws.com/eris/source/
metadata_expire=1d
enabled=1
gpgkey=http://${AWS_S3_RPM_REPO}.s3.amazonaws.com/eris/RPM-GPG-KEY
gpgcheck=1
EOF
echo
echo "  \$ yum update"
echo "  \$ yum install eris-cli"
echo
