package repository

import (
	"encoding/json"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/conf"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/dto"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/mattermost/mattermost-server/v6/shared/mlog"
	"strconv"
	"time"
)

func GetUserCalendarLastUpdate(pluginAPI plugin.API, userId string) *time.Time {
	lastUpdateBytes, _ := pluginAPI.KVGet(userId + conf.LastUpdate)
	if lastUpdateBytes != nil {
		lastUpdate, _ := time.Parse(time.RFC3339, string(lastUpdateBytes))
		return &lastUpdate
	}
	return nil
}

func GetUserCronJobIds(pluginAPI plugin.API, userId string) (*int, *int) {
	var eventCronId *int
	var updateCronId *int
	eventCronIdBytes, _ := pluginAPI.KVGet(userId + conf.EventCronId)
	if eventCronIdBytes != nil {
		val, _ := strconv.Atoi(string(eventCronIdBytes))
		eventCronId = &val
	}
	updateCronIdBytes, _ := pluginAPI.KVGet(userId + conf.UpdateCronId)
	if updateCronIdBytes != nil {
		val, _ := strconv.Atoi(string(updateCronIdBytes))
		updateCronId = &val
	}
	return eventCronId, updateCronId
}

func SaveLastUpdate(pluginAPI plugin.API, userId string, lastUpdate time.Time) {
	pluginAPI.KVSet(userId+conf.LastUpdate, []byte(lastUpdate.Format(time.RFC3339)))
}

func SaveSettings(pluginAPI plugin.API, userId string, settings dto.Settings) {
	settingsJson, marshalErr := json.Marshal(settings)
	if marshalErr != nil {
		mlog.Error("Error on Marshal settings for user:"+userId, mlog.Err(marshalErr))
	}
	err := pluginAPI.KVSet(userId+conf.Settings, settingsJson)
	if err != nil {
		mlog.Error("Error on save settings to store for user:"+userId, mlog.Err(err))
	}
}

func GetSettings(pluginAPI plugin.API, userId string) *dto.Settings {
	settingBytes, kvErr := pluginAPI.KVGet(userId + conf.Settings)
	if kvErr != nil {
		mlog.Error("Error on getting settings from store for user:"+userId, mlog.Err(kvErr))
	}
	if settingBytes == nil {
		mlog.Warn("Couldn't find settings for user:" + userId)
		return nil
	}
	var settings *dto.Settings
	err := json.Unmarshal(settingBytes, &settings)
	if err != nil {
		mlog.Warn("Error on parse settings from storage for user:"+userId, mlog.Err(err))
		return nil
	}
	return settings
}

func GetCalendarHomeSet(pluginAPI plugin.API, userId string) string {
	calendarHomeSetBytes, _ := pluginAPI.KVGet(userId + conf.HomeSet)
	if calendarHomeSetBytes != nil {
		return string(calendarHomeSetBytes)
	}
	return ""
}

func SaveState(pluginAPI plugin.API, userId string, state dto.State) {
	stateJson, marshalErr := json.Marshal(state)
	if marshalErr != nil {
		mlog.Error("Error on Marshal state for user:"+userId, mlog.Err(marshalErr))
	}
	err := pluginAPI.KVSet(userId+conf.State, stateJson)
	if err != nil {
		mlog.Error("Error on save state to store for user:"+userId, mlog.Err(err))
	}
}

func GetState(pluginAPI plugin.API, userId string) *dto.State {
	stateBytes, kvErr := pluginAPI.KVGet(userId + conf.State)
	if kvErr != nil {
		mlog.Error("Error on getting state from store for user:"+userId, mlog.Err(kvErr))
	}
	if stateBytes == nil {
		mlog.Warn("Couldn't find state for user:" + userId)
		return nil
	}
	var state *dto.State
	err := json.Unmarshal(stateBytes, &state)
	if err != nil {
		mlog.Warn("Error on parse state from storage for user:"+userId, mlog.Err(err))
		return nil
	}
	return state
}

func SaveCredentials(pluginAPI plugin.API, userId string, credentials dto.Credentials) {
	json, marshalErr := json.Marshal(credentials)
	if marshalErr != nil {
		mlog.Error("Error on marshal credentials for user:"+userId, mlog.Err(marshalErr))
	}
	err := pluginAPI.KVSet(userId+conf.Credentials, json)
	if err != nil {
		mlog.Error("Error on save credentials to storage for user:"+userId, mlog.Err(err))
	}
}

func GetCredentials(pluginAPI plugin.API, userId string) *dto.Credentials {
	bytes, kvErr := pluginAPI.KVGet(userId + conf.Credentials)
	if kvErr != nil {
		mlog.Error("Error on getting credentials from storage for user:"+userId, mlog.Err(kvErr))
	}
	if bytes == nil {
		mlog.Warn("Couldn't find credentials for user:" + userId)
		return nil
	}
	var credentials *dto.Credentials
	err := json.Unmarshal(bytes, &credentials)
	if err != nil {
		mlog.Warn("Error on parse credentials from storage for user:"+userId, mlog.Err(err))
		return nil
	}
	return credentials
}
