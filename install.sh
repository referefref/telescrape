#!/bin/bash

# Function to install Node.js using NVM
install_node_with_nvm() {
    echo "Installing NVM (Node Version Manager)..."
    curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/master/install.sh | bash
    export NVM_DIR="$([ -z "${XDG_CONFIG_HOME-}" ] && printf %s "${HOME}/.nvm" || printf %s "${XDG_CONFIG_HOME}/nvm")"
    [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh" # This loads nvm
    nvm install 20
    nvm use 20
    nvm alias default 20
}

# Check if Node.js is installed and its version
check_node_version() {
    if command -v node > /dev/null; then
        NODE_VERSION=$(node -v)
        echo "Node.js is already installed with version $NODE_VERSION."
        if [[ "$NODE_VERSION" != v20* ]]; then
            echo "But we need version 20. Attempting to install using NVM..."
            install_node_with_nvm
        else
            echo "Required Node.js version 20 is already installed."
        fi
    else
        echo "Node.js is not installed. Installing version 20 using NVM..."
        install_node_with_nvm
    fi
}

# Function to install Puppeteer dependencies on Debian/Ubuntu
install_puppeteer_deps() {
    echo "Installing Puppeteer dependencies..."
    sudo apt-get update
    sudo apt-get install -y wget unzip fontconfig locales gconf-service libasound2 libatk1.0-0 libc6 libcairo2 libcups2 libdbus-1-3 libexpat1 libfontconfig1 libgcc1 libgconf-2-4 libgdk-pixbuf2.0-0 libglib2.0-0 libgtk-3-0 libnspr4 libpango-1.0-0 libpangocairo-1.0-0 libstdc++6 libx11-6 libx11-xcb1 libxcb1 libxcomposite1 libxcursor1 libxdamage1 libxext6 libxfixes3 libxi6 libxrandr2 libxrender1 libxss1 libxtst6 ca-certificates fonts-liberation libappindicator1 libnss3 lsb-release xdg-utils wget
}

# Function to install GoLang on macOS using Homebrew
install_golang_mac() {
    echo "Installing GoLang on macOS..."
    brew install go
}

# Function to install GoLang on Debian/Ubuntu
install_golang_debian() {
    echo "Installing GoLang on Debian/Ubuntu..."
    wget https://dl.google.com/go/go1.22.0.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
    echo "export PATH=$PATH:/usr/local/go/bin" >> $HOME/.profile
    source $HOME/.profile
}

# Function to check for GoLang installation
check_golang() {
    if ! command -v go &> /dev/null; then
        echo "GoLang could not be found"
        if [[ "$OS" == "Linux" ]]; then
            install_golang_debian
        elif [[ "$OS" == "macOS" ]]; then
            install_golang_mac
        fi
    else
        echo "GoLang is already installed."
    fi
}

# Detect operating system
OS="`uname`"
case $OS in
  'Linux')
    OS='Linux'
    install_puppeteer_deps
    check_node_version
    check_golang
    ;;
  'Darwin')
    OS='macOS'
    check_node_version
    check_golang
    ;;
  *)
    OS='unknown'
    echo "Unsupported operating system, please install Node.js and GoLang manually."
    exit 1
    ;;
esac 

#Create temp folder for attachments
mkdir -p ./temp

# Navigate to the puppeteer_scraper directory and install Node modules
cd ./pupeteer_scraper || exit
npm install

# build go application
cd .. || exit
go mod tidy
go build
echo "Installation complete."
