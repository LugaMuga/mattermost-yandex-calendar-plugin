package service

import (
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/conf"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/dto"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/repository"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/util"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"time"
)

type User struct {
	logger                    *util.Logger
	pluginAPI                 plugin.API
	supportedUserCustomStatus bool
	credentialsRepo           *repository.CredentialsRepo
	sender                    *Sender
	calendar                  *Calendar
}

func NewUserService(
	logger *util.Logger,
	plugin plugin.API,
	supportedUserCustomStatus bool,
	credentialsRepo *repository.CredentialsRepo,
	sender *Sender,
	calendar *Calendar) *User {
	return &User{
		logger:                    logger,
		pluginAPI:                 plugin,
		supportedUserCustomStatus: supportedUserCustomStatus,
		credentialsRepo:           credentialsRepo,
		sender:                    sender,
		calendar:                  calendar,
	}
}

func (u *User) Connect(userId string, triggerId string, rootId string, credentials dto.Credentials) {
	u.credentialsRepo.SaveCredentials(userId, credentials)
	calendarHomeSet, err := u.calendar.GetCalendarHomeSet(userId)
	if err != nil {
		u.sender.SendBotDMPost(userId, err.Error())
		return
	}
	repository.SaveCalendarHomeSet(u.pluginAPI, userId, calendarHomeSet)
	u.sender.SendWelcomePost(userId)
	u.Settings(userId, triggerId, rootId)
}

func (u *User) Settings(userId string, triggerId string, rootId string) {
	calendars, _ := u.calendar.FindCalendars(userId)
	settings := repository.GetSettings(u.pluginAPI, userId)
	if settings == nil {
		settings = dto.DefaultSettings()
	}
	err := u.sender.OpenSettingsDialog(triggerId, rootId, calendars, settings)
	if err != nil {
		u.logger.LogError("Couldn't open settings dialog", &userId, err)
	}
}

func (u *User) UserEventsHandler(userId string) {
	userSettings := repository.GetSettings(u.pluginAPI, userId)
	events := repository.GetEvents(u.pluginAPI, userId)
	u.remindUser(userId, userSettings.GetUserNow(), userSettings, events)
	u.updateUserEventStatus(userId, userSettings.GetUserNow(), userSettings, events)
}

func (u *User) remindUser(userId string, userNow time.Time, userSettings *dto.Settings, events []dto.Event) {
	if userSettings.DailyNotifyTime != nil &&
		userNow.Hour() == userSettings.DailyNotifyTime.Hour() &&
		userNow.Minute() == userSettings.DailyNotifyTime.Minute() {
		u.sender.SendEvents(userId, conf.GetTodayEventsTitle(userNow), events)
	}
	tenMinutesLater := userNow.Add(10 * time.Minute)
	for _, event := range events {
		if event.StartBefore(userNow) || event.StartAfter(tenMinutesLater) {
			continue
		}
		//TODO check attendees
		if userSettings.TenMinutesNotify && event.StartEquals(tenMinutesLater) {
			u.sender.SendEvent(userId, conf.TenMinutesEventTitle, event)
		}
		oneMinuteLater := userNow.Add(1 * time.Minute)
		if userSettings.OneMinutesNotify && event.StartEquals(oneMinuteLater) {
			u.sender.SendEvent(userId, conf.OneMinuteEventTitle, event)
		}
	}
}

func (u *User) updateUserEventStatus(userId string, userNow time.Time, userSettings *dto.Settings, events []dto.Event) {
	if !u.supportedUserCustomStatus || !userSettings.ChangeStatusOnMeet || len(events) == 0 {
		return
	}
	userState := repository.GetState(u.pluginAPI, userId)
	if userState != nil && userState.CurrentEvent != nil &&
		userState.CurrentEvent.StartBeforeOrEquals(userNow) &&
		userState.CurrentEvent.EndAfterOrEquals(userNow) {
		return
	}
	var currentEvent *dto.Event
	for _, event := range events {
		if event.StartBeforeOrEquals(userNow) && event.EndAfterOrEquals(userNow) {
			currentEvent = &event
			break
		}
	}
	if currentEvent != nil {
		end := currentEvent.EndTime
		err := u.pluginAPI.UpdateUserCustomStatus(userId, &model.CustomStatus{
			Emoji:     "calendar",
			Text:      "In meeting",
			Duration:  "date_and_time",
			ExpiresAt: time.Date(userNow.Year(), userNow.Month(), userNow.Day(), end.Hour(), end.Minute(), 0, 0, userNow.Location()),
		})
		if err != nil {
			u.logger.LogWarn("Error in update custom status", &userId, err)
		}
	}
	userState = &dto.State{
		CurrentEvent: currentEvent,
	}
	repository.SaveState(u.pluginAPI, userId, *userState)
}

func (u *User) LoadEventUpdates(userId string) {
	addedEvents, updatedEvents := u.calendar.LoadCalendarUpdates(userId)
	if addedEvents != nil {
		u.sender.SendEvents(userId, conf.AddedEventsTitle, addedEvents)
	}
	if updatedEvents != nil {
		u.sender.SendEvents(userId, conf.UpdatedEventsTitle, updatedEvents)
	}
}

func (u *User) IsUserExist(userId string) bool {
	status, err := u.pluginAPI.GetUserStatus(userId)
	if err == nil && status == nil {
		return false
	}
	return true
}
