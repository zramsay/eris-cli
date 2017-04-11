---

layout: single
title: Tutorial | Install Eris blockchain tools on IoT (ARM) devices
aliases:
   - /docs/deprecated/install-arm
---

## Introduction

<div class="note">
{{% data_sites rename_docs %}}
</div>

This tutorial uses Raspberry Pi 3 as referenced IoT device to demonstrate how to install eris blockchain tools on IoT devices.

## Prerequisite

* Raspberry Pi2 (RPi2) or RPi3.
* At least 32 GB micro-SD card to run Blockchain.
* Latest Raspbian image or NOOBS system installer (https://www.raspberrypi.org/downloads/).


## Initialize your brand new Raspberry Pi

Let's assume that you have a brand new Raspberry Pi board for this section. If you are using one that has all the basic development tools, skip these steps as you want.

1. Install Raspberry Pi through NOOBS installer or flash the Raspbian image directly to the micro-SD card.
2. Boot the installed Raspbian system (default login credential {username: pi, password: raspberry}) and reconfigure the time zone:

```bash
sudo dpkg-reconfigure tzdata
```

Reconfigure the keyboard corresponding to your country layout:

```bash
sudo dpkg-reconfigure keyboard-configuration
```

Reconfigure the language locales ( to `en_US.UTF-8` for example):

```bash
sudo dpkg-reconfigure locales
```

3. Setup the wireless network.

   If you are using RPi3 or have Wi-Fi dongle connected to RPi2, it's easy to get connected to network through the wireless interface `wlan0`.

   1. Edit the `/etc/network/interfaces` file to change the wlan0 from `manual` mode to `dhcp`

```bash
...
allow-hotplug wlan0
iface wlan0 inet dhcp
    wpa-conf /etc/wpa_supplicant/wpa_supplicant.conf
...
```

   2. Append the Wi-Fi login credential to `wpa_supplicant.conf`.

```bash
wpa_passphrase {SSID} {PASSWORD} | sudo tee -a /etc/wpa_supplicant/wpa_supplicant.conf
```

   3. Restart `wlan0` by `sudo ifdown wlan0 && sudo ifup wlan0`

4. After you get the network working, update the system:

```bash
sudo apt-get update && sudo apt-get upgrade -y --no-install-recommends
```

5. Create a new {USER} account and replace the default `pi` with {USER} in the sudoer file.

   Since the default `pi` account comes with low security configurations and over-privileged, it's more secure to create and use your own {USER} account instead of the default `pi` account.

```bash
sudo  useradd -m -s /bin/bash {USER}
sudo passwd {USER}
sudo  sed -i 's/\bpi\b/{USER}/' /etc/sudoers
reboot && login as {USRR}
sudo userdel -r pi
```

6. Install frequently used tools and applications for IoT development.

```bash
sudo apt-get install -y --no-install-recommends \
   vim screen git
```

## Install docker and swarm cluster management tools

Using `docker-machine` with swarm lets us manage the IoT nodes in a `master vs. slave` scheme, which orchestrates the IoT blockchain fleet elegantly. In the following description, we use **DM** to refer the docker machine node, **swarm master** to refer the master node in the swarm protocol and **swarm slave** to refer the slave nodes. There are one master node and multiple slave nodes. Use `-m` option to specify the node you want to run the eris blockchain.

We're going to use [Hypriot](http://blog.hypriot.com/downloads/) docker to provision the nodes. Since the hypriot docker uses Hypriot OS as default OS environment, we need to change the Raspbian release declaration file to make the provisioned OS compatible to the Hypriot OS:

```bash
sudo sed -i 's/ID=raspbian/ID=debian/g' /etc/os-release
```

### Install hypriot docker on the swarm nodes and copy ssh login credentials from DM to swarm nodes

```bash
wget https://downloads.hypriot.com/docker-hypriot_{VERSION}_armhf.deb \
  && sudo dpkg -i docker-hypriot_{VERSION}_armhf.deb \
  && rm docker-hypriot_{VERSION}_armhf.deb
```

Add current user to `docker` group `sudo usermod -a -G docker {USER}`. You need to restart/re-login to let the group modification take effect.

Then copy the ssh credential to each nodes by `ssh-copy-id USER@TARGET_NODE_IP`.

### Provision swarm master and slave machines.

1. Create swarm discovery token

```bash
export TOKEN=$(for i in $(seq 1 32); do echo -n $(echo "obase=16; $(($RANDOM % 16))" | bc); done; echo)
```

2. Provision master node

```bash
docker-machine create  --engine-storage-driver devicemapper -d generic --swarm --swarm-master \
 --swarm-image hypriot/rpi-swarm:latest --swarm-discovery token://{TOKEN} --generic-ip-address {MASTER_IP_ADDR} \
 --generic-ssh-user {USER} {MACHINE_NAME}
```

3. Provision slave node

```bash
docker-machine create  --engine-storage-driver devicemapper -d generic --swarm --swarm-image hypriot/rpi-swarm:latest \
--swarm-discovery token://{TOKEN} --generic-ip-address {MASTER_IP_ADDR} --generic-ssh-user {USER} {MACHINE_NAME}
```

4. Check swarm nodes information.

```bash
eval `docker-machine env --swarm {MASTER_MACHINE}`; docker info
```

To unset the swarm docker environment, `docker-machine env --unset`.


## Install eris Debian package

We built the Debian repo of the eris command line tool to make it to be easily installed just by `apt-get`.

```bash
curl https://eris-iot-repo.s3.amazonaws.com/eris-deb/APT-GPG-KEY | sudo apt-key add -
echo "deb https://eris-iot-repo.s3.amazonaws.com/eris-deb DIST experimental" | sudo tee /etc/apt/sources.list.d/eris.list

apt-get update
apt-get install eris
```

## Install nodejs and build dependencies (For smart contract)

```bash
curl -sSL https://deb.nodesource.com/setup_{NODEVERSION}.x | sudo -E bash - &>/dev/null
sudo apt-get install -y bc jq gcc git build-essential nodejs &>/dev/null
```

## Start blockchain, run contracts!

Learn more and play with eris blockchain, go to [Getting Started With Eris](/docs/getting-started)



## [<i class="fa fa-chevron-circle-left" aria-hidden="true"></i> All Deprecated](/docs/deprecated/)


