package scripts

import (
	"net/http"

	"os/exec"

	"github.com/gin-gonic/gin"
)

func Update(c *gin.Context) {
	err := exec.Command("/bin/sh", "-c", `
	update_ubuntu_debian_raspbian() {
		sudo apt update && sudo apt upgrade -y
	}

	update_fedora_RedHat() {
		sudo dnf update && sudo dnf upgrade --refresh -y
	}

	update_arch(){
		sudo pacman -Syu -y
	}

	if [ -f /etc/os-release ]; then
		. /etc/os-release
		case "$ID" in
			ubuntu|debian|raspbian)
				update_ubuntu_debian_raspbian
				;;
			fedora|ultramarine|rhel)
				update_fedora_RedHat
				;;
			arch|manjaro)
				update_arch
				;;
			*)
				echo "Unsupported distribution $ID"
				;;
		esac
	else
		echo "ERROR"
	fi
	`).Run()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to run update script"})
		LogError(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "update script executed"})
	LogMessage("Update script executed")

}
