package scripts

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

func randomKey(length int) string { //fully random key generator
	bytes := make([]byte, length)

	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(bytes)[:length]
}

var Store = sessions.NewCookieStore([]byte(randomKey(128)))

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
		LogError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "session error"})
		return nil, err
	}
	return session, nil
}
