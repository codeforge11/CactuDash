package scripts

import (
	"github.com/codeforge11/betterLogs"
)

var logConfig = betterLogs.Config{
	MainFileName:     "logsfile",
	MainFolder:       "logs",
	OldLogsFilesName: "",
	OldLogsFolder:    "old_logs",
}

var BetterLogs = betterLogs.New(logConfig)
