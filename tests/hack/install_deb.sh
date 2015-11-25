#!/bin/bash

unknown_os ()
{
  echo "Unfortunately, your operating system distribution and version are not supported by this script."
  exit 1
}

os=
dist=

# some systems dont have lsb-release yet have the lsb_release binary and
# vice-versa
if [ -e /etc/lsb-release ]; then
  . /etc/lsb-release
  os=${DISTRIB_ID}
  dist=${DISTRIB_CODENAME}

  if [ -z "$dist" ]; then
    dist=${DISTRIB_RELEASE}
  fi

elif [ `which lsb_release 2>/dev/null` ]; then
  dist=`lsb_release -c | cut -f2`
  os=`lsb_release -i | cut -f2 | awk '{ print tolower($1) }'`

elif [ -e /etc/debian_version ]; then
  # some Debians have jessie/sid in their /etc/debian_version
  # while others have '6.0.7'
  os=`cat /etc/issue | head -1 | awk '{ print tolower($1) }'`
  if grep -q '/' /etc/debian_version; then
    dist=`cut --delimiter='/' -f1 /etc/debian_version`
  else
    dist=`cut --delimiter='.' -f1 /etc/debian_version`
  fi

else
  unknown_os
fi

if [ -z "$dist" ]; then
  unknown_os
fi

os=`echo $os | awk '{ print tolower($1) }'`
echo "Detected operating system as $os/$dist."

echo -n "Installing apt-transport-https... "
apt-get install -y apt-transport-https &> /dev/null
echo "done."

apt_source_path="/etc/apt/sources.list.d/eris.list"
apt_url="https://apt.eris.industries"

echo -n "Importing Eris Industries' gpg key... "
apt-key adv --keyserver hkp://pool.sks-keyservers.net --recv-keys DDA1D0AB
echo "done."

echo -n "Setting apt-get sources... "
mkdir -p /etc/apt/sources.list.d
echo deb https://apt.eris.industries ${dist} main > $apt_source_path
echo "done."

echo -n "Running apt-get update... "
# update apt on this system
apt-get update
echo "done."

echo -n "Installing eris... "
apt-get install eris
echo "Installer complete."