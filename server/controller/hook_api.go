package controller

import (
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/conf"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/dto"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/repository"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/service"
	"github.com/mattermost/mattermost-plugin-api/experimental/command"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
	"strings"
	"time"
)

const summaryDateFormat = "01.15.2006"

//CommandHelp - about
const CommandHelp = `###### Mattermost Yandex (CALDav) Calendar Plugin - Slash Command Help
* |/calendar connect [login] [token]| - Connect your Yandex Calendar with your Mattermost account
* |/calendar disconnect| - Disable Yandex Calendar integration
* |/calendar update| - Load updates from Yandex Calendar and show if something added/updated in future
* |/calendar setting| - Change Mattermost Bot settings
* |/calendar summary [date]| - Get a break down of a particular date.
	* |date| should be a date in the format of dd.MM.YYYY or can be "yesterday", "today", "tomorrow" or can be left blank. By default retrieves today's summary breakdown
`

type HookController struct {
	pluginAPI plugin.API
	botId     string
	calendar  *service.Calendar
	user      *service.User
	sender    *service.Sender
	scheduler *service.Scheduler
	workspace *service.Workspace
}

func NewHookController(
	plugin plugin.API,
	botId string,
	calendar *service.Calendar,
	user *service.User,
	sender *service.Sender,
	scheduler *service.Scheduler,
	workspace *service.Workspace) *HookController {
	return &HookController{
		pluginAPI: plugin,
		botId:     botId,
		calendar:  calendar,
		user:      user,
		sender:    sender,
		scheduler: scheduler,
		workspace: workspace,
	}
}

//ExecuteCommand inside plugin
func (hc *HookController) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	split := strings.Fields(args.Command)
	command := split[0]
	action := ""

	if len(split) > 1 {
		action = split[1]
	}

	if command != "/calendar" {
		return &model.CommandResponse{}, nil
	}

	switch action {
	case "connect":
		hc.connect(args)
	case "disconnect":
		hc.disconnect(args)
	case "update":
		hc.update(args)
	case "settings":
		hc.settings(args)
	case "summary":
		hc.summary(args)
	case "help":
		hc.help(args)
	}
	return &model.CommandResponse{}, nil
}

func GetHookCommand(pluginAPI plugin.API) (*model.Command, error) {
	iconData, err := command.GetIconData(pluginAPI, "assets/icon.svg")

	if err != nil {
		return nil, errors.Wrap(err, "failed to get icon data")
	}

	return &model.Command{
		Trigger:              "calendar",
		DisplayName:          "Google Calendar",
		Description:          "Integration with Google Calendar",
		AutoComplete:         true,
		AutoCompleteDesc:     "Available commands: connect, list, summary, create, help",
		AutoCompleteHint:     "[command]",
		AutocompleteData:     getAutocompleteData(),
		AutocompleteIconData: iconData,
	}, nil
}

func getAutocompleteData() *model.AutocompleteData {
	cal := model.NewAutocompleteData("calendar", "[command]", "Available commands: connect, list, summary, create, help")

	connect := model.NewAutocompleteData("connect", "[login] [token]", "Connect your Yandex Calendar with your login and token")
	cal.AddCommand(connect)

	disconnect := model.NewAutocompleteData("disconnect", "", "Disable Yandex Calendar integration")
	cal.AddCommand(disconnect)

	update := model.NewAutocompleteData("update", "", "Update events from server")
	cal.AddCommand(update)

	settings := model.NewAutocompleteData("settings", "", "Change calendar and notification settings for bot")
	cal.AddCommand(settings)

	summary := model.NewAutocompleteData("summary", "[date]", "Get a breakdown of a particular date")
	summary.AddTextArgument("The date to view in dd.MM.YYYY format", "[date]", "")
	cal.AddCommand(summary)

	help := model.NewAutocompleteData("help", "", "Display usage")
	cal.AddCommand(help)
	return cal
}

func (hc *HookController) connect(args *model.CommandArgs) {
	split := strings.Fields(args.Command)
	if len(split) < 4 {
		hc.sender.SendBotDMPost(args.UserId, "Wrong command format. Please read **[instruction](https://github.com/LugaMuga/mattermost-yandex-calendar-plugin/docs/readme.md)**")
		return
	}
	credentials := &dto.Credentials{
		Login: split[2],
		Token: split[3],
	}
	hc.user.Connect(args.UserId, args.TriggerId, args.RootId, *credentials)
}

func (hc *HookController) settings(args *model.CommandArgs) {
	hc.user.Settings(args.UserId, args.TriggerId, args.RootId)
}

func (hc *HookController) disconnect(args *model.CommandArgs) {
	userId := args.UserId
	hc.scheduler.DeleteCronJobs(userId)
	hc.workspace.DeleteUser(userId)
	hc.sender.SendBotDMPost(userId, "Bye, bye :wave:")
}

func (hc *HookController) update(args *model.CommandArgs) {
	userId := args.UserId
	hc.user.LoadEventUpdates(userId)
}

func (hc *HookController) summary(args *model.CommandArgs) {
	split := strings.Fields(args.Command)
	userId := args.UserId
	userSettings := repository.GetSettings(hc.pluginAPI, userId)
	if userSettings == nil || userSettings.TimeZone == "" {
		return
	}
	day := "today"
	if len(split) >= 3 {
		day = split[2]
	}

	userNow := time.Now().In(userSettings.GetUserLocation())
	var start, end time.Time
	var title string

	switch day {
	case "yesterday":
		yesterday := userNow.AddDate(0, 0, -1)
		start = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())
		end = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 0, yesterday.Location())
		title = conf.GetEventsTitle(conf.YesterdayEventsTitle, start)
	case "today":
		start, end = hc.calendar.GetTodayDateTimes(userId)
		title = conf.GetEventsTitle(conf.TodayEventsTitle, start)
	case "tomorrow":
		tomorrow := userNow.AddDate(0, 0, 1)
		start = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, tomorrow.Location())
		end = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 23, 59, 59, 0, tomorrow.Location())
		title = conf.GetEventsTitle(conf.TomorrowEventsTitle, start)
	default:
		date, err := time.ParseInLocation(summaryDateFormat, split[2], userSettings.GetUserLocation())
		if err != nil {
			hc.sender.SendBotDMPost(userId, "Can't parse date. Please use format dd.MM.YYYY")
		}
		start = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		end = time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, date.Location())
		title = conf.GetEventsTitle("", start)
	}

	events, err := hc.calendar.LoadEvents(userId, start, end)
	if err != nil {
		hc.sender.SendBotDMPost(userId, "Catch error in load events")
	}

	hc.calendar.SortEvents(events)
	hc.sender.SendEvents(userId, title, events)
}

func (hc *HookController) help(args *model.CommandArgs) {
	hc.sender.SendBotDMPost(args.UserId, strings.Replace(CommandHelp, "|", "`", -1))
}
