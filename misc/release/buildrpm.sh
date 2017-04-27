#! /bin/bash

echo
echo ">>> Sanity checks"
echo
if [ -z "${MONAX_VERSION}" -o -z "${MONAX_RELEASE}" -o -z "${MONAX_BRANCH}" ]
then
    echo "The MONAX_VERSION, MONAX_RELEASE or MONAX_BRANCH environment variables are not set, aborting"
    echo
    echo "Please start this container from the 'release.sh' script"
    exit 1
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

cat > /root/.rpmmacros <<EOF
%_signature gpg
%_gpg_path /root/.gnupg
%_gpg_name ${KEY_NAME}
%_gpgbin %{_bindir}/gpg2
EOF

expect <<EOF
set timeout 300
spawn rpmbuild -ba --sign /root/rpmbuild/SPECS/monax.spec
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
echo ">>> Creating repos"
echo
mkdir yum yum/x86_64 yum/source
gpg2 --armor --export "${KEY_NAME}" > yum/RPM-GPG-KEY
cp /root/rpmbuild/RPMS/x86_64/* yum/x86_64
cp /root/rpmbuild/SRPMS/* yum/source
createrepo yum/x86_64
createrepo yum/source

echo
echo ">>> Verifying the signature"
echo
rpm --import yum/RPM-GPG-KEY
rpm --checksig yum/x86_64/*rpm yum/source/*rpm

echo
echo ">>> Generating monax.repo template"
echo
cat > yum/monax.repo <<EOF
[monax]
name=Monax
baseurl=https://${AWS_S3_PKGS_URL}/yum/x86_64/
metadata_expire=1d
enabled=1
gpgkey=https://${AWS_S3_PKGS_URL}/yum/RPM-GPG-KEY
gpgcheck=1

[monax-source]
name=Monax Source
baseurl=https://${AWS_S3_PKGS_URL}/yum/source/
metadata_expire=1d
enabled=1
gpgkey=https://${AWS_S3_PKGS_URL}/yum/RPM-GPG-KEY
gpgcheck=1
EOF

echo
echo ">>> Syncing repos to Amazon S3"
echo
aws s3 sync yum s3://${AWS_S3_PKGS_BUCKET}/yum/ --acl public-read

echo
echo ">>> Installation instructions"
echo
echo "  \$ sudo curl -L https://${AWS_S3_PKGS_URL}/yum/monax.repo >/etc/yum.repos.d/monax.repo"
echo
echo "  \$ sudo yum update"
echo "  \$ sudo yum install monax"
echo
