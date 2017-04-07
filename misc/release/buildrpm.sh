#! /bin/bash

echo
echo ">>> Sanity checks"
echo
if [ -z "${MONAX_VERSION}" -o -z "${MONAX_RELEASE}" ]
then
    echo "The MONAX_VERSION or MONAX_RELEASE environment variables are not set, aborting"
    echo
    echo "Please start this container from the 'release.sh' script"
    exit 1
fi

export MONAX_BRANCH=master
if [ ! -z "$1" ]; then
  export MONAX_BRANCH="$1"
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
echo ">>> Building and signing the RPM package (#${MONAX_BRANCH})"
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
spawn rpmbuild -ba --sign rpmbuild/SPECS/eris.spec
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

s3cmd put rpmbuild/RPMS/x86_64/* s3://${AWS_S3_RPM_FILES}
s3cmd put rpmbuild/SRPMS/* s3://${AWS_S3_RPM_FILES}

if [ "$MONAX_BRANCH" != "master" ]
then
   echo
   echo ">>> Not recreating a repo for #${MONAX_BRANCH} branch"
   echo
   exit 0
fi

echo
echo ">>> Creating repos"
echo
mkdir yum yum/x86_64 yum/source
gpg2 --armor --export "${KEY_NAME}" > yum/RPM-GPG-KEY
cp rpmbuild/RPMS/x86_64/* yum/x86_64
cp rpmbuild/SRPMS/* yum/source
createrepo yum/x86_64
createrepo yum/source

echo
echo ">>> Verifying the signature"
echo
rpm --import yum/RPM-GPG-KEY
rpm --checksig yum/x86_64/*rpm yum/source/*rpm

echo
echo ">>> Generating eris.repo template"
echo
cat > yum/eris.repo <<EOF
[eris]
name=Monax
baseurl=https://${AWS_S3_RPM_REPO}/yum/x86_64/
metadata_expire=1d
enabled=1
gpgkey=https://${AWS_S3_RPM_REPO}/yum/RPM-GPG-KEY
gpgcheck=1

[eris-source]
name=Monax Source
baseurl=https://${AWS_S3_RPM_REPO}/yum/source/
metadata_expire=1d
enabled=1
gpgkey=https://${AWS_S3_RPM_REPO}/yum/RPM-GPG-KEY
gpgcheck=1
EOF

echo
echo ">>> Syncing repos to Amazon S3"
echo
s3cmd sync yum s3://${AWS_S3_RPM_REPO}

echo
echo ">>> Installation instructions"
echo
echo "  \$ sudo curl -L https://${AWS_S3_RPM_REPO}/yum/eris.repo >/etc/yum.repos.d/eris.repo"
echo
echo "  \$ sudo yum update"
echo "  \$ sudo yum install eris"
echo
