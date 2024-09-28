update_debian() {
    sudo apt update && sudo apt upgrade -y
}

update_fedora() {
    sudo dnf update && sudo dnf upgrade --refresh -y
}

if [ -f /etc/os-release ]; then
    . /etc/os-release
    case "$ID" in
        ubuntu|debian)
            update_debian
            ;;
        fedora)
            update_fedora
            ;;
        *)
            echo "Unsupported distribution $ID"
            ;;
    esac
else
    echo "ERROR"
fi
