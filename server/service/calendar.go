package service

import (
	"fmt"
	"github.com/lugamuga/go-webdav"
	"github.com/lugamuga/go-webdav/caldav"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/convertor"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/dto"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/repository"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/util"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/pkg/errors"
	"net/http"
	"sort"
	"time"
)

type Calendar struct {
	logger          *util.Logger
	pluginAPI       plugin.API
	serverUrl       string
	credentialsRepo *repository.CredentialsRepo
}

func NewCalendarService(
	logger *util.Logger,
	plugin plugin.API,
	serverUrl string,
	credentialsRepo *repository.CredentialsRepo) *Calendar {
	return &Calendar{
		logger:          logger,
		pluginAPI:       plugin,
		serverUrl:       serverUrl,
		credentialsRepo: credentialsRepo,
	}
}

func (c *Calendar) getClient(userId string) (*caldav.Client, error) {
	credentials := c.credentialsRepo.GetCredentials(userId)
	if credentials == nil {
		return nil, errors.New("Could not found credentials")
	}
	httpClient := webdav.HTTPClientWithBasicAuth(&http.Client{}, credentials.Login, credentials.Token)
	client, err := caldav.NewClient(httpClient, c.serverUrl)
	return client, err
}

func (c *Calendar) GetCalendarHomeSet(userId string) (string, error) {
	client, err := c.getClient(userId)
	if err != nil {
		c.logger.LogError("Error get client for principal", &userId, err)
		return "", errors.New(fmt.Sprintf("Error get client in principal method for user %s", userId))
	}
	principal, err := client.FindCurrentUserPrincipal()
	if err != nil {
		c.logger.LogError("Error get principal", &userId, err)
		return "", errors.New(fmt.Sprintf("Error get principal for user %s", userId))
	}
	return client.FindCalendarHomeSet(principal)
}

func (c *Calendar) FindCalendars(userId string) ([]caldav.Calendar, error) {
	calendarHomeSet := repository.GetCalendarHomeSet(c.pluginAPI, userId)
	client, err := c.getClient(userId)
	if err != nil {
		c.logger.LogError("Error get calendars", &userId, err)
		return make([]caldav.Calendar, 0), errors.New(fmt.Sprintf("Error get calendars for user %s", userId))
	}
	return client.FindCalendars(calendarHomeSet)
}

func (c *Calendar) LoadCalendar(userId string) ([]dto.Event, error) {
	events, _ := c.loadTodayEvents(userId)
	c.SortEvents(events)
	repository.SaveEvents(c.pluginAPI, userId, events)
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
	existingEvents := repository.GetEvents(c.pluginAPI, userId)
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
	repository.SaveEvents(c.pluginAPI, userId, events)
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
	client, err := c.getClient(userId)
	if err != nil {
		c.logger.LogError("Can't get client for calendar "+userSettings.Calendar, &userId, err)
		return events, errors.New("Can't get client for calendar")
	}
	calendarObjects, err := c.queryCalendarEventsByTimeRange(client, userSettings.Calendar, start, end)
	if calendarObjects == nil {
		c.logger.LogError("Can't get events for calendar "+userSettings.Calendar, &userId, err)
		return events, errors.New("Can't get events from calendar")
	}
	timezone, err := convertor.GetTimezone(calendarObjects)
	if err != nil {
		c.logger.LogWarn("Can't get timezone for calendar "+userSettings.Calendar, &userId, err)
	}
	eventDtos, err := convertor.CalendarObjectToEventArray(calendarObjects, timezone)
	if err != nil {
		c.logger.LogWarn("Can't parse events for calendar "+userSettings.Calendar, &userId, err)
		return events, errors.New("Can't parse events from calendar")
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
