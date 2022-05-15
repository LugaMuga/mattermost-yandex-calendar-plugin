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
	scheduler.cron.Start()
	return scheduler
}

func (s *Scheduler) InitCronJobs() {
	for userId := range s.workspace.GetUserIds() {
		s.AddCronJobs(userId)
	}
}

func (s *Scheduler) StopCronJobs() {
	if s.cron == nil {
		return
	}
	s.cron.Stop()
}

func (s *Scheduler) AddCronJobs(userId string) {
	eventCronId, updateCronId := s.getActiveCronJobIds(userId)

	if eventCronId == nil {
		eventCronEntryId, eventError := s.cron.AddFunc(UserEventHandlerCronSpec, func() {
			if _, err := s.pluginAPI.GetUser(userId); err != nil {
				s.DeleteCronJobs(userId)
				s.workspace.DeleteUser(userId)
				return
			}
			s.user.UserEventsHandler(userId)
		})
		if eventError != nil {
			s.logger.Warn("Error in create Event CRON", &userId)
		} else {
			repository.SaveEventCronJob(s.pluginAPI, userId, int(eventCronEntryId))
		}
	}
	if updateCronId == nil {
		updateCronEntryId, updateError := s.cron.AddFunc(UserEventUpdaterCronSpec, func() {
			s.user.LoadEventUpdates(userId)
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
