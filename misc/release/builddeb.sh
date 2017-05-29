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
gpg --import linux-public-key.asc
gpg --import linux-private-key.asc
echo "digest-algo sha256" >> $HOME/.gnupg/gpg.conf

export GOREPO=${GOPATH}/src/github.com/monax/monax
git clone https://github.com/monax/monax ${GOREPO}
pushd ${GOREPO}/cmd/monax
git fetch origin ${MONAX_BRANCH}
git checkout ${MONAX_BRANCH}
echo
echo ">>> Building the Monax binary"
echo
go get
go build -ldflags "-X github.com/monax/monax/version.COMMIT=`git rev-parse --short HEAD 2>/dev/null`"
popd

echo
echo ">>> Building the Debian package (#${MONAX_BRANCH})"
echo
mkdir -p deb/usr/bin deb/usr/share/doc/monax deb/usr/share/man/man1 deb/DEBIAN
cp ${GOREPO}/cmd/monax/monax deb/usr/bin
${GOREPO}/cmd/monax/monax man --dump > deb/usr/share/man/man1/monax.1
cat > deb/DEBIAN/control <<EOF
Package: monax
Version: ${MONAX_VERSION}-${MONAX_RELEASE}
Section: devel
Architecture: amd64
Priority: standard
Homepage: https://monax.io/docs
Maintainer: Monax <support@monax.io>
Build-Depends: debhelper (>= 9.1.0), golang-go (>= 1.6)
Standards-Version: 3.9.4
Description: ecosystem application platform for building, testing, maintaining, and operating distributed applications.
EOF
cp ${GOREPO}/README.md deb/usr/share/doc/monax/README
cat > deb/usr/share/doc/monax/copyright <<EOF
Files: *
Copyright: $(date +%Y) Monax <support@monax.io>
License: GPL-3
EOF
dpkg-deb --build deb
PACKAGE=monax_${MONAX_VERSION}-${MONAX_RELEASE}_amd64.deb
mv deb.deb ${PACKAGE}
aws s3 cp ${PACKAGE} s3://${AWS_S3_PKGS_BUCKET}/dl/${PACKAGE} --acl public-read

echo
echo ">>> Creating an APT repository"
echo
mkdir -p apt/conf
gpg --armor --export "${KEY_NAME}" > apt/APT-GPG-KEY

cat > apt/conf/options <<EOF
verbose
basedir /root/monax
ask-passphrase
EOF

DISTROS="precise trusty utopic vivid wheezy jessie stretch wily xenial yakkety"
for distro in ${DISTROS}
do
  cat >> apt/conf/distributions <<EOF
Origin: Monax <support@monax.io>
Codename: ${distro}
Components: main
Architectures: i386 amd64
SignWith: $(gpg --keyid-format=long --list-keys --with-colons|fgrep "${KEY_NAME}"|cut -d: -f5)

EOF
done

for distro in ${DISTROS}
do
  echo
  echo ">>> Adding package to ${distro}"
  echo
  expect <<-EOF
    set timeout 5
    spawn reprepro -Vb apt includedeb ${distro} ${PACKAGE}
    expect {
            timeout                    { send_error "Failed to submit password"; exit 1 }
            "Please enter passphrase:" { send -- "${KEY_PASSWORD}\r";
                                         send_user "********";
                                         exp_continue
                                       }
    }
    wait
    exit 0
EOF
done

echo
echo ">>> After adding we have the following"
echo
reprepro -b apt ls monax

echo
echo ">>> Syncing repos to Amazon S3"
echo
aws s3 sync apt s3://${AWS_S3_PKGS_BUCKET}/apt/ --acl public-read

echo
echo ">>> Installation instructions"
echo
echo "  \$ sudo add-apt-repository https://${AWS_S3_PKGS_URL}/apt"
echo "  \$ curl -L https://${AWS_S3_PKGS_URL}/apt/APT-GPG-KEY | sudo apt-key add -"
echo
echo "  \$ sudo apt-get update"
echo "  \$ sudo apt-get install monax"
echo
