package scripts

import (
	"net/http"
	"os/exec"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

func Power(c *gin.Context) {

	store := sessions.NewCookieStore([]byte("secret-key"))

	session, err := store.Get(c.Request, "session-name")
	if err != nil {
		LogError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "session error"})
		return
	}

	var powerOption struct {
		Option bool `json:"option"`
	}

	if err := c.BindJSON(&powerOption); err != nil {
		LogError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	switch powerOption.Option {
	//false for reboot
	case false:
		{

			// End the session
			session.Values["loggedin"] = false
			session.Save(c.Request, c.Writer)

			c.JSON(http.StatusOK, gin.H{"reboot": "rebooting"})
			LogMessage("Restart server...")

			err = exec.Command("reboot").Run()
			if err != nil {
				LogError(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reboot"})
				return
			}

		}
	//true for shutdown
	case true:
		{

			// End the session
			session.Values["loggedin"] = false
			session.Save(c.Request, c.Writer)

			c.JSON(http.StatusOK, gin.H{"shutdown": "Shutting down"})
			LogMessage("Shutting down server...")

			err = exec.Command("shutdown", "-h", "now").Run()
			if err != nil {
				LogError(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to shutdown"})
				return
			}
		}
	default:
		{
			LogError(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Incorrect power option"})
			return
		}
	}

}
