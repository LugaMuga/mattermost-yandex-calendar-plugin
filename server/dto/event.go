package dto

import (
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/util"
	"strings"
	"time"
)

const (
	timeFormat = "15:04"
)

type Event struct {
	Id                  string
	Name                string
	Description         string
	Url                 string
	TimeZone            string
	StartTime           time.Time
	EndTime             time.Time
	StartTimeHourMinute int
	EndTimeHourMinute   int
	LastModifiedTime    time.Time
}

func NewEvent(
	Id string,
	Name string,
	Description string,
	Url string,
	TimeZone string,
	StartTime time.Time,
	EndTime time.Time,
	LastModifiedTime time.Time,
) *Event {
	return &Event{
		Id:                  Id,
		Name:                Name,
		Description:         Description,
		Url:                 Url,
		TimeZone:            TimeZone,
		StartTime:           StartTime,
		EndTime:             EndTime,
		LastModifiedTime:    LastModifiedTime,
		StartTimeHourMinute: util.HoursMinutes(StartTime),
		EndTimeHourMinute:   util.HoursMinutes(EndTime),
	}
}

func (e *Event) GetStartTimeFormatted() string {
	return e.StartTime.Format(timeFormat)
}

func (e *Event) GetEndTimeFormatted() string {
	return e.EndTime.Format(timeFormat)
}

func (e *Event) GetDescriptionFormatted() string {
	return strings.Replace(e.Description, "\\n", "\n", -1)
}

func (e *Event) StartBefore(dt time.Time) bool {
	return util.HoursMinutes(dt) > e.StartTimeHourMinute
}

func (e *Event) StartBeforeOrEquals(dt time.Time) bool {
	return util.HoursMinutes(dt) >= e.StartTimeHourMinute
}

func (e *Event) StartAfter(dt time.Time) bool {
	return util.HoursMinutes(dt) < e.StartTimeHourMinute
}

func (e *Event) StartEquals(dt time.Time) bool {
	return dt.Hour() == e.StartTime.Hour() && dt.Minute() == e.StartTime.Minute()
}

func (e *Event) EndAfterOrEquals(dt time.Time) bool {
	return util.HoursMinutes(dt) <= e.EndTimeHourMinute
}

func (e *Event) EndBefore(dt time.Time) bool {
	return util.HoursMinutes(dt) > e.EndTimeHourMinute
}
