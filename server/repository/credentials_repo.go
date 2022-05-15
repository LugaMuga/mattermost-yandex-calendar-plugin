package repository

import (
	"encoding/json"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/dto"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/util"
	"github.com/mattermost/mattermost-server/v6/plugin"
)

type CredentialsRepo struct {
	logger    *util.Logger
	pluginAPI plugin.API
}

func NewCredentialsRepo(logger *util.Logger, plugin plugin.API) *CredentialsRepo {
	return &CredentialsRepo{
		logger:    logger,
		pluginAPI: plugin,
	}
}

func (cr *CredentialsRepo) SaveCredentials(userId string, credentials dto.Credentials) {
	jsonVal, marshalErr := json.Marshal(credentials)
	if marshalErr != nil {
		cr.logger.LogError("Error on marshal credentials", &userId, marshalErr)
	}
	err := cr.pluginAPI.KVSet(userId+credentialsKey, jsonVal)
	if err != nil {
		cr.logger.LogError("Error on save credentials to storage", &userId, marshalErr)
	}
}

func (cr *CredentialsRepo) GetCredentials(userId string) *dto.Credentials {
	bytes, kvErr := cr.pluginAPI.KVGet(userId + credentialsKey)
	if kvErr != nil {
		cr.logger.LogError("Error on getting credentials from storage", &userId, kvErr)
	}
	if bytes == nil {
		cr.logger.Error("Couldn't find credentials", &userId)
		return nil
	}
	var credentials *dto.Credentials
	err := json.Unmarshal(bytes, &credentials)
	if err != nil {
		cr.logger.LogError("Error on parse credentials from storage", &userId, err)
		return nil
	}
	return credentials
}
