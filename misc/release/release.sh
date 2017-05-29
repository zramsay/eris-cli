#!/usr/bin/env bash
#
# Monax CLI Github and Linux packages release script.
#
# Prerequisites:
#
#  1. For full release -- release tagged (`git tag`) and master branch
#     checked out.
#
#  2. `github-release` utility installed (go get github.com/aktau/github-release)
#     and GITHUB_TOKEN environment variable set
#    (with release permissions for github.com/monax/monax).
#
#  2.a `xgo` installed for cross-compilation:
#
#   docker pull karalabe/xgo-1.7
#   go get github.com/karalabe/xgo
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
#    KEY_NAME="Monax <ops@monax.io>"
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
#  8. Environment variables pointing to s3 buckets
#
REPO=${GOPATH}/src/github.com/monax/monax
BUILD_DIR=${REPO}/builds
vers=$(grep -w VERSION ${REPO}/version/version.go | cut -d \  -f 4 | tr -d '"')
MONAX_VERSION=${MONAX_VERSION:-$vers}
MONAX_VERSION_MINOR=$(grep -w VERSION version/version.go | cut -d \  -f 4 | tr -d '"' | cut -d . -f 1,2)
LATEST_TAG=$(git tag | xargs -I@ git log --format=format:"%ai @%n" -1 @ | sort | awk '{print $4}' | tail -n 1 | cut -c 2-)
MONAX_RELEASE=${MONAX_RELEASE:-1}
MONAX_BRANCH=${MONAX_BRANCH:-release-$MONAX_VERSION_MINOR}

# NOTE: Set these up before continuing:
# export GITHUB_TOKEN=
# export AWS_ACCESS_KEY=
# export AWS_SECRET_ACCESS_KEY=

export AWS_DEFAULT_REGION=eu-central-1
export AWS_S3_PKGS_BUCKET=code.monax.io/pkgs
export AWS_S3_PKGS_URL=pkgs.monax.io
export AWS_S3_RPM_URL=${AWS_S3_PKGS_URL}/yum
export KEY_NAME="Monax (PACKAGES SIGNING KEY) <ops@monax.io>"
export KEY_PASSWORD="one1two!three"

pre_check() {
  read -p "Have you tagged the release and filled out the changelog yet? (y/n) " -n 1 -r
  echo
  if [[ ! ${REPLY} =~ ^[Yy]$ ]]
  then
    echo "OK. Not doing anything. Rerun me after you've done that"
    exit 1
  fi
  echo "OK. Moving on then"
  echo ""
  echo ""
  # todo remove this
  if ! echo ${LATEST_TAG}|grep ${MONAX_VERSION}
  then
    echo "Something isn't right. The last tagged version does not match the version to be released"
    echo "Last tagged: ${LATEST_TAG}"
    echo "This version: ${MONAX_VERSION}"
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
}

token_check() {
  if ! type "github-release" 2>/dev/null
  then
    echo "You have to install github-release tool first"
    echo "Try 'go get -u github.com/aktau/github-release'"
    exit 1
  fi

  if [ -z "${GITHUB_TOKEN}" ]
  then
    echo "You have to have the GITHUB_TOKEN variable set to publish releases"
    exit 1
  fi
}

cross_compile() {
  if ! type "xgo" 2>/dev/null
  then
    echo "You have to install xgo tool first"
    echo "Try 'go get -u github.com/karalabe/xgo'"
    exit 1
  fi

  pushd ${REPO}/cmd/monax
  echo "Starting cross compile"

  LDFLAGS="-X github.com/monax/monax/version.COMMIT=`git rev-parse --short HEAD 2>/dev/null`"

  xgo -go 1.7 -branch ${MONAX_BRANCH} --targets=linux/amd64,linux/386,darwin/amd64,darwin/386 -dest ${BUILD_DIR}/ -out monax-${MONAX_VERSION} --pkg cmd/monax github.com/monax/monax
  # todo add build number
  aws s3 cp ${REPO}/CHANGELOG.md s3://${AWS_S3_PKGS_BUCKET}/dl/CHANGELOG --acl public-read
  echo "Cross compile completed"
  echo ""
  echo ""
  popd
}

release_binaries() {
  echo "Uploading binaries & Informing Github"
  DESCRIPTION="$(git show v${LATEST_TAG})"
  desc=$(echo -e "\n\n### To Download a Binary:\n\n")
  desc+=$(echo -e "\n* apt-get\n\n\`\`\`bash\nsudo add-apt-repository https://${AWS_S3_PKGS_URL}/apt && \\ \n  curl -L https://${AWS_S3_PKGS_URL}/apt/APT-GPG-KEY | sudo apt-key add - && \\ \n  sudo apt-get update && sudo apt-get install monax\n\`\`\`")
  desc+=$(echo -e "\n* yum\n\n\`\`\`bash\nsudo curl -L https://${AWS_S3_PKGS_URL}/yum/monax.repo >/etc/yum.repos.d/monax.repo && \\ \n  sudo yum update && sudo yum install monax\n\`\`\`")

  pushd ${BUILD_DIR}
  for file in *
  do
    echo "Uploading: ${file}"
    aws s3 cp ${file} s3://${AWS_S3_PKGS_BUCKET}/dl/${file} --acl public-read
    desc+=$(echo -e "\n* ${file}\n\n\`\`\`bash\nsudo curl -L https://${AWS_S3_PKGS_URL}/dl/${file} >/usr/local/bin/monax\nsudo chmod +x /usr/local/bin/monax\n\`\`\`")
  done
  popd
  DESCRIPTION+=${desc}
  echo "Uploading completed"
  echo ""
  echo ""

  if [[ "$1" == "pre" ]]
  then
    github-release release \
      --user monax \
      --repo monax \
      --tag v${LATEST_TAG} \
      --name "Release of Version: ${LATEST_TAG}" \
      --description "${DESCRIPTION}" \
      --pre-release
  else
    github-release release \
      --user monax \
      --repo monax \
      --tag v${LATEST_TAG} \
      --name "Release of Version: ${LATEST_TAG}" \
      --description "${DESCRIPTION}"
    if [ "$?" -ne 0 ]
    then
      github-release edit \
        --user monax \
        --repo monax \
        --tag v${LATEST_TAG} \
        --name "Release of Version: ${LATEST_TAG}" \
        --description "${DESCRIPTION}"
    fi
  fi
  echo "Finished sending release info to Github"
  echo ""
  echo ""
}

release_deb() {
  echo "Releasing Debian packages"
  shift
  mkdir -p ${BUILD_DIR}

  if [ ! -z "$@" ]
  then
    MONAX_RELEASE="$@"
  fi

  # reprepro(1) doesn't allow '-' in version numbers (that is '-rc1', etc).
  # Debian versions are not SemVer compatible.
  MONAX_DEB_VERSION=${MONAX_VERSION//-/}

  docker build -f ${REPO}/misc/release/Dockerfile-deb -t builddeb ${REPO}/misc/release \
  && docker run \
    -t \
    --rm \
    --name builddeb \
    -e MONAX_BRANCH=${MONAX_BRANCH} \
    -e MONAX_VERSION=${MONAX_DEB_VERSION} \
    -e MONAX_RELEASE=${MONAX_RELEASE} \
    -e AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY} \
    -e AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} \
    -e AWS_DEFAULT_REGION=${AWS_DEFAULT_REGION} \
    -e AWS_S3_PKGS_BUCKET=${AWS_S3_PKGS_BUCKET} \
    -e AWS_S3_PKGS_URL=${AWS_S3_PKGS_URL} \
    -e KEY_NAME="${KEY_NAME}" \
    -e KEY_PASSWORD="${KEY_PASSWORD}" \
    -v /var/run/docker.sock:/var/run/docker.sock \
    builddeb "$@"
  echo "Finished releasing Debian packages"
}

release_rpm() {
  echo "Releasing RPM packages"
  shift
  mkdir -p ${BUILD_DIR}

  if [ ! -z "$@" ]
  then
    MONAX_RELEASE="$@"
  fi

  # rpmbuild(1) doesn't allow '-' in version numbers (that is '-rc1', etc).
  # RPM versions are not SemVer compatible.
  MONAX_RPM_VERSION=${MONAX_VERSION//-/_}

  docker build -f ${REPO}/misc/release/Dockerfile-rpm -t buildrpm ${REPO}/misc/release \
  && docker run \
    -t \
    --rm \
    --name buildrpm \
    -e MONAX_BRANCH=${MONAX_BRANCH} \
    -e MONAX_VERSION=${MONAX_RPM_VERSION} \
    -e MONAX_RELEASE=${MONAX_RELEASE} \
    -e AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY} \
    -e AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} \
    -e AWS_DEFAULT_REGION=${AWS_DEFAULT_REGION} \
    -e AWS_S3_PKGS_BUCKET=${AWS_S3_PKGS_BUCKET} \
    -e AWS_S3_PKGS_URL=${AWS_S3_PKGS_URL} \
    -e KEY_NAME="${KEY_NAME}" \
    -e KEY_PASSWORD="${KEY_PASSWORD}" \
    -v /var/run/docker.sock:/var/run/docker.sock \
    buildrpm "$@"
  echo "Finished releasing RPM packages"
}

cleanup() {
  sudo rm -rf ${BUILD_DIR}
}

usage() {
  echo "Usage: release.sh [pre|build|pkgs|rpm|deb|help]"
  echo "Release Monax CLI to Github. Publish Linux packages to Amazon S3"
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
    token_check "$@"
    cross_compile "$@"
    release_deb "$@"
    release_rpm "$@"
    release_binaries "$@"
    cleanup "$@"
  esac
  return $?
}

main "$@"
