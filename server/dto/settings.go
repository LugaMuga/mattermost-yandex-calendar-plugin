package dto

import (
	"time"
)

type Settings struct {
	TenMinutesNotify bool
	OneMinutesNotify bool
	Calendar         string
	TimeZone         string
	DailyNotifyTime  *time.Time
}

func DefaultSettings() *Settings {
	defaultDailyNotifyTime := time.Date(1, 1, 1, 7, 0, 0, 0, time.UTC)
	return &Settings{
		TenMinutesNotify: true,
		OneMinutesNotify: true,
		Calendar:         "",
		TimeZone:         "",
		DailyNotifyTime:  &defaultDailyNotifyTime,
	}
}

func (s *Settings) GetUserLocation() *time.Location {
	location, _ := time.LoadLocation(s.TimeZone)
	return location
}

func (s *Settings) GetUserNow() time.Time {
	location, _ := time.LoadLocation(s.TimeZone)
	return time.Now().In(location).Truncate(time.Minute)
}
