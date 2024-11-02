package scripts

import (
	"bufio"
	"log"
	"os"
)

var (
	logfile        *os.File
	logger_Error   *log.Logger
	logger_Message *log.Logger
)

func checkLogFile() {
	file, err := os.OpenFile("logs/logsfile.txt", os.O_RDWR, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %s", err)
	}
	defer file.Close()

	lineCount := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineCount++
	}

	if lineCount > 525 {
		err = file.Truncate(0)
		if err != nil {
			log.Fatalf("Failed to truncate log file: %s", err)
		}
		_, err = file.Seek(0, 0)
		if err != nil {
			log.Fatalf("Failed to seek log file: %s", err)
		}
	}
}

func init() {
	var err error

	if _, err = os.Stat("logs"); os.IsNotExist(err) {
		err = os.Mkdir("logs", 0755)
		if err != nil {
			log.Fatalf("Failed to create logs directory: %s", err)
		}
	}

	logfile, err = os.OpenFile("logs/logsfile.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open error log file: %s", err)
	}

	logger_Error = log.New(logfile, "ERROR: ", log.Ldate|log.Ltime)

	logger_Message = log.New(logfile, "", log.Ldate|log.Ltime)

	checkLogFile()
}

func LogError(err error) {
	logger_Error.Printf("|-| %s", err)
}

func LogMessage(message string) {
	logger_Message.Printf("|-| %s", message)
}
