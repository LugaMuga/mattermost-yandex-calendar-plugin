package conf

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	IconImgName = "icon.png"
)

func GetIconUrl(siteUrl string, manifestId string) string {
	return fmt.Sprintf("%s/plugins/%s/%s", siteUrl, strings.ToLower(manifestId), IconImgName)
}

const (
	ApiV1Prefix      = "/api/v1"
	CalendarSettings = "/calendar/settings"
)

func ResolveUrlByPlugin(manifestId string, path string) string {
	return fmt.Sprintf("/plugins/%s%s%s", strings.ToLower(manifestId), ApiV1Prefix, path)
}

const (
	SelectCalendarDialogOption     = "calendar"
	SelectTimezoneDialogOption     = "timezone"
	DailyNotifyTimeDialogOption    = "dailyNotifyTime"
	TenMinuteNotifyDialogOption    = "tenMinutesNotify"
	OneMinuteNotifyDialogOption    = "oneMinuteNotify"
	ChangeStatusOnMeetDialogOption = "changeStatusOnMeet"
)

const (
	DailyNotifyTimeDisableOption = "Never"
)

const (
	YesterdayEventsTitle = "##### :calendar: Yesterday"
	TomorrowEventsTitle  = "##### :calendar: Tomorrow"
	TodayEventsTitle     = "##### :calendar: Today"
	AddedEventsTitle     = "##### :new: Added events"
	UpdatedEventsTitle   = "##### :arrows_counterclockwise: Updated events"
	TenMinutesEventTitle = "##### :clock10: 10 minutes until event"
	OneMinuteEventTitle  = "##### :alarm_clock: 1 minute until event"
)

func GetTodayEventsTitle(dt time.Time) string {
	return GetEventsTitle(TodayEventsTitle, dt)
}

func GetEventsTitle(name string, dt time.Time) string {
	part := name + " - " + dt.Weekday().String()
	if name == "" {
		part = "##### :calendar: " + dt.Weekday().String()
	}
	return fmt.Sprintf("%s, %s %s", part, dt.Month(), strconv.Itoa(dt.Day()))
}
