package scripts

import (
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

func CreateDocker(c *gin.Context) {

	var Docker struct {
		Type bool   `json:"type"`
		Code string `json:"code"`
		Name string `json:"name"`
	}

	if err := c.BindJSON(&Docker); err != nil {
		log.Println(err)

		LogError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	switch Docker.Type {

	//True for Docker
	case true:
		{
			session, err := Store.Get(c.Request, "session-name")
			if err != nil {
				log.Println(err)

				LogError(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Session error"})
				return
			}

			if loggedIn, ok := session.Values["loggedin"].(bool); ok && loggedIn {
				if strings.Contains(Docker.Code, "docker run") || strings.Contains(Docker.Code, "sudo docker run") {

					cmd := exec.Command("bash", "-c", Docker.Code)
					output, err := cmd.CombinedOutput()

					if err != nil {
						LogError(err)
						c.JSON(http.StatusInternalServerError, gin.H{
							"error":   "Failed to execute docker run command",
							"details": err.Error(),
							"output":  string(output),
						})
						LogError(errors.New("Docker run command failed: " + err.Error() + " | " + string(output)))
						return
					}

					LogMessage("Docker run command executed successfully")

				} else {
					LogMessage("error: incorrect docker run command")
					c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect docker run command"})
					return
				}

				c.JSON(http.StatusOK, gin.H{"message": "Docker image built successfully"})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
				return
			}
		}

	//False for Docker Compose
	case false:
		{
			dir := "workDirectory/" + Docker.Name

			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Println(err)

				LogError(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
				return
			}

			composeFilePath := dir + "/compose.yaml"
			file, err := os.Create(composeFilePath)
			if err != nil {
				log.Println(err)

				LogMessage("error: Failed to create compose.yaml")
				c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create compose.yaml"})
				return
			}
			defer file.Close()

			_, err = file.WriteString(Docker.Code)
			if err != nil {
				log.Println(err)

				LogError(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write to compose.yaml"})
				return
			}

			cmd := exec.Command("docker-compose", "-f", composeFilePath, "up", "-d")
			cmd.Dir = dir
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.Println(err)

				LogError(err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to execute docker-compose up",
					"details": err.Error(),
					"output":  string(output),
				})
				return
			}

			LogMessage("Docker Compose started successfully")
			c.JSON(http.StatusOK, gin.H{"message": "Docker Compose started successfully"})

			err = exec.Command("rm", "-rf", dir).Run()
			if err != nil {
				log.Println("Failed to delete compose working folder: " + dir)
				LogMessage("Failed to delete compose working folder: " + dir)
				LogError(err)
			}

		}

	default:
		// err := exec.Command("rm", "-rf", "./workDirectory").Run()
		// if err != nil {
		// 	log.Println("Failed to delete compose working folder: " + "./workDirectory")
		// 	LogMessage("Failed to delete compose working folder: " + "./workDirectory")
		// 	LogError(err)
		// }

		LogMessage("Incorrect docker option.")
		log.Println("Incorrect docker option.")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Incorrect docker option"})
		return
	}
}
