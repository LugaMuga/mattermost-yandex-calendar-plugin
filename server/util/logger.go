package util

import (
	"github.com/mattermost/mattermost-server/v6/plugin"
)

type Logger struct {
	pluginAPI plugin.API
}

func NewLogger(plugin plugin.API) *Logger {
	return &Logger{
		pluginAPI: plugin,
	}
}

func (l *Logger) LogDebug(msg string, keyValuePairs ...interface{}) {
	l.pluginAPI.LogDebug(msg, keyValuePairs)
}

func (l *Logger) LogInfo(msg string, keyValuePairs ...interface{}) {
	l.pluginAPI.LogInfo(msg, keyValuePairs)
}

func (l *Logger) LogWarn(msg string, userId *string, err error) {
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}
	l.logWarn(msg, &errMsg, userId)
}

func (l *Logger) Warn(msg string, userId *string) {
	l.logWarn(msg, nil, userId)
}

func (l *Logger) logWarn(msg string, err *string, userId *string) {
	if userId != nil && err != nil {
		l.pluginAPI.LogWarn(msg, "userId", userId, "error", err)
	} else if userId != nil {
		l.pluginAPI.LogWarn(msg, "userId", userId)
	} else if err != nil {
		l.pluginAPI.LogWarn(msg, "error", err)
	}
}

func (l *Logger) LogCustomError(msg string, keyValuePairs ...interface{}) {
	l.pluginAPI.LogError(msg, keyValuePairs)
}

func (l *Logger) LogError(msg string, userId *string, err error) {
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}
	l.logError(msg, &errMsg, userId)
}

func (l *Logger) Error(msg string, userId *string) {
	l.logError(msg, nil, userId)
}

func (l *Logger) logError(msg string, err *string, userId *string) {
	if userId != nil && err != nil {
		l.pluginAPI.LogError(msg, "userId", userId, "error", err)
	} else if userId != nil {
		l.pluginAPI.LogError(msg, "userId", userId)
	} else if err != nil {
		l.pluginAPI.LogError(msg, "error", err)
	}
}
