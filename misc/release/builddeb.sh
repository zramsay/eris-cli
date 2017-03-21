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

export GOREPO=${GOPATH}/src/github.com/eris-ltd/eris
git clone https://github.com/eris-ltd/eris ${GOREPO}
pushd ${GOREPO}/cmd/eris
git fetch origin ${ERIS_BRANCH}
git checkout ${ERIS_BRANCH}
echo
echo ">>> Building the Eris binary"
echo
go get
go build -ldflags "-X github.com/eris-ltd/eris/version.COMMIT=`git rev-parse --short HEAD 2>/dev/null`"
popd

echo
echo ">>> Building the Debian package (#${ERIS_BRANCH})"
echo
mkdir -p deb/usr/bin deb/usr/share/doc/eris deb/usr/share/man/man1 deb/DEBIAN
cp ${GOREPO}/cmd/eris/eris deb/usr/bin
${GOREPO}/cmd/eris/eris man --dump > deb/usr/share/man/man1/eris.1
cat > deb/DEBIAN/control <<EOF
Package: eris
Version: ${ERIS_VERSION}-${ERIS_RELEASE}
Section: devel
Architecture: amd64
Priority: standard
Homepage: https://monax.io/docs
Maintainer: Monax Industries <support@monax.io>
Build-Depends: debhelper (>= 9.1.0), golang-go (>= 1.6)
Standards-Version: 3.9.4
Description: ecosystem application platform for building, testing, maintaining, and operating distributed applications. 
EOF
cp ${GOREPO}/README.md deb/usr/share/doc/eris/README
cat > deb/usr/share/doc/eris/copyright <<EOF
Files: *
Copyright: $(date +%Y) Monax Industries  <support@monax.io>
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
s3cmd put ${PACKAGE} s3://${AWS_S3_DEB_FILES}

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
Origin: Monax Industries <support@monax.io>
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
echo "  \$ sudo add-apt-repository http://${AWS_S3_DEB_REPO}"
echo "  \$ curl -L http://${AWS_S3_DEB_REPO}/APT-GPG-KEY | sudo apt-key add -"
echo
echo "  \$ sudo apt-get update"
echo "  \$ sudo apt-get install eris"
echo
