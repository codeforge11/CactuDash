package scripts

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"strings"
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

func compareVersions(x, y string) int {
	x = strings.TrimPrefix(x, "v")
	y = strings.TrimPrefix(y, "v")
	xs := strings.Split(x, ".")
	ys := strings.Split(y, ".")
	for i := 0; i < len(xs) && i < len(ys); i++ {
		if xs[i] != ys[i] {
			return strings.Compare(xs[i], ys[i])
		}
	}
	return len(xs) - len(ys)
}

func GetLastGitTagName() (string, error) {
	url := "https://api.github.com/repos/codeforge11/CactuDash/tags?per_page=100"

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

	if len(tags) == 0 {
		return "", errors.New("no tags found")
	}

	sort.Slice(tags, func(i, j int) bool {
		return compareVersions(tags[i].Name, tags[j].Name) > 0
	})

	lastGitTagLock.Lock()
	LastGitTag = tags[0].Name
	lastGitTagLock.Unlock()
	return tags[0].Name, nil
}

func GetLastGitTag(c *gin.Context) {
	tag, err := GetLastGitTagName()
	if err != nil {
		LogError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"last_tag": tag})
	// LogMessage("Last tag: " + tag)
}
