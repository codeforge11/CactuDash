package scripts

import (
	"encoding/gob"
	"log"
	"net/http"
	"time"

	betterLogs "github.com/codeforge11/betterLogs"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

// func randomKey(length int) string { //fully random key generator
// 	bytes := make([]byte, length)
// 	_, err := rand.Read(bytes)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return base64.StdEncoding.EncodeToString(bytes)[:length]
// }

var Store = sessions.NewCookieStore([]byte("super-secret-key"))

var SessionExpiration = 15 * time.Minute // session time
var ServerStartTime = time.Now()

func init() {
	gob.Register(time.Time{})
	Store.Options = &sessions.Options{
		MaxAge:   int(SessionExpiration.Seconds()),
		HttpOnly: true,
	}
}

func GetSession(c *gin.Context) (*sessions.Session, error) {
	session, err := Store.Get(c.Request, "session-name")
	if err != nil {
		log.Println(err)
		betterLogs.LogError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "session error"})
		return nil, err
	}
	return session, nil
}
