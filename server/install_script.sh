
# Function to detect system architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64) echo "amd64" ;;
        aarch64) echo "arm64" ;;
        *) echo "Unsupported architecture: $(uname -m)"; exit 1 ;;
    esac
}

# Function to detect OS
detect_os() {
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        echo "$ID"
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

# Function to install Docker on Ubuntu/Debian
install_docker_ubuntu_debian() {
    echo "Installing Docker on Ubuntu/Debian..."
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

# Main installation function
main() {
    ARCH=$(detect_arch)
    OS=$(detect_os)
    GO_VERSION="1.23.1"
    GO_TAR="go${GO_VERSION}.linux-${ARCH}.tar.gz"
    GO_URL="https://go.dev/dl/${GO_TAR}"

    echo "Installing Go for architecture: $ARCH"
    wget -q "$GO_URL" || { echo "Error downloading Go"; exit 1; }

    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf "$GO_TAR"
    rm "$GO_TAR"

    # Add Go to PATH
    if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
        echo "export PATH=\$PATH:/usr/local/go/bin" >> ~/.bashrc
        source ~/.bashrc
    fi

    echo "Go installation complete!"

    case "$OS" in
        fedora) install_docker_fedora ;;
        ubuntu|debian) install_docker_ubuntu_debian ;;
        *) echo "Unsupported operating system: $OS"; exit 1 ;;
    esac

    add_user_to_docker_group
    echo "Installation complete. Please reboot or log in again for the changes to take effect."
}

main
