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
			err := exec.Command("/bin/sh", "-c", `pacman -Syu -y`).Run()

			if err != nil {
				log.Fatal("Error running update script:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to run update script"})
				LogError(err)
				return
			}

		case "debian", "ubuntu":
			err := exec.Command("/bin/sh", "-c", `apt update -y && apt upgrade -y`).Run()

			if err != nil {
				log.Fatal("Error running update script:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to run update script"})
				LogError(err)
				return
			}

		case "fedora":
			err := exec.Command("/bin/sh", "-c", `dnf update -y`).Run()

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
