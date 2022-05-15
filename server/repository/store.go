package repository

import (
	"encoding/json"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/dto"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/mattermost/mattermost-server/v6/shared/mlog"
	"strconv"
	"time"
)

func SaveCalendarHomeSet(pluginAPI plugin.API, userId string, calendarHomeSet string) {
	err := pluginAPI.KVSet(userId+calendarHomeSetKey, []byte(calendarHomeSet))
	if err != nil {
		mlog.Error("Error on save HomeSet for user:"+userId, mlog.Err(err))
	}
}

func GetCalendarHomeSet(pluginAPI plugin.API, userId string) string {
	calendarHomeSetBytes, _ := pluginAPI.KVGet(userId + calendarHomeSetKey)
	if calendarHomeSetBytes != nil {
		return string(calendarHomeSetBytes)
	}
	return ""
}

func SaveLastUpdate(pluginAPI plugin.API, userId string, lastUpdate time.Time) {
	err := pluginAPI.KVSet(userId+lastUpdateKey, []byte(lastUpdate.Format(time.RFC3339)))
	if err != nil {
		mlog.Error("Error on save lastUpdated for user:"+userId, mlog.Err(err))
	}
}

func GetUserCalendarLastUpdate(pluginAPI plugin.API, userId string) *time.Time {
	lastUpdateBytes, _ := pluginAPI.KVGet(userId + lastUpdateKey)
	if lastUpdateBytes != nil {
		lastUpdate, _ := time.Parse(time.RFC3339, string(lastUpdateBytes))
		return &lastUpdate
	}
	return nil
}

func GetUserCronJobIds(pluginAPI plugin.API, userId string) (*int, *int) {
	var eventCronId *int
	var updateCronId *int
	eventCronIdBytes, _ := pluginAPI.KVGet(userId + eventCronIdKey)
	if eventCronIdBytes != nil {
		val, _ := strconv.Atoi(string(eventCronIdBytes))
		eventCronId = &val
	}
	updateCronIdBytes, _ := pluginAPI.KVGet(userId + updateCronIdKey)
	if updateCronIdBytes != nil {
		val, _ := strconv.Atoi(string(updateCronIdBytes))
		updateCronId = &val
	}
	return eventCronId, updateCronId
}

func SaveEventCronJob(pluginAPI plugin.API, userId string, eventCronJonId int) {
	err := pluginAPI.KVSet(userId+eventCronIdKey, []byte(strconv.Itoa(eventCronJonId)))
	if err != nil {
		mlog.Error("Error on save EventCronId for user:"+userId, mlog.Err(err))
	}
}

func SaveUpdateCronJob(pluginAPI plugin.API, userId string, updateCronJonId int) {
	err := pluginAPI.KVSet(userId+updateCronIdKey, []byte(strconv.Itoa(updateCronJonId)))
	if err != nil {
		mlog.Error("Error on save UpdateCronId for user:"+userId, mlog.Err(err))
	}
}

func SaveEvents(pluginAPI plugin.API, userId string, events []dto.Event) {
	jsonVal, marshalErr := json.Marshal(events)
	if marshalErr != nil {
		mlog.Error("Error on marshal events for user:"+userId, mlog.Err(marshalErr))
	}
	err := pluginAPI.KVSet(userId+eventsKey, jsonVal)
	if err != nil {
		mlog.Error("Error on save state to store for user:"+userId, mlog.Err(err))
	}
}

func GetEvents(pluginAPI plugin.API, userId string) []dto.Event {
	bytes, kvErr := pluginAPI.KVGet(userId + eventsKey)
	if kvErr != nil {
		mlog.Error("Error on getting events from storage for user:"+userId, mlog.Err(kvErr))
	}
	if bytes == nil {
		mlog.Warn("Couldn't find events for user:" + userId)
		return nil
	}
	var events []dto.Event
	err := json.Unmarshal(bytes, &events)
	if err != nil {
		mlog.Warn("Error on parse events from storage for user:"+userId, mlog.Err(err))
		return nil
	}
	return events
}

func SaveSettings(pluginAPI plugin.API, userId string, settings dto.Settings) {
	settingsJson, marshalErr := json.Marshal(settings)
	if marshalErr != nil {
		mlog.Error("Error on Marshal settings for user:"+userId, mlog.Err(marshalErr))
	}
	err := pluginAPI.KVSet(userId+settingsKey, settingsJson)
	if err != nil {
		mlog.Error("Error on save settings to store for user:"+userId, mlog.Err(err))
	}
}

func GetSettings(pluginAPI plugin.API, userId string) *dto.Settings {
	settingBytes, kvErr := pluginAPI.KVGet(userId + settingsKey)
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

func SaveState(pluginAPI plugin.API, userId string, state dto.State) {
	jsonVal, marshalErr := json.Marshal(state)
	if marshalErr != nil {
		mlog.Error("Error on Marshal state for user:"+userId, mlog.Err(marshalErr))
	}
	err := pluginAPI.KVSet(userId+stateKey, jsonVal)
	if err != nil {
		mlog.Error("Error on save state to store for user:"+userId, mlog.Err(err))
	}
}

func GetState(pluginAPI plugin.API, userId string) *dto.State {
	stateBytes, kvErr := pluginAPI.KVGet(userId + stateKey)
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
