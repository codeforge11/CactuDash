package scripts

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func JsLog(c *gin.Context) {

	var jsLogMes struct {
		Type    bool   `json:"type"`
		Message string `json:"message"`
	}

	if err := c.BindJSON(&jsLogMes); err != nil {
		LogError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if jsLogMes.Type { //Normal logs
		LogMessage(jsLogMes.Message)
		c.Status(http.StatusOK)

	} else { //Errors
		LogError(errors.New(jsLogMes.Message))
		c.Status(http.StatusOK)
	}
}
