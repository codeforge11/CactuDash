package scripts

import (
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

// Structure for containers
type Container struct {
	Id     string `json:"Id"`
	Image  string `json:"Image"`
	Status string `json:"Status"`
	Name   string `json:"Name"`
}

// Function to get Docker containers
func GetContainers(c *gin.Context) {
	out, err := exec.Command("docker", "ps", "-a", "--format", "{{.ID}};{{.Image}};{{.Ports}};{{.Status}};{{.Names}}").Output()
	if err != nil {
		log.Println("Error executing docker command:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error executing docker command"})
		LogError(err)
		return
	}

	var containers []Container
	for _, line := range strings.Split(string(out), "\n") {
		if line != "" {
			fields := strings.Split(line, ";")
			if len(fields) < 5 {
				log.Println("Unexpected format in docker output:", line)
				continue
			}
			containers = append(containers, Container{Id: fields[0], Image: fields[1], Status: fields[3], Name: fields[4]})
		}
	}

	c.JSON(http.StatusOK, containers)
}

// Function to change docker state
func ToggleContainerState(c *gin.Context) {
	id := strings.TrimPrefix(c.Param("id"), "/toggle/")
	out, err := exec.Command("docker", "inspect", "--format={{.State.Running}}", id).Output()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		LogError(err)
		return
	}

	running := strings.TrimSpace(string(out))
	if running == "true" {
		if err := exec.Command("docker", "stop", id).Run(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stop container"})
			LogError(err)
			return
		}
		LogMessage("Container stopped:" + id)
	} else {
		if err := exec.Command("docker", "start", id).Run(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start container"})
			LogError(err)
			return
		}
		LogMessage("Container started:" + id)
		// fmt.Println("Container started:", id)
	}
	c.Status(http.StatusNoContent)
}

// Function to restart docker
func RestartContainer(c *gin.Context) {
	id := strings.TrimPrefix(c.Param("id"), "/restart/")

	if err := exec.Command("docker", "restart", id).Run(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restart container"})
		LogError(err)
		return
	}
	LogMessage("Container restarted:" + id)

}

// Function to remove docker
func RemoveContainer(c *gin.Context) {
	id := strings.TrimPrefix(c.Param("id"), "/remove/")

	if err := exec.Command("docker", "rm", id).Run(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove container"})
		LogError(err)
		return
	}
	LogMessage("Container removed:" + id)

}
