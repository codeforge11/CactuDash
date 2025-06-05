package scripts

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

type Tag struct {
	Name string `json:"name"`
}

var (
	LastGitTag     string
	lastGitTagLock sync.RWMutex
)

func GetLastGitTagName() (string, error) {
	url := "https://api.github.com/repos/codeforge11/CactuDash/tags"

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to fetch tags")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tags []Tag
	err = json.Unmarshal(body, &tags)
	if err != nil {
		return "", err
	}

	if len(tags) > 0 {
		lastGitTagLock.Lock()
		LastGitTag = tags[0].Name
		lastGitTagLock.Unlock()
		return tags[0].Name, nil
	}
	return "", errors.New("no tags found")
}

func GetLastGitTag(c *gin.Context) {
	tag, err := GetLastGitTagName()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"last_tag": tag})
}
