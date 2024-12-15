package scripts

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Tag struct {
	Name string `json:"name"`
}

func GetLastGitTag(c *gin.Context) {
	url := "https://api.github.com/repos/codeforge11/CactuDash/tags" //url of project

	resp, err := http.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tags"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
			return
		}

		var tags []Tag
		err = json.Unmarshal(body, &tags)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal JSON"})
			return
		}

		if len(tags) > 0 {
			c.JSON(http.StatusOK, gin.H{"last_tag": tags[0].Name})
			return
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "No tags found"})
			return
		}
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tags"})
}
