package service

import (
	"encoding/json"
	"fmt"
	"github.com/lugamuga/go-webdav"
	"github.com/lugamuga/go-webdav/caldav"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/conf"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/convertor"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/dto"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/repository"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/mattermost/mattermost-server/v5/shared/mlog"
	"github.com/pkg/errors"
	"net/http"
	"sort"
	"time"
)

const CalendarEndpoint = "https://caldav.yandex.ru"

type Calendar struct {
	pluginAPI plugin.API
}

func NewCalendarService(plugin plugin.API) *Calendar {
	return &Calendar{pluginAPI: plugin}
}

func (c *Calendar) getClient(userId string) (*caldav.Client, error) {
	credentials := repository.GetCredentials(c.pluginAPI, userId)
	if credentials == nil {
		return nil, errors.New("Could not found credentials")
	}
	httpClient := webdav.HTTPClientWithBasicAuth(&http.Client{}, credentials.Login, credentials.Token)
	client, err := caldav.NewClient(httpClient, CalendarEndpoint)
	return client, err
}

func (c *Calendar) GetCalendarHomeSet(userId string) (string, error) {
	client, _ := c.getClient(userId)
	principal, err := client.FindCurrentUserPrincipal()
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error get principal for user %s", userId))
	}
	return client.FindCalendarHomeSet(principal)
}

func (c *Calendar) FindCalendars(userId string) ([]caldav.Calendar, error) {
	calendarHomeSet := repository.GetCalendarHomeSet(c.pluginAPI, userId)
	client, _ := c.getClient(userId)
	return client.FindCalendars(calendarHomeSet)
}

func (c *Calendar) LoadCalendar(userId string) ([]dto.Event, error) {
	events, _ := c.loadTodayEvents(userId)
	c.SortEvents(events)
	eventsJson, _ := json.Marshal(events)
	c.pluginAPI.KVSet(userId+conf.Events, eventsJson)
	repository.SaveLastUpdate(c.pluginAPI, userId, getNowForLastUpdated())
	return events, nil
}

func (c *Calendar) LoadCalendarUpdates(userId string) ([]dto.Event, []dto.Event) {
	now := getNowForLastUpdated()
	var lastUpdate = repository.GetUserCalendarLastUpdate(c.pluginAPI, userId)
	if lastUpdate == nil {
		lastUpdate = &now
	}
	var events []dto.Event
	var updatedEvents []dto.Event
	var addedEvents []dto.Event
	loadedEvents, _ := c.loadTodayEvents(userId)
	existingEventsBytes, _ := c.pluginAPI.KVGet(userId + conf.Events)
	var existingEvents []*dto.Event
	err := json.Unmarshal(existingEventsBytes, &existingEvents)
	if err != nil {
		mlog.Warn("error on parse events from storage", mlog.Err(err))
	}
	existingEventById := convertor.SliceEventToMapById(existingEvents)
	for _, event := range loadedEvents {
		events = append(events, event)
		if event.LastModifiedTime.After(*lastUpdate) && event.StartAfter(now) {
			if _, ok := existingEventById[event.Id]; ok {
				updatedEvents = append(updatedEvents, event)
			} else {
				addedEvents = append(addedEvents, event)
			}
		}
	}
	c.SortEvents(events)
	eventsJson, _ := json.Marshal(events)
	c.pluginAPI.KVSet(userId+conf.Events, eventsJson)
	repository.SaveLastUpdate(c.pluginAPI, userId, now)
	return addedEvents, updatedEvents
}

func (c *Calendar) loadTodayEvents(userId string) ([]dto.Event, error) {
	start, end := c.GetTodayDateTimes(userId)
	return c.LoadEvents(userId, start, end)
}

func (c *Calendar) LoadEvents(userId string, start time.Time, end time.Time) ([]dto.Event, error) {
	var events []dto.Event
	userSettings := repository.GetSettings(c.pluginAPI, userId)
	client, _ := c.getClient(userId)

	calendarObjects, err := c.queryCalendarEventsByTimeRange(client, userSettings.Calendar, start, end)
	if calendarObjects == nil {
		c.pluginAPI.LogError("Can't get events for calendar "+userSettings.Calendar, "err", err)
	}
	timezone, err := convertor.GetTimezone(calendarObjects)
	if err != nil {
		c.pluginAPI.LogWarn("Can't get timezone for calendar "+userSettings.Calendar, "err", err)
	}
	eventDtos, err := convertor.CalendarObjectToEventArray(calendarObjects, timezone)
	if err != nil {
		c.pluginAPI.LogWarn("Can't parse events for calendar "+userSettings.Calendar, "err", err)
	}
	events = append(events, eventDtos...)
	return events, nil
}

func (c *Calendar) queryCalendarEventsByTimeRange(
	client *caldav.Client,
	calendarPath string,
	start time.Time,
	end time.Time) ([]caldav.CalendarObject, error) {

	query := &caldav.CalendarQuery{
		CompFilter: caldav.CompFilter{
			Name: "VCALENDAR",
			Comps: []caldav.CompFilter{{
				Name:  "VEVENT",
				Start: start.UTC(),
				End:   end.UTC(),
			}},
		},
	}
	return client.QueryCalendar(calendarPath, query)
}

func (c *Calendar) GetTodayDateTimes(userId string) (time.Time, time.Time) {
	userSettings := repository.GetSettings(c.pluginAPI, userId)
	now := time.Now().In(userSettings.GetUserLocation())
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	return start, end
}

func (c *Calendar) SortEvents(events []dto.Event) {
	sort.SliceStable(events, func(i, j int) bool {
		if events[i].StartTime.Hour() != events[j].StartTime.Hour() {
			return events[i].StartTime.Hour() < events[j].StartTime.Hour()
		}
		if events[i].StartTime.Minute() != events[j].StartTime.Minute() {
			return events[i].StartTime.Minute() < events[j].StartTime.Minute()
		}
		return events[i].Name < events[j].Name
	})
}

func getNowForLastUpdated() time.Time {
	return time.Now().UTC()
}
