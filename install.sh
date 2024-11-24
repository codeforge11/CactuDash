#!/bin/bash

# Function to detect processor architecture
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

# Function to detect OS
detect_os() {
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        case "${ID,,}" in
            raspbian) echo "raspbian" ;;
            ubuntu) echo "ubuntu" ;;
            debian) echo "debian" ;;
            fedora) echo "fedora" ;;
            *) echo "${ID,,}" ;;
        esac
    else
        echo "Unable to detect the operating system."
        exit 1
    fi
}

# Function to install Docker on Fedora
install_docker_fedora() {
    echo "Installing Docker on Fedora..."
    sudo dnf update -y
    sudo dnf install -y dnf-plugins-core
    sudo dnf config-manager --add-repo https://download.docker.com/linux/fedora/docker-ce.repo
    sudo dnf install -y docker-ce docker-ce-cli containerd.io
    sudo systemctl enable --now docker
}

# Function to install Docker on Red Hat
install_docker_redhat() {
    echo "Installing Docker on Red Hat..."
    sudo yum update -y
    sudo yum install -y yum-utils
    sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
    sudo yum install -y docker-ce docker-ce-cli containerd.io
    sudo systemctl enable --now docker
}

# Function to install Docker on Ubuntu/Debian/Raspbian
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

# Function to add user to Docker group
add_user_to_docker_group() {
    echo "Adding user $USER to the docker group..."
    sudo usermod -aG docker $USER
}

# Function to install and configure MariaDB
install_mariadb() {
    echo "Installing MariaDB on Docker..."

    root_password="CactuDash"

    sudo docker run -d \
        --name CactuDash_server \
        --restart unless-stopped \
        -e MYSQL_ROOT_PASSWORD="$root_password" \
        -p 3031:3306 mariadb:latest

    if [ $? -ne 0 ]; then
        echo "Failed to start MariaDB container. Check Docker logs."
        exit 1
    fi

    echo "Waiting for MariaDB container to initialize..."

    max_attempts=20
    attempt=1
    until sudo docker exec CactuDash_server mariadb -uroot -p"$root_password" -e "SELECT 1" &>/dev/null; do
        if [ $attempt -gt $max_attempts ]; then
            echo "MariaDB failed to initialize within the expected time. Exiting."
            sudo docker logs CactuDash_server
            exit 1
        fi
        echo "Attempt $attempt: Waiting for MariaDB to be ready..."
        attempt=$((attempt + 1))
        sleep 7
    done

    echo "MariaDB is ready. Configuring the database..."

    sudo docker exec CactuDash_server mariadb -uroot -p"$root_password" -e \
        "ALTER USER 'root'@'%' IDENTIFIED BY '$root_password';"

    sudo docker exec CactuDash_server mariadb -uroot -p"$root_password" -e \
        "CREATE DATABASE IF NOT EXISTS CactuDB;"

    sudo docker exec CactuDash_server mariadb -uroot -p"$root_password" -e \
        "USE CactuDB; 
        CREATE TABLE IF NOT EXISTS userlogin (
            id SERIAL PRIMARY KEY, 
            username TEXT NOT NULL, 
            password CHAR(125) NOT NULL
        );"

    sudo docker exec CactuDash_server mariadb -uroot -p"$root_password" -e \
        "USE CactuDB; 
        INSERT INTO userlogin (username, password) VALUES 
        ('admin', '$2a$10$eUY8TH.NXKdR2cWYyYLFZu1IyiijSKaDTEXr6HELPod01sjz3EJU.');"

    echo "MariaDB has been successfully configured."
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
        *)
            echo "No specific cleanup required for this operating system."
            ;;
    esac

    echo "Cleanup complete!"
}

main() {
    ARCH=$(detect_arch)
    OS=$(detect_os)

    echo "                                                             "
    echo "   _____              _           _____               _      "
    echo "  / ____|            | |         |  __ \             | |     "
    echo " | |      __ _   ___ | |_  _   _ | |  | |  __ _  ___ | |__   "
    echo " | |     / _` | / __|| __|| | | || |  | | / _` |/ __|| '_ \  "
    echo " | |____| (_| || (__ | |_ | |_| || |__| || (_| |\__ \| | | | "
    echo "  \_____|\__,_| \___| \__| \__,_||_____/  \__,_||___/|_| |_| "
    echo " Created by @codeforge11"
    echo " "
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

    # Verify root password
    echo "$root_password" | sudo -S echo "Root password verified." || { echo "Invalid root password. Installation aborted."; exit 1; }

    case "$OS" in
        fedora) install_docker_fedora ;;
        redhat) install_docker_redhat ;;
        ubuntu|debian|raspbian) install_docker_ubuntu_debian_raspbian ;;
        *) echo "Unsupported operating system: $OS"; exit 1 ;;
    esac

    add_user_to_docker_group
    install_mariadb
    clear_after_installation
    echo "Installation complete. Please reboot or log in again for the changes to take effect."
}

main
