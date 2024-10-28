package scripts

import (
	"log"
	"os"
)

var (
	errorLog *os.File
	logger   *log.Logger
)

func init() {
	var err error

	if _, err = os.Stat("logs"); os.IsNotExist(err) {
		err = os.Mkdir("logs", 0755)
		if err != nil {
			log.Fatalf("Failed to create logs directory: %s", err)
		}
	}

	errorLog, err = os.OpenFile("logs/error_log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open error log file: %s", err)
	}

	logger = log.New(errorLog, "ERROR: ", log.Ldate|log.Ltime)
}

func LogError(err error) {
	logger.Printf("|-| %s", err)
}
