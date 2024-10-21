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
install_mariadb_server() {
    echo "Installing MariaDB..."
    sudo apt update -y
    sudo apt install -y mariadb-server
    echo "Configuring MariaDB..."
    sudo systemctl start mariadb
    sudo systemctl enable mariadb
    root_password="CactuDash"
    sudo mysql -uroot -p"$root_password" -e "ALTER USER 'root'@'localhost' IDENTIFIED BY 'CactuDash';"
    sudo mysql -uroot -p"CactuDash" -e "CREATE DATABASE CactuDB;"
    sudo mysql -uroot -p"CactuDash" -e "USE CactuDB; CREATE TABLE userlogin (id SMALLINT(3) UNSIGNED PRIMARY KEY AUTO_INCREMENT, username TEXT NOT NULL, password CHAR(60) NOT NULL);"
    sudo mysql -uroot -p"CactuDash" -e "USE CactuDB; INSERT INTO userlogin VALUES (NULL, 'admin', '\$2a\$10\$VXivP/o1tuQaALdmdECeyOAVfF830qgxcv3Nw71ATSD3RNz3qJMBa');"

    echo "Configuring MariaDB to listen on port 3031..."
    sudo sed -i "s/port\s*=\s*3306/port = 3031/" /etc/mysql/mariadb.conf.d/50-server.cnf
    sudo systemctl restart mariadb

    echo "MariaDB installation and configuration complete!"
}


# Function to build Go from source
build_go() {
    echo "Building Go from source for architecture: $ARCH"
    wget https://golang.org/dl/go1.23.2.src.tar.gz
    tar -xzf go1.23.2.src.tar.gz
    cd go
    bash make.bash
    cd ..
    sudo rm -rf go1.23.2.src.tar.gz go
}

# Function to install or update Go
install_go() {
    ARCH=$(detect_arch)
    OS=$(detect_os)

    # Check if Go is already installed
    if command -v go &> /dev/null; then
        echo "Go is already installed. Updating..."
        if [[ "$ARCH" == "armv6" ]]; then
            build_go
        else
            GO_VERSION="1.23.2"
            GO_TAR="go${GO_VERSION}.${OS}-${ARCH}.tar.gz"
            GO_URL="https://golang.org/dl/${GO_TAR}"

            echo "Downloading Go for architecture: $ARCH"
            wget -q "$GO_URL" || { echo "Error downloading Go"; exit 1; }
            sudo rm -rf /usr/local/go
            sudo tar -C /usr/local -xzf "$GO_TAR"
            rm "$GO_TAR"
        fi
    else
        echo "Installing Go for architecture: $ARCH"
        if [[ "$ARCH" == "armv6" ]]; then
            build_go
        else
            GO_VERSION="1.23.2"
            GO_TAR="go${GO_VERSION}.${OS}-${ARCH}.tar.gz"
            GO_URL="https://golang.org/dl/${GO_TAR}"

            if wget --spider "$GO_URL" 2>/dev/null; then
                wget -q "$GO_URL" || { echo "Error downloading Go"; exit 1; }
            else
                echo "Go binary for architecture $ARCH is not available."
                exit 1
            fi

            sudo rm -rf /usr/local/go
            sudo tar -C /usr/local -xzf "$GO_TAR"
            rm "$GO_TAR"
        fi
    fi

    # Add Go to PATH
    if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
        echo "export PATH=\$PATH:/usr/local/go/bin" >> ~/.bashrc
        source ~/.bashrc
    fi

    echo "Go installation/upgrade complete!"
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

    install_go

    case "$OS" in
        fedora) install_docker_fedora ;;
        redhat) install_docker_redhat ;;
        ubuntu|debian|raspbian) install_docker_ubuntu_debian_raspbian ;;
        *) echo "Unsupported operating system: $OS"; exit 1 ;;
    esac

    add_user_to_docker_group
    install_mariadb_server
    clear_after_installation
    echo "Installation complete. Please reboot or log in again for the changes to take effect."
}

main
