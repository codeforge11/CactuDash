package scripts

import (
	"bufio"
	"log"
	"os"
	"runtime"
	"strings"
)

func RetrieveDistroInfo() (string, bool) {
	var detectedID, osName, detectedIDLike string
	var supportStatus bool

	file, err := os.Open("/etc/os-release") //for detect distro name

	if err != nil {
		log.Println("Error opening /etc/os-release:", err)
		LogError(err)

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
			LogError(err)
		}
	}

	if runtime.GOOS == "linux" {
		osName = detectedID
	} else {
		osName = runtime.GOOS
	}
	switch osName {
	case detectedID:
		supportStatus = true
	case detectedIDLike:
		supportStatus = true
	default:
		supportStatus = false
	}

	return osName, supportStatus
}
