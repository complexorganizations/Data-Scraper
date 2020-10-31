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
  fi
}

# Check Operating System
dist-check

# Pre-Checks
function installing-system-requirements() {
  # shellcheck disable=SC2233,SC2050
  if ([ "$DISTRO" == "ubuntu" ] || [ "$DISTRO" == "debian" ] || [ "DISTRO" == "raspbian" ]); then
    apt-get update && apt-get install git golang-go -y
  fi
  # shellcheck disable=SC2233,SC2050
  if ([ "$DISTRO" == "fedora" ] || [ "$DISTRO" == "centos" ] || [ "DISTRO" == "rhel" ]); then
    yum update -y && apt-get install git -y
  fi
  if [ "$DISTRO" == "arch" ]; then
    pacman -Syu --noconfirm git
  fi
}

# Run the function and check for requirements
installing-system-requirements

if [ ! -f "/etc/data-scraper/settings.json" ]; then

function install-the-app() {
  if ([ "$DISTRO" == "ubuntu" ] || [ "$DISTRO" == "debian" ] || [ "$DISTRO" == "arch" ] || [ "$DISTRO" == "raspbian" ] || [ "$DISTRO" == "centos" ] || [ "$DISTRO" == "fedora" ] || [ "$DISTRO" == "rhel" ]); then
    git clone https://github.com/complexorganizations/data-scraper.git /etc/
    go build /etc/data-scraper/
    chmod +x /etc/data-scraper/data-scraper
  fi
}

# run the function
install-the-app

# configure service here
function config-service() {
  if ([ "$DISTRO" == "ubuntu" ] || [ "$DISTRO" == "debian" ] || [ "$DISTRO" == "arch" ] || [ "$DISTRO" == "raspbian" ] || [ "$DISTRO" == "centos" ] || [ "$DISTRO" == "fedora" ] || [ "$DISTRO" == "rhel" ]); then
    echo "[Unit]
Description=data-scraper
After=network.target

[Service]
Type=simple
Restart=always
WorkingDirectory=/etc/data-scraper/
ExecStart=/etc/data-scraper/data-scraper

[Install]
WantedBy=multi-user.target" >> /lib/systemd/system/data-scraper.service
  # enable the file and reload config
  chmod 755 /lib/systemd/system/data-scraper.service
  systemctl daemon-reload
  fi
  if pgrep systemd-journal; then
    systemctl enable data-scraper
    systemctl start data-scraper
  else
    service data-scraper enable
    service data-scraper start
  fi
}

# run the function
config-service

else

# take user input
function take-user-input() {
    echo "What do you want to do?"
    echo "   1) Update"
    echo "   2) Uninstall"
    until [[ "$USER_OPTIONS" =~ ^[1-2]$ ]]; do
      read -rp "Select an Option [1-2]: " -e -i 1 USER_OPTIONS
    done
    case $USER_OPTIONS in
    1)
      git pull /etc/data-scraper
      ;;
    2)
      rm -rf /etc/data-scraper
      rm -f /lib/systemd/system/data-scraper.service
      ;;
    esac
}

# run the function
take-user-input

fi
