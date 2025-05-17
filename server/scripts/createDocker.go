package scripts

import (
	"net/http"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
)

func CreateDocker(c *gin.Context) {

	var Docker struct {
		Type bool   `json:"type"`
		Code string `json:"code"`
		Name string `json:"name"`
	}

	if err := c.BindJSON(&Docker); err != nil {
		LogError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	switch Docker.Type {

	//True for Docker
	case true:
		{
			LogMessage(Docker.Code)

			dockerfilePath := "Dockerfile.temp"
			err := os.WriteFile(dockerfilePath, []byte(Docker.Code), 0644)
			if err != nil {
				LogError(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Dockerfile"})
				return
			}

			cmd := exec.Command("docker", "build", "-f", dockerfilePath, "-t", Docker.Name, ".")
			output, err := cmd.CombinedOutput()
			if err != nil {
				LogError(err)
				os.Remove(dockerfilePath)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Docker build failed", "details": string(output)})
				return
			}

			if err := os.Remove(dockerfilePath); err != nil {
				LogError(err)
			}

			c.JSON(http.StatusOK, gin.H{"message": "Docker image built successfully", "output": string(output)})
		}

	//False for Docker Compose
	case false:
		{
			LogMessage("false")
		}
	default:
		LogError(nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Incorrect docker option"})
		return
	}
}
