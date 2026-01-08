package scripts

import (
	"log"
	"net/http"
	"os/exec"

	"github.com/codeforge11/betterLogs"
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
				betterLogs.LogError(err)
				return
			}

		case "debian", "ubuntu":
			err := exec.Command("/bin/sh", "-c", `apt update -y && apt upgrade -y`).Run()

			if err != nil {
				log.Fatal("Error running update script:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to run update script"})
				betterLogs.LogError(err)
				return
			}

		case "fedora":
			err := exec.Command("/bin/sh", "-c", `dnf update -y`).Run()

			if err != nil {
				log.Fatal("Error running update script:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to run update script"})
				betterLogs.LogError(err)
				return
			}
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unsupported distribution"})
			betterLogs.LogMessage("Unsupported distribution")
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "update script executed"})
		betterLogs.LogMessage("Update script executed")
	}
}
