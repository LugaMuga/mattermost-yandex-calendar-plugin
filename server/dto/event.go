package dto

import (
	"strings"
	"time"
)

const (
	timeFormat = "15:04"
)

type Event struct {
	Id               string
	Name             string
	Description      string
	Url              string
	TimeZone         string
	StartTime        time.Time
	EndTime          time.Time
	LastModifiedTime time.Time
}

type Events []*Event

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
	return dt.Hour() > e.StartTime.Hour() && dt.Minute() > e.StartTime.Minute()
}

func (e *Event) StartBeforeOrEquals(dt time.Time) bool {
	return dt.Hour() >= e.StartTime.Hour() && dt.Minute() >= e.StartTime.Minute()
}

func (e *Event) StartAfter(dt time.Time) bool {
	return dt.Hour() < e.StartTime.Hour() && dt.Minute() < e.StartTime.Minute()
}

func (e *Event) StartEquals(dt time.Time) bool {
	return dt.Hour() == e.StartTime.Hour() && dt.Minute() == e.StartTime.Minute()
}

func (e *Event) EndAfterOrEquals(dt time.Time) bool {
	return dt.Hour() <= e.EndTime.Hour() && dt.Minute() <= e.EndTime.Minute()
}
