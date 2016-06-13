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
  ERIS_BRANCH="$1"
fi

echo
echo ">>> Importing GPG keys"
echo
gpg --import linux-public-key.asc
gpg --import linux-private-key.asc

export GOREPO=${GOPATH}/src/github.com/eris-ltd/eris-cli
git clone https://github.com/eris-ltd/eris-cli ${GOREPO}
pushd ${GOREPO}/cmd/eris
git fetch origin ${ERIS_BRANCH}
git checkout ${ERIS_BRANCH}
echo
echo ">>> Building the Eris binary"
echo
go get
go build
popd

echo
echo ">>> Building the Debian package (#${ERIS_BRANCH})"
echo
mkdir -p deb/usr/bin deb/usr/share/doc/eris deb/DEBIAN
cp ${GOREPO}/cmd/eris/eris deb/usr/bin
cat > deb/DEBIAN/control <<EOF
Package: eris
Version: ${ERIS_VERSION}-${ERIS_RELEASE}
Section: devel
Architecture: amd64
Priority: standard
Homepage: https://docs.erisindustries.com
Maintainer: Eris Industries <support@erisindustries.com>
Build-Depends: debhelper (>= 9.1.0), golang-go (>= 1.6)
Standards-Version: 3.9.4
Description: platform for building, testing, maintaining, and operating
  distributed applications with a blockchain backend. Eris makes it easy
  and simple to wrangle the dragons of smart contract blockchains.
EOF
# TODO: manual page addition is pending the issue
# https://github.com/eris-ltd/eris-cli/issues/712.
cp ${GOREPO}/README.md deb/usr/share/doc/eris/README
cat > deb/usr/share/doc/eris/copyright <<EOF
Files: *
Copyright: $(date +%Y) Eris Industries, Ltd. <support@erisindustries.com>
License: GPL-3
EOF
dpkg-deb --build deb
PACKAGE=eris_${ERIS_VERSION}-${ERIS_RELEASE}_amd64.deb
mv deb.deb ${PACKAGE}

echo
echo ">>> Copying Debian packages to Amazon S3"
echo
cat > ${HOME}/.s3cfg <<EOF
[default]
access_key = ${AWS_ACCESS_KEY}
secret_key = ${AWS_SECRET_ACCESS_KEY}
EOF
s3cmd put ${PACKAGE} s3://${AWS_S3_DEB_PACKAGES}

if [ "$ERIS_BRANCH" != "master" ]
then
   echo
   echo ">>> Not recreating a repo for #${ERIS_BRANCH} branch"
   echo
   exit 0
fi

echo
echo ">>> Creating an APT repository"
echo
mkdir -p eris/conf
gpg --armor --export "${KEY_NAME}" > eris/APT-GPG-KEY

cat > eris/conf/options <<EOF
verbose
basedir /root/eris
ask-passphrase
EOF

DISTROS="precise trusty utopic vivid wheezy jessie stretch wily xenial"
for distro in ${DISTROS}
do
  cat >> eris/conf/distributions <<EOF
Origin: Eris Industries <support@erisindustries.com>
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
    spawn reprepro -Vb eris includedeb ${distro} ${PACKAGE}
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
reprepro -b eris ls eris

echo
echo ">>> Syncing repos to Amazon S3"
echo
s3cmd sync eris/APT-GPG-KEY s3://${AWS_S3_DEB_REPO}
s3cmd sync eris/db s3://${AWS_S3_DEB_REPO}
s3cmd sync eris/dists s3://${AWS_S3_DEB_REPO}
s3cmd sync eris/pool s3://${AWS_S3_DEB_REPO}

echo
echo ">>> Installation instructions"
echo
echo "  \$ curl https://${AWS_S3_DEB_REPO}.s3.amazonaws/APT-GPG-KEY | apt-key add -"
echo "  \$ echo \"deb https://eris-deb-repo.s3.amazonaws.com DIST main\" > /etc/apt/sources.list.d"
echo
echo "  \$ apt-get update"
echo "  \$ apt-get install eris"
echo
