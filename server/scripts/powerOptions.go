package scripts

import (
	"net/http"
	"os/exec"

	"github.com/codeforge11/betterLogs"
	"github.com/gin-gonic/gin"
)

func Power(c *gin.Context) {

	session, err := GetSession(c)
	if err != nil {
		betterLogs.LogError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "session error"})
		return
	}

	var powerOption struct {
		Option bool `json:"option"`
	}

	if err := c.BindJSON(&powerOption); err != nil {
		betterLogs.LogError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	switch powerOption.Option {
	//false for reboot
	case false:
		{

			// Move user to login page
			c.JSON(http.StatusOK, gin.H{"redirect": "/"})

			// End the session
			session.Values["loggedin"] = false
			session.Options.MaxAge = 0 // Remove cookie
			session.Save(c.Request, c.Writer)

			c.JSON(http.StatusOK, gin.H{"reboot": "rebooting"})
			betterLogs.LogMessage("Restart server...")

			err = exec.Command("reboot").Run()
			if err != nil {
				betterLogs.LogError(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reboot"})
				return
			}

		}
	//true for shutdown
	case true:
		{

			// Move user to login page
			c.JSON(http.StatusOK, gin.H{"redirect": "/"})

			// End the session
			session.Values["loggedin"] = false
			session.Options.MaxAge = 0 // Remove cookie
			session.Save(c.Request, c.Writer)

			c.JSON(http.StatusOK, gin.H{"shutdown": "Shutting down"})
			betterLogs.LogMessage("Shutting down server...")

			err = exec.Command("shutdown", "-h", "now").Run()
			if err != nil {
				betterLogs.LogError(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to shutdown"})
				return
			}
		}
	default:
		{
			betterLogs.LogError(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Incorrect power option"})
			return
		}
	}

}
