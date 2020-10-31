#!/bin/bash

# Require script to be run as root (or with sudo)
function super-user-check() {
  if [ "$EUID" -ne 0 ]; then
    echo "You need to run this script as super user."
    exit
  fi
}

# Check for root
super-user-check

# Detect Operating System
function dist-check() {
  # shellcheck disable=SC1090
  if [ -e /etc/os-release ]; then
    # shellcheck disable=SC1091
    source /etc/os-release
    DISTRO=$ID
    # shellcheck disable=SC2034
    DISTRO_VERSION=$VERSION_ID
  fi
}

# Check Operating System
dist-check

# Pre-Checks
function installing-system-requirements() {
  # shellcheck disable=SC2233,SC2050
  if ([ "$DISTRO" == "ubuntu" ] && [ "$DISTRO" == "debian" ] && [ "DISTRO" == "raspbian" ]); then
    apt-get update && apt-get install curl -y
  fi
  # shellcheck disable=SC2233,SC2050
  if ([ "$DISTRO" == "fedora" ] && [ "$DISTRO" == "centos" ] && [ "DISTRO" == "rhel" ]); then
    yum update -y && yum install epel-release curl -y
  fi
  if [ "$DISTRO" == "arch" ]; then
    pacman -Syu --noconfirm curl
  fi
}

# Run the function and check for requirements
installing-system-requirements
