package main

import (
	"fmt"
	"io"

	"github.com/codeforge11/CactuDash/scripts"

	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
)

type Credentials struct {
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func checkAuthenticated() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := scripts.Store.Get(c.Request, "session-name")
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

		if session.Values["server_start_time"] == nil || session.Values["server_start_time"].(time.Time).Before(scripts.ServerStartTime) {
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

	if creds.Username == "" || creds.Password == "" {
		scripts.LogMessage("invalid credentials")
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}

	userInfo, err := exec.Command("getent", "passwd", creds.Username).Output()
	if err != nil || len(userInfo) == 0 {
		scripts.LogMessage("invalid user")
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}

	cmd := exec.Command("su", "-", creds.Username, "-c", "exit")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		scripts.LogError(err)
		c.JSON(500, gin.H{"error": "internal error"})
		return
	}
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, creds.Password+"\n")
	}()
	if err := cmd.Run(); err != nil {
		scripts.LogMessage("invalid password")
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}

	session, err := scripts.Store.Get(c.Request, "session-name")
	if err != nil {
		log.Println("Error creating session:", err)
		scripts.LogError(err)
		c.JSON(500, gin.H{"error": "session error"})
		return
	}

	session.Values["loggedin"] = true
	session.Values["expires_at"] = time.Now().Add(scripts.SessionExpiration)
	session.Values["server_start_time"] = scripts.ServerStartTime
	if err := session.Save(c.Request, c.Writer); err != nil {
		log.Println("Error saving session:", err)
		scripts.LogError(err)
		c.JSON(500, gin.H{"error": "session save error"})
		return
	}

	c.Redirect(http.StatusFound, "/welcome")

	scripts.CheckLogFile() // Checks the number of rulers

	scripts.LogMessage("Successful login")
}

func loginHandler_debug(c *gin.Context) {
	// running in debug

	var creds Credentials
	if err := c.Bind(&creds); err != nil {
		scripts.LogError(err)
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	if creds.Username == "debug" && creds.Password == "debug" {

		session, err := scripts.Store.Get(c.Request, "session-name")

		if err != nil {
			log.Println("Error creating session:", err)
			scripts.LogError(err)
			c.JSON(500, gin.H{"error": "session error"})
			return
		}

		session.Values["loggedin"] = true
		session.Values["expires_at"] = time.Now().Add(scripts.SessionExpiration)
		session.Values["server_start_time"] = scripts.ServerStartTime

		if err := session.Save(c.Request, c.Writer); err != nil {
			log.Println("Error saving session:", err)
			scripts.LogError(err)
			c.JSON(500, gin.H{"error": "session save error"})
			return
		}

		c.Redirect(http.StatusFound, "/welcome")

		scripts.CheckLogFile() // Checks the number of rulers

		scripts.LogMessage("Successful login in debug mode")
		return
	} else {
		scripts.LogMessage("invalid credentials in debug mode")
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}
}

func systemInfoHandler(c *gin.Context) {

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
		scripts.LogError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to get hostname"})
		return
	}

	distro, supportStatus := scripts.RetrieveDistroInfo()
	arch := runtime.GOARCH

	c.JSON(http.StatusOK, gin.H{
		"hostname":      hostname,
		"nameOfOs":      distro,
		"arch":          arch,
		"supportStatus": supportStatus,
	})
}

// Get disk usage
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

// Get CactuDash version
func cactuDashDataHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"version": scripts.Version})
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

		// Containers in WebSocket
		out, err := exec.Command("docker", "ps", "-a", "--format", "{{.ID}};{{.Image}};{{.Ports}};{{.Status}};{{.Names}}").Output()
		var containers []scripts.Container
		if err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				if line != "" {
					fields := strings.Split(line, ";")
					if len(fields) < 5 {
						continue
					}
					containers = append(containers, scripts.Container{
						Id:     fields[0],
						Image:  fields[1],
						Status: fields[3],
						Name:   fields[4],
					})
				}
			}
		}

		// Send the CPU usage and containers to the client
		err = conn.WriteJSON(gin.H{
			"cpu_usage":  usage[0],
			"containers": containers,
		})
		if err != nil {
			log.Println("Error writing WebSocket message:", err)
			scripts.LogError(err)
			break
		}

		checkAuthenticated()

		time.Sleep(10 * time.Second) //WebSocket refresh time
	}
}

func clearOldLogs(c *gin.Context) {
	cmd := exec.Command("/bin/sh", "-c", `rm -rf ./logs/old_logs`)

	if err := cmd.Run(); err != nil {
		log.Println("Error running clear cash script:", err)
		scripts.LogError(err)
	}
}

func main() {

	flag.Parse()

	if !*scripts.DebugMode {
		gin.SetMode(gin.ReleaseMode) //run server in release mode
		fmt.Println("Server starting in: " + (scripts.GetIpAddr().String()) + ":3030")

		if !scripts.CheckReq(nil) {
			panic("CactuDash cannot verify Docker installation. Check logs/logsfile.txt for details.")
		}

	}

	router := gin.Default()

	scripts.LogMessage("Server version: " + scripts.Version)

	// Set trusted proxies
	err := router.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		log.Fatal("Error setting trusted proxies:", err)
		scripts.LogError(err)
	}

	router.Static("/static", "./static")

	if !*scripts.DebugMode {
		if scripts.Checkdb(nil) {
			router.GET("/", func(c *gin.Context) {
				c.File("sites/login.html")
			})

		} else {
			router.GET("/", func(c *gin.Context) {
				c.File("sites/register.html")
			})
			router.POST("/register", scripts.Register)
		}
	} else {

		go func() {
			http.ListenAndServe("localhost:6060", nil)

			fmt.Println("Profiling started on localhost:6060")
			scripts.LogMessage("Profiling started on localhost:6060")
			log.Println(fmt.Println("Profiling started on localhost:6060"))

		}() //pprof in debug

		router.GET("/", func(c *gin.Context) {
			c.File("sites/login.html")
		})
	}

	if *scripts.DebugMode {
		router.POST("/auth", loginHandler_debug)
		scripts.LogMessage("Running in debug mode")
	} else {
		router.POST("/auth", loginHandler)
	}

	router.GET("/welcome", checkAuthenticated(), func(c *gin.Context) {
		c.File("sites/welcome.html")
	})

	// System info
	router.GET("/system-info", systemInfoHandler)

	router.GET("/disk-usage", diskUsageHandler) // New route for disk usage

	// WebSocket route
	router.GET("/ws", wsHandler)

	router.GET("/cactu-dash", cactuDashDataHandler)

	router.GET("/lastTag", scripts.GetLastGitTag)

	router.POST("/power", scripts.Power) //Reboot and shutdown function

	router.POST("/logout", scripts.Logout)

	router.POST("/update", scripts.Update) //Update function

	router.GET("/containers", scripts.GetContainers)

	router.POST("/toggle/:id", scripts.ToggleContainerState)

	router.POST("/restart/:id", scripts.RestartContainer)

	router.POST("/remove/:id", scripts.RemoveContainer)

	router.POST("/log", scripts.JsLog) //Logs from js file

	router.POST("/createDockerType", scripts.CreateDocker) //Create Docker or Docker Compose

	router.POST("/clearOldLogs", clearOldLogs)

	err = router.Run(":3030") //Start server on port 3030
	if err != nil {
		log.Panic("Error starting the server:", err)
		scripts.LogError(err)
	}
}
