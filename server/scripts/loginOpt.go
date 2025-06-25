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
		LogError(err)

		c.Redirect(http.StatusFound, "/")
		return
	}
	session.Values["loggedin"] = false
	session.Options.MaxAge = -1 // Remove cookie
	session.Save(c.Request, c.Writer)
	c.Redirect(http.StatusFound, "/")
}
