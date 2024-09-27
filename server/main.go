package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"

	"golang.org/x/crypto/bcrypt"
)

type Credentials struct {
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password"`
}

var store = sessions.NewCookieStore([]byte("secret-key"))

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, this should be stricter to prevent cross-origin attacks
	},
}

func checkAuthenticated() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := store.Get(c.Request, "session-name")
		if err != nil {
			log.Println("Error getting session:", err)
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}

		if auth, ok := session.Values["loggedin"].(bool); !ok || !auth {
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}
		c.Next()
	}
}

func connectDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./database/CactuDash.db")
	if err != nil {
		return nil, err
	}
	return db, nil
}

func loginHandler(c *gin.Context) {
	var creds Credentials
	if err := c.Bind(&creds); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	db, err := connectDB()
	if err != nil {
		log.Println("Error connecting to database:", err)
		c.JSON(500, gin.H{"error": "database connection error"})
		return
	}
	defer db.Close()

	var username string
	var password string
	err = db.QueryRow("SELECT username, password FROM userlogin WHERE username = ?", creds.Username).Scan(&username, &password)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(401, gin.H{"error": "invalid credentials"})
		} else {
			log.Println("Error querying database:", err)
			c.JSON(500, gin.H{"error": "database query error"})
		}
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(password), []byte(creds.Password)) != nil {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}

	session, err := store.Get(c.Request, "session-name")
	if err != nil {
		log.Println("Error creating session:", err)
		c.JSON(500, gin.H{"error": "session error"})
		return
	}

	session.Values["loggedin"] = true
	if err := session.Save(c.Request, c.Writer); err != nil {
		log.Println("Error saving session:", err)
		c.JSON(500, gin.H{"error": "session save error"})
		return
	}

	c.Redirect(http.StatusFound, "/welcome")
}

func systemInfoHandler(c *gin.Context) {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to get hostname"})
		return
	}

	kernelVersion, err := exec.Command("uname", "-r").Output()
	if err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to get kernel version"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hostname":       hostname,
		"kernel_version": strings.TrimSpace(string(kernelVersion)),
	})
}

func diskUsageHandler(c *gin.Context) {
	usageStat, err := disk.Usage("/")
	if err != nil {
		log.Println("Error getting disk usage:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to get disk usage"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"used":  usageStat.Used,
		"total": usageStat.Total,
	})
}

func cactuDashDataHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"version": Version})
}

func reboot(c *gin.Context) {
	err := exec.Command("reboot").Run()
	if err != nil {
		log.Fatal(err)
	}
	c.JSON(http.StatusOK, gin.H{"reboot": "rebooting"})
}

// WebSocket handler
func wsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to set WebSocket upgrade:", err)
		return
	}
	defer conn.Close()

	for {
		usage, err := cpu.Percent(0, false)
		if err != nil {
			log.Println("Error getting CPU usage:", err)
			break
		}

		// Send the CPU usage to the client
		err = conn.WriteJSON(gin.H{"cpu_usage": usage[0]})
		if err != nil {
			log.Println("Error writing WebSocket message:", err)
			break
		}

		// Wait for 2 seconds before sending the next message
		time.Sleep(2 * time.Second)
	}
}

func main() {
	router := gin.Default()

	// Set trusted proxies
	err := router.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		log.Fatal("Error setting trusted proxies:", err)
	}

	router.Static("/static", "./static")

	router.GET("/", func(c *gin.Context) {
		c.File("sites/login.html")
	})

	router.POST("/auth", loginHandler)

	router.GET("/welcome", checkAuthenticated(), func(c *gin.Context) {
		c.File("sites/welcome.html")
	})

	// System info
	router.GET("/system-info", systemInfoHandler)
	router.GET("/disk-usage", diskUsageHandler) // New route for disk usage

	// WebSocket route
	router.GET("/ws", wsHandler)

	router.GET("/cactu-dash", cactuDashDataHandler)

	router.POST("/reboot", reboot)

	err = router.Run(":3030")
	if err != nil {
		log.Fatal("Error starting the server:", err)
	}
}
