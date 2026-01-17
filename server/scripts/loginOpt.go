package scripts

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Logout(c *gin.Context) {
	session, err := Store.Get(c.Request, "session-name")
	if err != nil {
		log.Println("Error getting session:", err)
		BetterLogs.LogError(err)

		c.Redirect(http.StatusFound, "/")
		return
	}
	session.Values["loggedin"] = false
	session.Options.MaxAge = -1 // Remove cookie

	if err := session.Save(c.Request, c.Writer); err != nil {
		log.Println("Error saving session:", err)
		BetterLogs.LogError(err)

		// Move user into /
		c.Redirect(http.StatusFound, "/")
		return
	}

	// Move user into /
	c.Redirect(http.StatusFound, "/")
}
