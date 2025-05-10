package scripts

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
)

func Update(c *gin.Context) {

	var detectedID string
	var detectedIDLike string
	var supportStatus bool
	var osName string

	file, err := os.Open("/etc/os-release") //for detect distro name
	if err != nil {
		log.Println("Error opening /etc/os-release:", err)
		LogError(err)
	} else {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "ID=") {
				detectedID = strings.TrimPrefix(line, "ID=")
				detectedID = strings.Trim(detectedID, `"`)

			}
			if strings.HasPrefix(line, "ID_LIKE=") {
				detectedIDLike = strings.TrimPrefix(line, "ID_LIKE=")
				detectedIDLike = strings.Trim(detectedIDLike, `"`)

			}
		}
		if err := scanner.Err(); err != nil {
			log.Println("Error reading file:", err)
			LogError(err)
		}
	}

	if runtime.GOOS == "linux" {
		osName = detectedID
	} else {
		osName = runtime.GOOS
	}

	if osName == detectedID {
		supportStatus = true
	} else if osName == detectedIDLike {
		supportStatus = true
	} else {
		supportStatus = false
	}

	if supportStatus {
		switch osName {
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
}
