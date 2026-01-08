package scripts

import (
	"errors"
	"net/http"

	"github.com/codeforge11/betterLogs"
	"github.com/gin-gonic/gin"
)

func JsLog(c *gin.Context) {

	var jsLogMes struct {
		Type    bool   `json:"type"`
		Message string `json:"message"`
	}

	if err := c.BindJSON(&jsLogMes); err != nil {
		betterLogs.LogError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if jsLogMes.Type { //Normal logs
		betterLogs.LogMessage(jsLogMes.Message)
		c.Status(http.StatusOK)

	} else { //Errors
		betterLogs.LogError(errors.New(jsLogMes.Message))
		c.Status(http.StatusOK)
	}
}
