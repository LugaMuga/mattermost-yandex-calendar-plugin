package service

import (
	"encoding/json"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/conf"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/mattermost/mattermost-server/v6/shared/mlog"
	"sync"
)

type Workspace struct {
	pluginAPI plugin.API
	sync.Mutex
}

func NewWorkspaceService(plugin plugin.API) *Workspace {
	return &Workspace{pluginAPI: plugin}
}

func (w *Workspace) AddUser(userId string) {
	w.Lock()
	defer w.Unlock()
	userIds := w.GetUserIds()
	if userIds == nil {
		userIds = make(map[string]bool)
	}
	if !userIds[userId] {
		userIds[userId] = true
		w.saveUserIds(userIds)
	}
}

//nolint:golint,errcheck
func (w *Workspace) DeleteUser(userId string) {
	w.Lock()
	defer w.Unlock()
	userIds := w.GetUserIds()
	if userIds[userId] {
		delete(userIds, userId)
		w.saveUserIds(userIds)
	}
	w.pluginAPI.KVDelete(userId + conf.Credentials)
	w.pluginAPI.KVDelete(userId + conf.HomeSet)
	w.pluginAPI.KVDelete(userId + conf.Events)
	w.pluginAPI.KVDelete(userId + conf.LastUpdate)
	w.pluginAPI.KVDelete(userId + conf.Settings)
	w.pluginAPI.KVDelete(userId + conf.State)
	w.pluginAPI.KVDelete(userId + conf.EventCronId)
	w.pluginAPI.KVDelete(userId + conf.UpdateCronId)
}

func (w *Workspace) saveUserIds(userIds map[string]bool) {
	userIdsJson, marshalErr := json.Marshal(userIds)
	if marshalErr != nil {
		mlog.Error("Error on marshal user ids", mlog.Err(marshalErr))
	}
	err := w.pluginAPI.KVSet(conf.Users, userIdsJson)
	if err != nil {
		mlog.Error("Error on save user ids to storage", mlog.Err(err))
	}
}

func (w *Workspace) GetUserIds() map[string]bool {
	userIdBytes, kvErr := w.pluginAPI.KVGet(conf.Users)
	if kvErr != nil {
		mlog.Error("Error on getting user ids from storage", mlog.Err(kvErr))
	}
	if userIdBytes == nil {
		mlog.Warn("Couldn't find user ids")
		return make(map[string]bool)
	}
	var userIds map[string]bool
	err := json.Unmarshal(userIdBytes, &userIds)
	if err != nil {
		mlog.Warn("Error on parse user ids from storage", mlog.Err(err))
		return nil
	}
	return userIds
}
