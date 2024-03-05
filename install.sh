#!/bin/bash

# Function to install Node.js on macOS
install_node_mac() {
    echo "Installing Node.js on macOS..."
    brew install node
}

# Function to install Node.js on Debian/Ubuntu
install_node_debian() {
    echo "Installing Node.js on Debian/Ubuntu..."
    curl -sL https://deb.nodesource.com/setup_14.x | sudo -E bash -
    sudo apt-get install -y nodejs
}

# Function to check and install Puppeteer dependencies on Debian/Ubuntu
install_puppeteer_deps() {
    echo "Installing Puppeteer dependencies..."
    sudo apt-get install -y wget unzip fontconfig locales gconf-service libasound2 libatk1.0-0 libc6 libcairo2 libcups2 libdbus-1-3 libexpat1 libfontconfig1 libgcc1 libgconf-2-4 libgdk-pixbuf2.0-0 libglib2.0-0 libgtk-3-0 libnspr4 libpango-1.0-0 libpangocairo-1.0-0 libstdc++6 libx11-6 libx11-xcb1 libxcb1 libxcomposite1 libxcursor1 libxdamage1 libxext6 libxfixes3 libxi6 libxrandr2 libxrender1 libxss1 libxtst6 ca-certificates fonts-liberation libappindicator1 libnss3 lsb-release xdg-utils wget
}

# Detect operating system and install Node.js
OS="`uname`"
case $OS in
  'Linux')
    OS='Linux'
    install_node_debian
    install_puppeteer_deps
    ;;
  'Darwin') 
    OS='macOS'
    install_node_mac
    ;;
  *) 
    OS='unknown'
    echo "Unsupported operating system, please install Node.js manually."
    exit 1
    ;;
esac

# Navigate to the puppeteer_scraper directory and install Node modules
cd ./pupeteer_scraper || exit
npm install

echo "Installation complete."
