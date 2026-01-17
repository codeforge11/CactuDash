package scripts

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
	_ "github.com/glebarez/go-sqlite"
	"golang.org/x/crypto/bcrypt"
)

const Dbfile string = "CactuDB.db" //database file

type DBCredentials struct {
	Passwd string `form:"passwd2" json:"passwd2"`
}

func Register(c *gin.Context) {
	var creds DBCredentials
	if err := c.ShouldBind(&creds); err != nil {
		BetterLogs.LogError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		log.Println(err)
		return
	}

	// ? Hashing password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Passwd), bcrypt.DefaultCost)
	if err != nil {
		BetterLogs.LogError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		log.Println("Failed to hash password.")
		return
	}

	// Create the database file if it does not exist
	if _, err := os.Stat(Dbfile); os.IsNotExist(err) {

		if err := os.MkdirAll("./", os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create database directory"})
			BetterLogs.LogError(err)
			log.Println("Failed to create database directory.")
			return
		}

		file, err := os.Create(Dbfile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create database file"})
			BetterLogs.LogError(err)
			log.Println("Failed to create database file.")
			return
		}
		file.Close()
	}

	//Opening database f
	sqlDB, err := sql.Open("sqlite", Dbfile)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open SQLite database"})
		BetterLogs.LogError(err)
		return
	}
	// defer sqlDB.Close()

	// time.Sleep(3 * time.Second)

	// Create users table if not exists
	_, err = sqlDB.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE,
		password TEXT
	)`)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create users table"})
		BetterLogs.LogError(err)
		return
	}

	_, err = sqlDB.Exec(`INSERT OR REPLACE INTO users (username, password) VALUES (?, ?)`, "admin", hashedPassword) // place for admin user hashed password
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save admin password"})
		BetterLogs.LogError(err)
		return
	}

	if !*DebugMode {
		err = exec.Command("reboot").Run()
		if err != nil {
			log.Println(err)
			BetterLogs.LogError(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reboot"})
			return
		}
	}

}

func Checkdb(c *gin.Context) bool {

	_, err := os.Stat(Dbfile)
	if os.IsNotExist(err) {
		// Database file doesn't exist
		log.Println("Database file not detected")
		if c != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database file not detected"})
		}
		BetterLogs.LogError(err)
		BetterLogs.LogMessage("Database file not detected")

		return false
	}
	return true
}

func CheckType(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"dbExists": Checkdb(nil),
	})
}
