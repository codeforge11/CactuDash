package scripts

import (
	"net/http"
	"os/exec"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

func Reboot(c *gin.Context) {

	store := sessions.NewCookieStore([]byte("secret-key"))

	session, err := store.Get(c.Request, "session-name")
	if err != nil {
		LogError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "session error"})
		return
	}

	// End the session
	session.Values["loggedin"] = false
	session.Save(c.Request, c.Writer)

	err = exec.Command("reboot").Run()
	if err != nil {
		LogError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reboot"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reboot": "rebooting"})
	LogMessage("Restart server...")
}
