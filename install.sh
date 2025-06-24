#!/bin/bash

detect_arch() {
    case "$(uname -m)" in
        x86_64) echo "amd64" ;;
        aarch64) echo "arm64" ;;
        armv7l) echo "armv7" ;;
        armv6l) echo "armv6" ;;
        i386|i686) echo "386" ;;
        ppc64le) echo "ppc64le" ;;
        s390x) echo "s390x" ;;
        riscv64) echo "riscv64" ;;
        *) echo "Unsupported processor architecture: $(uname -m)"; exit 1 ;;
    esac
}

detect_os() {
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        case "${ID,,}" in
            raspbian) echo "raspbian" ;;
            ubuntu) echo "ubuntu" ;;
            debian) echo "debian" ;;
            fedora) echo "fedora" ;;
            arch) echo "arch" ;;
            *)
                if [[ -n "$ID_LIKE" ]]; then
                    case "${ID_LIKE,,}" in
                        *debian*) echo "debian" ;;
                        *ubuntu*) echo "ubuntu" ;;
                        *fedora*) echo "fedora" ;;
                        *arch*) echo "arch" ;;
                        *rhel*|*centos*) echo "redhat" ;;
                        *) echo "${ID,,}" ;;
                    esac
                else
                    echo "${ID,,}"
                fi
                ;;
        esac
    else
        echo "Unable to detect the operating system."
        exit 1
    fi
}

install_docker_fedora() {
    echo "Installing Docker on Fedora..."
    sudo dnf update -y
    sudo dnf install -y dnf-plugins-core
    sudo dnf config-manager --add-repo https://download.docker.com/linux/fedora/docker-ce.repo
    sudo dnf install -y docker-ce docker-ce-cli containerd.io
    sudo systemctl enable --now docker
}

install_docker_redhat() {
    echo "Installing Docker on Red Hat..."
    sudo yum update -y
    sudo yum install -y yum-utils
    sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
    sudo yum install -y docker-ce docker-ce-cli containerd.io
    sudo systemctl enable --now docker
}

install_docker_ubuntu_debian_raspbian() {
    echo "Installing Docker on Ubuntu/Debian/Raspbian..."
    sudo apt update -y
    sudo apt install -y apt-transport-https ca-certificates curl software-properties-common
    curl -fsSL https://download.docker.com/linux/$(detect_os)/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/$(detect_os) $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
    sudo apt update -y
    sudo apt install -y docker-ce docker-ce-cli containerd.io
    sudo systemctl enable --now docker
}

install_docker_arch(){
    echo "Installing Docker on Arch Linux..."
    sudo pacman -Syu --noconfirm
    sudo pacman -S --noconfirm docker
    sudo systemctl enable --now docker
}

add_user_to_docker_group() {
    echo "Adding user $USER to the docker group..."
    sudo usermod -aG docker $USER
}

clear_after_installation() {
    echo "Cleaning up installation files..."

    case "$(detect_os)" in
        fedora|redhat)
            sudo dnf clean all
            ;;
        ubuntu|debian|raspbian)
            sudo apt clean
            ;;
        arch)
            sudo pacman -Scc
            ;;
        *)
            echo "No specific cleanup required for this operating system."
            ;;
    esac

    echo "Cleanup complete!"
}

install_latest_version() {
    echo "Installing latest version of CactuDash..."

    if systemctl is-active --quiet CactuDash; then
        echo "CactuDash service is running. Stopping service..."
        sudo systemctl stop CactuDash

        INSTALL_DIR="/opt/CactuDash"
        if [ -d "$INSTALL_DIR" ]; then
            sudo find "$INSTALL_DIR" -mindepth 1 -maxdepth 1 ! -name "logs" -exec rm -rf {} +
        fi
    fi

    LATEST_RELEASE=$(curl -s "https://api.github.com/repos/codeforge11/CactuDash/releases/latest")
    LATEST_VERSION=$(echo "$LATEST_RELEASE" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$LATEST_VERSION" ]; then
        echo "Failed to get latest version information."
        exit 1
    fi

    ARCH=$(detect_arch)
    DOWNLOAD_URL="https://github.com/codeforge11/CactuDash/releases/download/$LATEST_VERSION/CactuDash-$LATEST_VERSION-$ARCH.zip"
    INSTALL_DIR="/opt/CactuDash"
    
    sudo mkdir -p "$INSTALL_DIR"

    if ! command -v unzip >/dev/null 2>&1; then
        OS=$(detect_os)
        echo "unzip not found, installing..."
        case "$OS" in
            ubuntu|debian|raspbian)
                sudo apt update -y
                sudo apt install -y unzip
                ;;
            fedora)
                sudo dnf install -y unzip
                ;;
            redhat)
                sudo yum install -y unzip
                ;;
            arch)
                sudo pacman -S --noconfirm unzip
                ;;
            *)
                echo "Please install 'unzip' manually for your OS."
                exit 1
                ;;
        esac
    fi

    TEMP_FILE=$(mktemp)
    curl -L "$DOWNLOAD_URL" -o "$TEMP_FILE" || {
        echo "Failed to download release."
        exit 1
    }

    TEMP_EXTRACT_DIR=$(mktemp -d)
    sudo unzip -o "$TEMP_FILE" -d "$TEMP_EXTRACT_DIR" || {
        echo "Failed to extract files."
        exit 1
    }
    rm "$TEMP_FILE"

    FIRST_DIR=$(sudo find "$TEMP_EXTRACT_DIR" -mindepth 1 -maxdepth 1 -type d | head -n 1)
    if [ -z "$FIRST_DIR" ]; then
        echo "Could not find extracted directory."
        sudo rm -rf "$TEMP_EXTRACT_DIR"
        exit 1
    fi
    sudo mv "$FIRST_DIR"/* "$INSTALL_DIR"/
    sudo rm -rf "$TEMP_EXTRACT_DIR"

    sudo chmod +x "$INSTALL_DIR/CactuDash"
    sudo chown -R root:root "$INSTALL_DIR"

    SERVICE_FILE="/etc/systemd/system/CactuDash.service"
    sudo bash -c "cat > $SERVICE_FILE" <<EOF
[Unit]
Description=CactuDash
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/CactuDash
Restart=always
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

    sudo systemctl daemon-reload
    sudo systemctl enable CactuDash
    sudo systemctl start CactuDash

    echo "CactuDash $LATEST_VERSION installed successfully and running as a service."
}

main() {
    ARCH=$(detect_arch)
    OS=$(detect_os)

    echo -e "\e[31m   _____              _           _____               _      \e[0m"
    echo -e "\e[33m  / ____|            | |         |  __ \             | |     \e[0m"
    echo -e "\e[32m | |      __ _   ___ | |_  _   _ | |  | |  __ _  ___ | |__   \e[0m"
    echo -e "\e[36m | |     / _\` | / __|| __|| | | || |  | | / _\` |/ __|| '_ \  \e[0m"
    echo -e "\e[34m | |____| (_| || (__ | |_ | |_| || |__| || (_| |\__ \| | | | \e[0m"
    echo -e "\e[35m  \_____|\__,_| \___| \__| \__,_||_____/  \__,_||___/|_| |_| \e[0m"
    echo " Created by @codeforge11"
    echo "--------------------------------------------------------------------------"
    echo "Detected system: $OS"
    echo "Detected architecture: $ARCH"
    read -p "Do you want to proceed with the installation? (y/n): " choice
    case "$choice" in
        y|Y ) echo "Proceeding with installation..." ;;
        n|N ) echo "Installation aborted."; exit 0 ;;
        * ) echo "Invalid choice. Installation aborted."; exit 1 ;;
    esac

    read -sp "Enter root password: " root_password
    echo

    echo "$root_password" | sudo -S echo "Root password verified." || { echo "Invalid root password. Installation aborted."; exit 1; }

    case "$OS" in
        fedora) install_docker_fedora ;;
        redhat) install_docker_redhat ;;
        ubuntu|debian|raspbian) install_docker_ubuntu_debian_raspbian ;;
        arch) install_docker_arch ;;
        *) echo "Unsupported operating system: $OS"; exit 1 ;;
    esac

    add_user_to_docker_group
    install_mariadb
    clear_after_installation
    install_latest_version

    echo "Installation complete. Please reboot."
}

main