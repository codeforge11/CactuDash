package main

import (
	"github.com/codeforge11/CactuDash/static/scripts"

	"bufio"
	"database/sql"
	"encoding/gob"
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"

	"golang.org/x/crypto/bcrypt"
)

type Credentials struct {
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password"`
}

// Structure for containers
type Container struct {
	Id     string `json:"Id"`
	Image  string `json:"Image"`
	Status string `json:"Status"`
	Name   string `json:"Name"`
}

var store sessions.Store

func init() {
	gob.Register(time.Time{})
	store = sessions.NewCookieStore([]byte("secret-key"))
	cookieStore := store.(*sessions.CookieStore)

	cookieStore.Options = &sessions.Options{
		MaxAge:   int(sessionExpiration.Seconds()),
		HttpOnly: true,
	}
}

var sessionExpiration = 10 * time.Minute // session time
var serverStartTime = time.Now()

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func connectDB() (*sql.DB, error) {
	connStr := "root:CactuDash@tcp(127.0.0.1:3031)/CactuDB" // Connect to MariaDB
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		scripts.LogError(err)
		return nil, err

	}
	return db, nil
}

func checkAuthenticated() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := store.Get(c.Request, "session-name")
		if err != nil {
			log.Println("Error getting session:", err)
			scripts.LogError(err)
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}

		if auth, ok := session.Values["loggedin"].(bool); !ok || !auth {
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}

		if session.Values["expires_at"] == nil || time.Now().After(session.Values["expires_at"].(time.Time)) {
			session.Values["loggedin"] = false
			session.Save(c.Request, c.Writer)
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}

		if session.Values["server_start_time"] == nil || session.Values["server_start_time"].(time.Time).Before(serverStartTime) {
			session.Values["loggedin"] = false
			session.Save(c.Request, c.Writer)
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}

		c.Next()
	}
}

func loginHandler(c *gin.Context) {
	var creds Credentials
	if err := c.Bind(&creds); err != nil {
		scripts.LogError(err)
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	db, err := connectDB()
	if err != nil {
		log.Println("Error connecting to database:", err)
		scripts.LogError(err)
		c.JSON(500, gin.H{"error": "The database has encountered an issue"})
		return
	}
	defer db.Close()

	var username string
	var password string
	err = db.QueryRow("SELECT username, password FROM userlogin WHERE username = ?", creds.Username).Scan(&username, &password)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(401, gin.H{"error": "invalid credentials"})
			scripts.LogMessage("invalid credentials")
		} else {
			log.Println("Error querying database:", err)
			scripts.LogError(err)
			c.JSON(500, gin.H{"error": "database didn't work"})
		}
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(password), []byte(creds.Password)) != nil {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		scripts.LogMessage("invalid credentials")
		return
	}

	session, err := store.Get(c.Request, "session-name")
	if err != nil {
		log.Println("Error creating session:", err)
		scripts.LogError(err)
		c.JSON(500, gin.H{"error": "session error"})
		return
	}

	session.Values["loggedin"] = true
	session.Values["expires_at"] = time.Now().Add(sessionExpiration)
	session.Values["server_start_time"] = serverStartTime
	if err := session.Save(c.Request, c.Writer); err != nil {
		log.Println("Error saving session:", err)
		scripts.LogError(err)
		c.JSON(500, gin.H{"error": "session save error"})
		return
	}

	c.Redirect(http.StatusFound, "/welcome")

	scripts.CheckLogFile() //Checks the number of rulers

	scripts.LogMessage("Successful login")
}

func systemInfoHandler(c *gin.Context) {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
		scripts.LogError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to get hostname"})
		return
	}

	nameOfOs := runtime.GOOS
	arch := runtime.GOARCH

	if nameOfOs == "linux" {
		file, err := os.Open("/etc/os-release") //for detect distro name
		if err != nil {
			log.Println("Error opening /etc/os-release:", err)
			scripts.LogError(err)
		} else {
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "NAME=") {
					nameOfOs = strings.TrimPrefix(line, "NAME=")
					nameOfOs = strings.Trim(nameOfOs, `"`)
					break
				}
			}
			if err := scanner.Err(); err != nil {
				log.Println("Error reading file:", err)
				scripts.LogError(err)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"hostname": hostname,
		"nameOfOs": nameOfOs,
		"arch":     arch,
	})
}

func diskUsageHandler(c *gin.Context) {
	usageStat, err := disk.Usage("/")
	if err != nil {
		log.Println("Error getting disk usage:", err)
		scripts.LogError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to get disk usage"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"used":  usageStat.Used,
		"total": usageStat.Total,
		"free":  usageStat.Free,
	})
}

func cactuDashDataHandler(c *gin.Context) {
	var ServerVersion string = scripts.Version
	c.JSON(http.StatusOK, gin.H{"version": ServerVersion})
	scripts.LogMessage("Server version: " + ServerVersion)
}

func reboot(c *gin.Context) {
	session, err := store.Get(c.Request, "session-name")
	if err != nil {
		scripts.LogError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "session error"})
		return
	}

	// End the session
	session.Values["loggedin"] = false
	session.Save(c.Request, c.Writer)

	err = exec.Command("reboot").Run()
	if err != nil {
		scripts.LogError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reboot"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reboot": "rebooting"})
	scripts.LogMessage("Restart server...")
}

// WebSocket handler
func wsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to set WebSocket upgrade:", err)
		scripts.LogError(err)
		return
	}
	defer conn.Close()

	for {
		usage, err := cpu.Percent(0, false)
		if err != nil {
			log.Println("Error getting CPU usage:", err)
			scripts.LogError(err)
			break
		}

		// Send the CPU usage to the client
		err = conn.WriteJSON(gin.H{"cpu_usage": usage[0]})
		if err != nil {
			log.Println("Error writing WebSocket message:", err)
			scripts.LogError(err)
			break
		}

		// Wait for 2 seconds before sending the next message
		time.Sleep(2 * time.Second)
	}
}

// func update(c *gin.Context) {
// 	cmd := exec.Command("/bin/sh", "static/scripts/update.sh")
// 	err := cmd.Run()
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to run update script"})
// 		logError(err)
// 		return
// 	}
// 	c.JSON(http.StatusOK, gin.H{"status": "update script executed"})
// }

// Function to get Docker containers
func getContainers(c *gin.Context) {
	out, err := exec.Command("docker", "ps", "-a", "--format", "{{.ID}};{{.Image}};{{.Ports}};{{.Status}};{{.Names}}").Output()
	if err != nil {
		log.Println("Error executing docker command:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error executing docker command"})
		scripts.LogError(err)
		return
	}

	var containers []Container
	for _, line := range strings.Split(string(out), "\n") {
		if line != "" {
			fields := strings.Split(line, ";")
			if len(fields) < 5 {
				log.Println("Unexpected format in docker output:", line)
				continue
			}
			if !strings.Contains(fields[2], "3031") { // ignore 3031 port
				containers = append(containers, Container{Id: fields[0], Image: fields[1], Status: fields[3], Name: fields[4]})
			}
		}
	}

	c.JSON(http.StatusOK, containers)
}

func start_stopContainer(c *gin.Context) {
	id := strings.TrimPrefix(c.Param("id"), "/toggle/")
	out, err := exec.Command("docker", "inspect", "--format={{.State.Running}}", id).Output()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		scripts.LogError(err)
		return
	}

	running := strings.TrimSpace(string(out))
	if running == "true" {
		if err := exec.Command("docker", "stop", id).Run(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stop container"})
			scripts.LogError(err)
			return
		}
		scripts.LogMessage("Container stopped:" + id)
	} else {
		if err := exec.Command("docker", "start", id).Run(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start container"})
			scripts.LogError(err)
			return
		}
		scripts.LogMessage("Container started:" + id)
		// fmt.Println("Container started:", id)
	}
	c.Status(http.StatusNoContent)
}

func jsLog(c *gin.Context) {

	var jsLogMes struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	}

	if err := c.BindJSON(&jsLogMes); err != nil {
		scripts.LogError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	switch jsLogMes.Type {
	case "LogError":
		scripts.LogError(errors.New(jsLogMes.Message))
		c.Status(http.StatusOK)

	case "LogMessage":
		scripts.LogMessage(jsLogMes.Message)
		c.Status(http.StatusOK)

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid log type"})
	}
}

func main() {
	router := gin.Default()

	// Set trusted proxies
	err := router.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		log.Fatal("Error setting trusted proxies:", err)
		scripts.LogError(err)
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

	router.POST("/reboot", reboot) //Reboot function

	router.POST("/update", scripts.Update) //Update function

	router.GET("/containers", getContainers)

	router.POST("/toggle/:id", start_stopContainer)

	router.POST("/log", jsLog)

	err = router.Run(":3030")
	if err != nil {
		log.Fatal("Error starting the server:", err)
		scripts.LogError(err)
	}
}
