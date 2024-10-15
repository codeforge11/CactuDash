update_ubuntu_debian_raspbian() {
    sudo apt update && sudo apt upgrade -y
}

update_fedora_RedHat() {
    sudo dnf update && sudo dnf upgrade --refresh -y
}

if [ -f /etc/os-release ]; then
    . /etc/os-release
    case "$ID" in
        ubuntu|debian|raspbian)
            update_ubuntu_debian_raspbian
            ;;
        fedora|)
            update_fedora_RedHat
            ;;
        *)
            echo "Unsupported distribution $ID"
            ;;
    esac
else
    echo "ERROR"
fi
