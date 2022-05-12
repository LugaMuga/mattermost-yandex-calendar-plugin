package service

import (
	"encoding/json"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/conf"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/dto"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/repository"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/mattermost/mattermost-server/v5/shared/mlog"
	"time"
)

type User struct {
	pluginAPI plugin.API
	sender    *Sender
	calendar  *Calendar
}

func NewUserService(plugin plugin.API, sender *Sender, calendar *Calendar) *User {
	return &User{
		pluginAPI: plugin,
		sender:    sender,
		calendar:  calendar,
	}
}

func (u *User) Connect(userId string, triggerId string, rootId string, credentials dto.Credentials) {
	repository.SaveCredentials(u.pluginAPI, userId, credentials)
	calendarHomeSet, err := u.calendar.GetCalendarHomeSet(userId)
	if err != nil {
		u.sender.SendBotDMPost(userId, err.Error())
		return
	}
	u.pluginAPI.KVSet(userId+conf.HomeSet, []byte(calendarHomeSet))
	u.sender.SendWelcomePost(userId)
	u.Settings(userId, triggerId, rootId)
}

func (u *User) Settings(userId string, triggerId string, rootId string) {
	calendars, _ := u.calendar.FindCalendars(userId)
	settings := repository.GetSettings(u.pluginAPI, userId)
	if settings == nil {
		settings = dto.DefaultSettings()
	}
	u.sender.OpenSettingsDialog(triggerId, rootId, calendars, settings)
}

func (u *User) UserEventsHandler(userId string) {
	eventsByte, _ := u.pluginAPI.KVGet(userId + conf.Events)
	userSettings := repository.GetSettings(u.pluginAPI, userId)

	var events []dto.Event
	err := json.Unmarshal(eventsByte, &events)
	if err != nil {
		mlog.Warn("error on parse events from storage")
	}
	u.remindUser(userId, userSettings.GetUserNow(), userSettings, events)
	// TODO uncomment, when will be possible to change user custom status
	//u.updateUserEventStatus(userId, userSettings.GetUserNow(), userSettings, events)
}

func (u *User) remindUser(userId string, userNow time.Time, userSettings *dto.Settings, events []dto.Event) {
	if userSettings.DailyNotifyTime != nil &&
		userNow.Hour() == userSettings.DailyNotifyTime.Hour() &&
		userNow.Minute() == userSettings.DailyNotifyTime.Minute() {
		u.sender.SendEvents(userId, conf.GetTodayEventsTitle(userNow), events)
	}

	for _, event := range events {
		if event.StartAfter(userNow) {
			continue
		}
		//TODO check attendees
		tenMinutesLater := userNow.Add(10 * time.Minute)
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
	if len(events) == 0 {
		return
	}
	userState := repository.GetState(u.pluginAPI, userId)
	if userState == nil {
		userState = dto.DefaultState()
	}
	userInEvent := false
	for _, event := range events {
		if event.StartBeforeOrEquals(userNow) && event.EndAfterOrEquals(userNow) {
			userInEvent = true
			break
		}
	}
	if !userState.InEvent && userInEvent {
		userStatus, _ := u.pluginAPI.GetUserStatus(userId)
		userState.SavedStatus = userStatus.Status
		// Can't find way to change user custom status
		u.pluginAPI.UpdateUserStatus(userId, userState.SavedStatus)
	} else if userState.InEvent && !userInEvent {
		u.pluginAPI.UpdateUserStatus(userId, userState.SavedStatus)
		userState.SavedStatus = ""
	}
	userState.InEvent = userInEvent
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
