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
    dnf update -y
    dnf install -y dnf-plugins-core
    dnf config-manager --add-repo https://download.docker.com/linux/fedora/docker-ce.repo
    dnf install -y docker-ce docker-ce-cli containerd.io
    systemctl enable --now docker
}

install_docker_redhat() {
    echo "Installing Docker on Red Hat..."
    yum update -y
    yum install -y yum-utils
    yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
    yum install -y docker-ce docker-ce-cli containerd.io
    systemctl enable --now docker
}

install_docker_ubuntu_debian_raspbian() {
    echo "Installing Docker on Ubuntu/Debian/Raspbian..."
    apt update -y
    apt install -y apt-transport-https ca-certificates curl software-properties-common
    curl -fsSL https://download.docker.com/linux/$(detect_os)/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/$(detect_os) $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
    apt update -y
    apt install -y docker-ce docker-ce-cli containerd.io
    systemctl enable --now docker
}

install_docker_arch(){
    echo "Installing Docker on Arch Linux..."
    pacman -Syu --noconfirm
    pacman -S --noconfirm docker
    systemctl enable --now docker
}

add_user_to_docker_group() {
    local user
    if [ -n "$SUDO_USER" ]; then
        user="$SUDO_USER"
    else
        user=$(logname 2>/dev/null || echo "$USER")
    fi
    echo "Adding user $user to the docker group..."
    usermod -aG docker "$user"
}

clear_after_installation() {
    echo "Cleaning up installation files..."

    case "$(detect_os)" in
        fedora|redhat)
            dnf clean all
            ;;
        ubuntu|debian|raspbian)
            apt clean
            ;;
        arch)
            pacman -Scc --noconfirm
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
        systemctl stop CactuDash

        INSTALL_DIR="/opt/CactuDash"
        if [ -d "$INSTALL_DIR" ]; then
            find "$INSTALL_DIR" -mindepth 1 -maxdepth 1 ! -name "logs" -exec rm -rf {} +
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
    
    mkdir -p "$INSTALL_DIR"

    if ! command -v unzip >/dev/null 2>&1; then
        OS=$(detect_os)
        echo "unzip not found, installing..."
        case "$OS" in
            ubuntu|debian|raspbian)
                apt update -y
                apt install -y unzip
                ;;
            fedora)
                dnf install -y unzip
                ;;
            redhat)
                yum install -y unzip
                ;;
            arch)
                pacman -S --noconfirm unzip
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
    unzip -o "$TEMP_FILE" -d "$TEMP_EXTRACT_DIR" || {
        echo "Failed to extract files."
        exit 1
    }
    rm "$TEMP_FILE"

    FIRST_DIR=$(find "$TEMP_EXTRACT_DIR" -mindepth 1 -maxdepth 1 -type d | head -n 1)
    if [ -z "$FIRST_DIR" ]; then
        echo "Could not find extracted directory."
        rm -rf "$TEMP_EXTRACT_DIR"
        exit 1
    fi
    mv "$FIRST_DIR"/* "$INSTALL_DIR"/
    rm -rf "$TEMP_EXTRACT_DIR"

    chmod +x "$INSTALL_DIR/CactuDash"
    chown -R root:root "$INSTALL_DIR"

    SERVICE_FILE="/etc/systemd/system/CactuDash.service"
    cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=CactuDash
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStartPre=/bin/chmod +x $INSTALL_DIR/CactuDash
ExecStartPre=/bin/chmod 777 $INSTALL_DIR/CactuDash
ExecStart=$INSTALL_DIR/CactuDash
Restart=always
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable CactuDash
    systemctl start CactuDash

    echo "CactuDash $LATEST_VERSION installed successfully and running as a service."
}

main() {
    if [ "$(id -u)" -ne 0 ]; then
        echo "This script must be run as root. Please use sudo."
        exit 1
    fi

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

    if [ -t 0 ]; then
        read -p "Do you want to proceed with the installation? (y/n): " choice
    else
        choice="y"
        echo "Proceeding with installation automatically (non-interactive mode)."
    fi

    case "$choice" in
        y|Y ) echo "Proceeding with installation..." ;;
        n|N ) echo "Installation aborted."; exit 0 ;;
        * ) echo "Invalid choice. Installation aborted."; exit 1 ;;
    esac

    case "$OS" in
        fedora) install_docker_fedora ;;
        redhat) install_docker_redhat ;;
        ubuntu|debian|raspbian) install_docker_ubuntu_debian_raspbian ;;
        arch) install_docker_arch ;;
        *) echo "Unsupported operating system: $OS"; exit 1 ;;
    esac

    add_user_to_docker_group
    clear_after_installation
    install_latest_version

    echo "Installation complete. Please reboot."
}

main