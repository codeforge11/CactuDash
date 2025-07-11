package main

import (
	"fmt"
	"io"
	"net"

	"github.com/codeforge11/CactuDash/scripts"

	"bufio"
	"flag"
	"log"
	"net/http"
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

// Structure for containers
type Container struct {
	Id     string `json:"Id"`
	Image  string `json:"Image"`
	Status string `json:"Status"`
	Name   string `json:"Name"`
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

// Get server device ip
func getIpAddr() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
		scripts.LogError(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

func systemInfoHandler(c *gin.Context) {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
		scripts.LogError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to get hostname"})
		return
	}

	var supportStatus bool
	var detectedID string
	var detectedIDLike string
	var osName string
	arch := runtime.GOARCH

	file, err := os.Open("/etc/os-release")            //for detect distro name
	if (err != nil) && (*scripts.DebugMode == false) { //for not showing in debug mode
		log.Println("Error opening /etc/os-release:", err)
		scripts.LogError(err)
	} else {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "ID=") {
				detectedID = strings.TrimPrefix(line, "ID=")
				detectedID = strings.Trim(detectedID, `"`)

			}
			if strings.HasPrefix(line, "ID_LIKE=") {
				detectedIDLike = strings.TrimPrefix(line, "ID_LIKE=")
				detectedIDLike = strings.Trim(detectedIDLike, `"`)

			}
		}
		if err := scanner.Err(); err != nil {
			log.Println("Error reading file:", err)
			scripts.LogError(err)
		}
	}

	if runtime.GOOS == "linux" {
		osName = detectedID
	} else {
		osName = runtime.GOOS
	}

	if osName == detectedID || osName == detectedIDLike {
		supportStatus = true
	} else {
		supportStatus = false
	}

	c.JSON(http.StatusOK, gin.H{
		"hostname":      hostname,
		"nameOfOs":      osName,
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
		var containers []Container
		if err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				if line != "" {
					fields := strings.Split(line, ";")
					if len(fields) < 5 {
						continue
					}
					containers = append(containers, Container{
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

		time.Sleep(10 * time.Second) //WebSocket refresh time
	}
}

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
			containers = append(containers, Container{Id: fields[0], Image: fields[1], Status: fields[3], Name: fields[4]})
		}
	}

	c.JSON(http.StatusOK, containers)
}

// Function to change docker state
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

// Function to restart docker
func restartContainer(c *gin.Context) {
	id := strings.TrimPrefix(c.Param("id"), "/restart/")

	if err := exec.Command("docker", "restart", id).Run(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restart container"})
		scripts.LogError(err)
		return
	}
	scripts.LogMessage("Container restarted:" + id)

}

// Function to remove docker
func removeContainer(c *gin.Context) {
	id := strings.TrimPrefix(c.Param("id"), "/remove/")

	if err := exec.Command("docker", "rm", id).Run(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove container"})
		scripts.LogError(err)
		return
	}
	scripts.LogMessage("Container removed:" + id)

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
		fmt.Println("Server starting in: " + (getIpAddr().String()) + ":3030")
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

	router.GET("/", func(c *gin.Context) {
		c.File("sites/login.html")
	})

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

	router.GET("/containers", getContainers)

	router.POST("/toggle/:id", start_stopContainer)

	router.POST("/restart/:id", restartContainer)

	router.POST("/remove/:id", removeContainer)

	router.POST("/log", scripts.JsLog) //Logs from js file

	router.POST("/createDockerType", scripts.CreateDocker) //Create Docker or Docker Compose

	router.POST("/clearOldLogs", clearOldLogs)

	err = router.Run(":3030") //Start server on port 3030
	if err != nil {
		log.Panic("Error starting the server:", err)
		scripts.LogError(err)
	}
}
