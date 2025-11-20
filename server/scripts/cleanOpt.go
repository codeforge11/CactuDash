package scripts

import (
	"log"
	"os/exec"

	"github.com/gin-gonic/gin"
)

// func to remove old logs files
func ClearOldLogs(c *gin.Context) {
	cmd := exec.Command("/bin/sh", "-c", `rm -rf ./logs/old_logs`)

	if err := cmd.Run(); err != nil {
		log.Println("Error running clear cash script:", err)
		LogError(err)
	}
}

// func to remove old work directories
func CleanOldWorkDirectory(c *gin.Context) {
	cmd := exec.Command("/bin/sh", "-c", `rm -rf ./workDirectory/*`)

	if err := cmd.Run(); err != nil {
		log.Println("Error running script to clean up old working directories:", err)
		LogError(err)
	}

}
