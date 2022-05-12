package main

import (
	calendarPlugin "github.com/lugamuga/mattermost-yandex-calendar-plugin/server/plugin"
	mmPlugin "github.com/mattermost/mattermost-server/v6/plugin"
)

func main() {
	mmPlugin.ClientMain(&calendarPlugin.Plugin{})
}
