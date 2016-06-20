#!/usr/bin/env bash
#
# Eris CLI Github and Linux packages release script.
#
# Prerequisites:
#
#  1. For full release -- release tagged (`git tag`) and master branch
#     checked out.
#
#  2. `github-release` utility installed (go get github.com/aktau/github-release)
#     and GITHUB_TOKEN environment variable set
#    (with release permissions for github.com/eris-ltd/eris-cli).
#
#  3. GPG release signing private key in `misc/release/linux-private-key.asc` file:
#
#    $ gpg2 --export-secret-keys -a KEYID > linux-private-key.asc
#
#  4. GPG release signing public key in `misc/release/linux-public-key.asc` file:
#
#    $ gpg2 --export -a KEYID > linux-public-key.asc
#
#  5. GPG release signing key name in KEY_NAME variable:
#
#    KEY_NAME="Eris Industries <support@erisindustries.com>"
#
#  6. GPG release signing key password in KEY_PASSWORD variable:
#
#    KEY_PASSWORD="*****"
#
#  7. Amazon AWS credentials environment variables set:
#
#    AWS_ACCESS_KEY=*****
#    AWS_SECRET_ACCESS_KEY=*****
#
#  8. Environment variables pointing to four S3 buckets with public access policy:
#
#    Use bucket names only, without s3:// prefix or s3.amazonaws.com paths.
#
#    AWS_S3_RPM_REPO                -- YUM master repository bucket
#    AWS_S3_RPM_PACKAGES            -- RPM downloadable packages bucket
#    AWS_S3_DEB_REPO                -- APT master repository bucket
#    AWS_S3_DEB_PACKAGES            -- Debian downloadable packages bucket
#
#      Copy pastable sample for public access policy:
#
#         {
#           "Version":"2012-10-17",
#           "Statement":[
#             {
#               "Sid":"AddPerm",
#               "Effect":"Allow",
#               "Principal": "*",
#               "Action":["s3:GetObject"],
#               "Resource":["arn:aws:s3:::examplebucket/*"]
#             }
#           ]
#         }
#
REPO=${GOPATH}/src/github.com/eris-ltd/eris-cli
BUILD_DIR=${REPO}/builds
ERIS_VERSION=$(grep -w VERSION ${REPO}/version/version.go | cut -d \  -f 4 | tr -d '"')
ERIS_RELEASE=1
# NOTE: Please, set these up before continuing:
# 
# AWS_ACCESS_KEY=
# AWS_SECRET_ACCESS_KEY=
AWS_S3_RPM_REPO=eris-rpm
AWS_S3_RPM_PACKAGES=eris-rpm-files
AWS_S3_DEB_REPO=eris-deb
AWS_S3_DEB_PACKAGES=eris-deb-files
KEY_NAME="Eris Industries (DISTRIBUTION SIGNATURE MASTER KEY) <support@erisindustries.com>"
KEY_PASSWORD="one1two!three"

pre_check() {
  read -p "Have you done the [git tag -a v${ERIS_VERSION}] and filled out the changelog yet? (y/n) " -n 1 -r
  echo
  if [[ ! ${REPLY} =~ ^[Yy]$ ]]
  then
    echo "OK. Not doing anything. Rerun me after you've done that"
    exit 1
  fi
  echo "OK. Moving on then"
  echo ""
  echo ""
  LATEST_TAG=$(git tag | xargs -I@ git log --format=format:"%ai @%n" -1 @ | sort | awk '{print $4}' | tail -n 1 | cut -c 2-)
  if [[ "${LATEST_TAG}" != "${ERIS_VERSION}}" ]]
  then
    echo "Something isn't right. The last tagged version does not match the version to be released"
    echo "Last tagged: ${LATEST_TAG}"
    echo "This version: ${ERIS_VERSION}"
    exit 1
  fi
}

keys_check() {
  if [ -z "${AWS_ACCESS_KEY}" -o -z "${AWS_SECRET_ACCESS_KEY}" ]
  then
    echo "Amazon AWS credentials should be set to proceed"
    exit 1
  fi
  if [ -z "${KEY_NAME}" -o -z "${KEY_PASSWORD}" ]
  then
    echo "GPG key name and password should be set to proceed"
    exit 1
  fi
  if [ ! -r ${REPO}/misc/release/linux-private-key.asc -o ! -r ${REPO}/misc/release/linux-public-key.asc ]
  then
    echo "GPG key file(s) linux-private-key.asc or linux-public-key.asc are missing"
    exit 1
  fi
  if [ -z "${AWS_S3_RPM_PACKAGES}" -o -z "${AWS_S3_DEB_PACKAGES}" ]
  then
    echo "Amazon S3 buckets have to be set to proceed"
    exit 1
  fi
}

cross_compile() {
  echo "Starting cross compile"
  pushd ${REPO}/cmd/eris
  GOOS=linux   GOARCH=386    go build -o ${BUILD_DIR}/eris_${ERIS_VERSION}_linux_386
  GOOS=linux   GOARCH=amd64  go build -o ${BUILD_DIR}/eris_${ERIS_VERSION}_linux_amd64
  GOOS=darwin  GOARCH=386    go build -o ${BUILD_DIR}/eris_${ERIS_VERSION}_darwin_386
  GOOS=darwin  GOARCH=amd64  go build -o ${BUILD_DIR}/eris_${ERIS_VERSION}_darwin_amd64
  GOOS=windows GOARCH=386    go build -o ${BUILD_DIR}/eris_${ERIS_VERSION}_windows_386.exe
  GOOS=windows GOARCH=amd64  go build -o ${BUILD_DIR}/eris_${ERIS_VERSION}_windows_amd64.exe
  popd
  echo "Cross compile completed"
  echo ""
  echo ""
}

prepare_gh() {
  echo "Pushing tags to Github and creating a Github release"
  git push origin --tags
  DESCRIPTION=$(git show v${ERIS_VERSION})
  if [[ "$1" == "pre" ]]
  then
    github-release release \
      --user eris-ltd \
      --repo eris-cli \
      --tag v${ERIS_VERSION} \
      --name "Release of Version: ${ERIS_VERSION}" \
      --description "${DESCRIPTION}" \
      --pre-release
  else
    github-release release \
      --user eris-ltd \
      --repo eris-cli \
      --tag v${ERIS_VERSION} \
      --name "Release of Version: ${ERIS_VERSION}" \
      --description "${DESCRIPTION}"
  fi
  echo "Finished sending tags and release info to Github"
  echo ""
  echo ""
}

release_gh() {
  echo "Uploading binaries to Github"
  for file in ${BUILD_DIR}/*
  do
    echo "Uploading: ${file}"
    github-release upload \
      --user eris-ltd \
      --repo eris-cli \
      --tag v${ERIS_VERSION} \
      --name ${file} \
      --file ${file}
  done
  echo "Uploading completed"
  echo ""
  echo ""
}

release_deb() {
  echo "Releasing Debian packages"
  shift
  mkdir -p ${BUILD_DIR}

  if [ ! -z "$@" ]
  then
    ERIS_RELEASE="$@"
  fi

  docker rm -f builddeb 2>&1 >/dev/null
  docker build -f ${REPO}/misc/release/Dockerfile-deb -t builddeb ${REPO}/misc/release \
  && docker run \
    -t \
    --name builddeb \
    -e ERIS_VERSION=${ERIS_VERSION} \
    -e ERIS_RELEASE=${ERIS_RELEASE} \
    -e AWS_ACCESS_KEY=${AWS_ACCESS_KEY} \
    -e AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} \
    -e AWS_S3_RPM_REPO=${AWS_S3_RPM_REPO} \
    -e AWS_S3_RPM_PACKAGES=${AWS_S3_RPM_PACKAGES} \
    -e AWS_S3_DEB_REPO=${AWS_S3_DEB_REPO} \
    -e AWS_S3_DEB_PACKAGES=${AWS_S3_DEB_PACKAGES} \
    -e KEY_NAME="${KEY_NAME}" \
    -e KEY_PASSWORD="${KEY_PASSWORD}" \
    builddeb "$@" \
  && docker cp builddeb:/root/eris_${ERIS_VERSION}-${ERIS_RELEASE}_amd64.deb ${BUILD_DIR} \
  && docker rm -f builddeb
  echo "Finished releasing Debian packages"
}

release_rpm() {
  echo "Releasing RPM packages"
  shift
  mkdir -p ${BUILD_DIR}

  if [ ! -z "$@" ]
  then
    ERIS_RELEASE="$@"
  fi

  docker rm -f buildrpm 2>&1 >/dev/null
  docker build -f ${REPO}/misc/release/Dockerfile-rpm -t buildrpm ${REPO}/misc/release \
  && docker run \
    -t \
    --name buildrpm \
    -e ERIS_VERSION=${ERIS_VERSION} \
    -e ERIS_RELEASE=${ERIS_RELEASE} \
    -e AWS_ACCESS_KEY=${AWS_ACCESS_KEY} \
    -e AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} \
    -e AWS_S3_RPM_REPO=${AWS_S3_RPM_REPO} \
    -e AWS_S3_RPM_PACKAGES=${AWS_S3_RPM_PACKAGES} \
    -e AWS_S3_DEB_REPO=${AWS_S3_DEB_REPO} \
    -e AWS_S3_DEB_PACKAGES=${AWS_S3_DEB_PACKAGES} \
    -e KEY_NAME="${KEY_NAME}" \
    -e KEY_PASSWORD="${KEY_PASSWORD}" \
    buildrpm "$@" \
  && docker cp buildrpm:/root/rpmbuild/RPMS/x86_64/eris-cli-${ERIS_VERSION}-${ERIS_RELEASE}.x86_64.rpm ${BUILD_DIR} \
  && docker rm -f buildrpm
  echo "Finished releasing RPM packages"
}

clean_up() {
  echo "Cleaning up and exiting... Billings Shipit!"
  rm -rf ${BUILD_DIR}
}

usage() {
  echo "Usage: release.sh [pre|build|pkgs|rpm|deb|help]"
  echo "Release Eris CLI to Github. Publish Linux packages to Amazon S3"
  echo
  echo "   release.sh              release #master"
  echo "   release.sh pre          prerelease #master"
  echo "   release.sh build        cross compile current branch "
  echo "                           for all supported architectures"
  echo "   release.sh pkgs         cross compile current branch"
  echo "                           and publish Linux packages"
  echo "   release.sh deb          publish Debian package and create APT repo"
  echo "   release.sh rpm          publish RPM package and create YUM repo"
  echo "   release.sh deb develop  publish Debian package for the #develop branch"
  echo "   release.sh rpm develop  publish RPM package for the #develop branch"
  echo
  exit 2
}

main() {
  case $1 in
  build)
    cross_compile "$@"
    ;;
  pkgs)
    keys_check "$@"
    release_deb "$@"
    release_rpm "$@"
    ;;
  rpm)
    keys_check "$@"
    release_rpm "$@"
    ;;
  deb)
    keys_check "$@"
    release_deb "$@"
    ;;
  help|-h|--help)
    usage "$@"
    ;;
  *)
    pre_check "$@"
    keys_check "$@"
    cross_compile "$@"
    prepare_gh "$@"
    release_gh "$@"
    release_deb "$@"
    release_rpm "$@"
    clean_up $?
  esac
  return $?
}

main "$@"
