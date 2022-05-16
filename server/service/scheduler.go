package service

import (
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/repository"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/util"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/robfig/cron/v3"
)

const (
	// UserEventHandlerCronSpec @every 1m
	UserEventHandlerCronSpec = "CRON_TZ=UTC */1 * * * *"
	// UserEventUpdaterCronSpec @every 10m
	UserEventUpdaterCronSpec = "CRON_TZ=UTC */10 * * * *"
)

type Scheduler struct {
	logger    *util.Logger
	pluginAPI plugin.API
	user      *User
	workspace *Workspace
	cron      *cron.Cron
}

func NewSchedulerService(logger *util.Logger, plugin plugin.API, workspace *Workspace, user *User) *Scheduler {
	scheduler := &Scheduler{
		logger:    logger,
		pluginAPI: plugin,
		workspace: workspace,
		user:      user,
		cron:      cron.New(),
	}
	return scheduler
}

func (s *Scheduler) InitCronJobs() {
	s.StopCronJobs()
	for userId := range s.workspace.GetUserIds() {
		repository.DeleteUserCronJobIds(s.pluginAPI, userId)
		s.AddCronJobs(userId)
	}
	s.cron.Start()
}

func (s *Scheduler) StopCronJobs() {
	if s.cron == nil {
		return
	}
	s.cron.Stop()
}

func (s *Scheduler) AddCronJobs(userId string) {
	if !s.user.IsUserExist(userId) {
		return
	}
	eventCronId, updateCronId := s.getActiveCronJobIds(userId)

	if eventCronId == nil {
		eventCronEntryId, eventError := s.cron.AddFunc(UserEventHandlerCronSpec, func() {
			s.runHandlerOrDeleteUser(userId, s.user.UserEventsHandler)
		})
		if eventError != nil {
			s.logger.Warn("Error in create Event CRON", &userId)
		} else {
			repository.SaveEventCronJob(s.pluginAPI, userId, int(eventCronEntryId))
		}
	}
	if updateCronId == nil {
		updateCronEntryId, updateError := s.cron.AddFunc(UserEventUpdaterCronSpec, func() {
			s.runHandlerOrDeleteUser(userId, s.user.LoadEventUpdates)
		})
		if updateError != nil {
			s.logger.Warn("Error in create Update CRON", &userId)
		} else {
			repository.SaveUpdateCronJob(s.pluginAPI, userId, int(updateCronEntryId))
		}
	}
}

func (s *Scheduler) DeleteCronJobs(userId string) {
	eventCronId, updateCronId := repository.GetUserCronJobIds(s.pluginAPI, userId)
	if eventCronId != nil {
		s.cron.Remove(cron.EntryID(*eventCronId))
	}
	if updateCronId != nil {
		s.cron.Remove(cron.EntryID(*updateCronId))
	}
}

func (s *Scheduler) runHandlerOrDeleteUser(userId string, handler func(string)) {
	if s.user.IsUserExist(userId) {
		handler(userId)
	} else {
		s.DeleteCronJobs(userId)
		s.workspace.DeleteUser(userId)
	}
}

func (s *Scheduler) getActiveCronJobIds(userId string) (*int, *int) {
	eventCronId, updateCronId := repository.GetUserCronJobIds(s.pluginAPI, userId)
	if eventCronId != nil {
		entry := s.cron.Entry(cron.EntryID(*eventCronId))
		if entry.ID == 0 {
			eventCronId = nil
		}
	}
	if updateCronId != nil {
		entry := s.cron.Entry(cron.EntryID(*updateCronId))
		if entry.ID == 0 {
			updateCronId = nil
		}
	}
	return eventCronId, updateCronId
}
