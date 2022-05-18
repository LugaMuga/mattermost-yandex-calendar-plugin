package repository

import (
	"encoding/json"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/util"
	"github.com/mattermost/mattermost-server/v6/plugin"
)

type WorkspaceRepo struct {
	logger    *util.Logger
	pluginAPI plugin.API
	usersKey  string
}

func NewWorkspaceRepo(logger *util.Logger, plugin plugin.API) *WorkspaceRepo {
	return &WorkspaceRepo{
		logger:    logger,
		pluginAPI: plugin,
		usersKey:  "users",
	}
}

func (wr *WorkspaceRepo) SaveUserIds(userIds map[string]bool) {
	userIdsJson, marshalErr := json.Marshal(userIds)
	if marshalErr != nil {
		wr.logger.LogError("Error on marshal user ids", nil, marshalErr)
	}
	err := wr.pluginAPI.KVSet(wr.usersKey, userIdsJson)
	if err != nil {
		wr.logger.LogError("Error on save user ids to storage", nil, err)
	}
}

func (wr *WorkspaceRepo) GetUserIds() *map[string]bool {
	userIdBytes, kvErr := wr.pluginAPI.KVGet(wr.usersKey)
	if kvErr != nil {
		wr.logger.LogWarn("Couldn't find user ids", nil, kvErr)
	}
	if userIdBytes == nil {
		wr.logger.Warn("Couldn't find user ids", nil)
		return nil
	}
	var userIds *map[string]bool
	err := json.Unmarshal(userIdBytes, &userIds)
	if err != nil {
		wr.logger.LogWarn("Couldn't find user ids", nil, err)
		return nil
	}
	return userIds
}

func (wr *WorkspaceRepo) DeleteUser(userId string) {
	wr.deleteKeyForUser(userId, credentialsKey)
	wr.deleteKeyForUser(userId, calendarHomeSetKey)
	wr.deleteKeyForUser(userId, eventsKey)
	wr.deleteKeyForUser(userId, lastUpdateKey)
	wr.deleteKeyForUser(userId, settingsKey)
	wr.deleteKeyForUser(userId, stateKey)
	wr.deleteKeyForUser(userId, eventCronIdKey)
	wr.deleteKeyForUser(userId, updateCronIdKey)
}

func (wr *WorkspaceRepo) deleteKeyForUser(userId string, key string) {
	err := wr.pluginAPI.KVDelete(userId + key)
	if err != nil {
		wr.logger.LogError("Error on delete "+key, &userId, err)
	}
}
