package scripts

import (
	"log"
	"net/http"
	"os/exec"

	"github.com/gin-gonic/gin"
)

func Update(c *gin.Context) {

	osName, supportStatus := RetrieveDistroInfo()

	if supportStatus {
		switch osName {
		case "arch":
			cmd := exec.Command("/bin/sh", "-c", `sudo pacman -Syu -y`)
			err := cmd.Run()
			if err != nil {
				log.Fatal("Error running update script:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to run update script"})
				LogError(err)
				return
			}

		case "debian", "ubuntu":
			cmd := exec.Command("/bin/sh", "-c", `sudo apt update -y && sudo apt upgrade -y`)

			err := cmd.Run()
			if err != nil {
				log.Fatal("Error running update script:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to run update script"})
				LogError(err)
				return
			}

		case "fedora":
			cmd := exec.Command("/bin/sh", "-c", `sudo dnf update -y`)

			err := cmd.Run()
			if err != nil {
				log.Fatal("Error running update script:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to run update script"})
				LogError(err)
				return
			}
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unsupported distribution"})
			LogMessage("Unsupported distribution")
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "update script executed"})
		LogMessage("Update script executed")
	}
}
