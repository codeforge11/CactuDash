package scripts

import (
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

func CheckReq(c *gin.Context) bool {
	distro, supportStatus := RetrieveDistroInfo()

	output, err := exec.Command("docker", "--version").Output()

	if err != nil {
		LogError(err)
	}

	if strings.Contains(string(output), "not found") {
		log.Println("Docker not detected", err)

		if supportStatus {

			switch distro {
			case "arch":
				log.Println("Installing docker for " + distro)
				LogMessage("Installing docker for " + distro)
				exec.Command("pacman", "-Sy", "--noconfirm", "docker", "docker-compose").Run()
				exec.Command("systemctl", "start", "docker").Run()
				exec.Command("systemctl", "enable", "docker").Run()
				return CheckReq(c)

			case "debian", "ubuntu":
				log.Println("Installing docker for " + distro)
				LogMessage("Installing docker for " + distro)
				exec.Command("apt-get", "update").Run()
				exec.Command("apt-get", "install", "-y", "docker.io", "docker-compose").Run()
				exec.Command("systemctl", "start", "docker").Run()
				exec.Command("systemctl", "enable", "docker").Run()
				return CheckReq(c)

			case "fedora":
				log.Println("Installing docker for " + distro)
				LogMessage("Installing docker for " + distro)
				exec.Command("dnf", "install", "-y", "docker", "docker-compose").Run()
				exec.Command("systemctl", "start", "docker").Run()
				exec.Command("systemctl", "enable", "docker").Run()
				return CheckReq(c)

			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "unsupported distribution"})
				LogMessage("Unsupported distribution")
				return false
			}
		}

		c.JSON(http.StatusOK, gin.H{"status": "update script executed"})
		LogMessage("Update script executed")

	} else {
		return true
	}
	return false
}
