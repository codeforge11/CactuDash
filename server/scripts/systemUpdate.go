package scripts

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

func Update(c *gin.Context) {
	var detectedID string

	file, err := os.Open("/etc/os-release")

	if err != nil {
		log.Fatal("Error opening /etc/os-release:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		LogError(err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID=") {
			detectedID = strings.TrimPrefix(line, "ID=")
			detectedID = strings.Trim(detectedID, `"`)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("Error reading file:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		LogError(err)
		return

	}

	switch detectedID {
	case "arch":
		cmd := exec.Command("/bin/sh", "-c", `sudo pacman -Syu -y`)

		err = cmd.Run()
		if err != nil {
			log.Fatal("Error running update command:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to run update script"})
			LogError(err)
			return
		}

	case "debian", "ubuntu":
		cmd := exec.Command("/bin/sh", "-c", `sudo apt update -y && sudo apt upgrade -y`)

		err = cmd.Run()
		if err != nil {
			log.Fatal("Error running update command:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to run update script"})
			LogError(err)
			return
		}

	case "fedora":
		cmd := exec.Command("/bin/sh", "-c", `sudo dnf update -y`)

		err = cmd.Run()
		if err != nil {
			log.Fatal("Error running update command:", err)
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
