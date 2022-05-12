package service

import (
	"github.com/lugamuga/go-webdav/caldav"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/conf"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/dto"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/mattermost/mattermost-server/v6/shared/mlog"
	"github.com/tkuchiki/go-timezone"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Sender struct {
	manifestId             string
	botId                  string
	pluginAPI              plugin.API
	serverConfig           *model.Config
	timezoneOptions        []*model.PostActionOptions
	dailyNotifyTimeOptions []*model.PostActionOptions
}

func NewSenderService(manifestId string, botId string, plugin plugin.API, serverConfig *model.Config) *Sender {
	return &Sender{
		manifestId:             manifestId,
		botId:                  botId,
		pluginAPI:              plugin,
		serverConfig:           serverConfig,
		timezoneOptions:        prepareTimezoneOptions(),
		dailyNotifyTimeOptions: prepareDailyNotifyTimeOptions(),
	}
}

func (s *Sender) SendBotDMPost(userId string, message string) *model.AppError {
	channel, err := s.pluginAPI.GetDirectChannel(userId, s.botId)
	if err != nil {
		//mlog.Error("Couldn't get bot's DM channel", mlog.String("user_id", userID))
		return err
	}

	post := &model.Post{
		UserId:    s.botId,
		ChannelId: channel.Id,
		Message:   message,
	}
	return s.sendPost(post)
}

func (s *Sender) SendWelcomePost(userId string) {
	message := "#### Welcome to the Mattermost Yandex Calendar Plugin!\n" +
		"Please type **/calendar help** to understand how to use this plugin. "
	s.SendBotDMPost(userId, message)
}

func (s *Sender) OpenSettingsDialog(triggerId string, rootId string, calendars []caldav.Calendar, settings *dto.Settings) error {
	siteURL := *s.serverConfig.ServiceSettings.SiteURL
	dialog := model.OpenDialogRequest{
		TriggerId: triggerId,
		URL:       conf.ResolveUrlByPlugin(strings.ToLower(s.manifestId), conf.CalendarSettings),
		Dialog:    s.getSettingsDialog(siteURL, rootId, calendars, settings),
	}

	if appErr := s.pluginAPI.OpenInteractiveDialog(dialog); appErr != nil {
		s.pluginAPI.LogWarn("failed to open create poll dialog", "err", appErr.Error())
		return appErr
	}
	return nil
}

func (s *Sender) getSettingsDialog(siteURL string, rootId string, calendars []caldav.Calendar, settings *dto.Settings) model.Dialog {
	var calendarOptions []*model.PostActionOptions
	for _, c := range calendars {
		calendarOptions = append(calendarOptions, &model.PostActionOptions{
			Text:  c.Name,
			Value: c.Path,
		})
	}
	var dialogElements []model.DialogElement
	dialogElements = append(dialogElements, model.DialogElement{
		Name:        conf.SelectCalendarDialogOption,
		DisplayName: "Select calendar for notifications",
		Type:        "select",
		Optional:    false,
		Default:     settings.Calendar,
		Options:     calendarOptions,
	})

	dialogElements = append(dialogElements, model.DialogElement{
		Name:        conf.SelectTimezoneDialogOption,
		DisplayName: "Select your timezone",
		Type:        "select",
		Optional:    false,
		Default:     settings.TimeZone,
		Options:     s.timezoneOptions,
	})

	defaultDailyNotifyTime := conf.DailyNotifyTimeDisableOption
	if settings.DailyNotifyTime != nil {
		defaultDailyNotifyTime = settings.DailyNotifyTime.Format(time.RFC3339)
	}
	dialogElements = append(dialogElements, model.DialogElement{
		Name:        conf.DailyNotifyTimeDialogOption,
		DisplayName: "Select time for get daily schedule",
		Type:        "select",
		Optional:    false,
		Default:     defaultDailyNotifyTime,
		Options:     s.dailyNotifyTimeOptions,
	})

	dialogElements = append(dialogElements, model.DialogElement{
		Name:        conf.ChangeStatusOnMeetDialogOption,
		DisplayName: "Setup 'In meeting' status automatically",
		Type:        "bool",
		Default:     strconv.FormatBool(settings.ChangeStatusOnMeet),
		Optional:    true,
	})

	dialogElements = append(dialogElements, model.DialogElement{
		Name:        conf.TenMinuteNotifyDialogOption,
		DisplayName: "Get notification in 10 minutes before event",
		Type:        "bool",
		Default:     strconv.FormatBool(settings.TenMinutesNotify),
		Optional:    true,
	})

	dialogElements = append(dialogElements, model.DialogElement{
		Name:        conf.OneMinuteNotifyDialogOption,
		DisplayName: "Get notification in 1 minute before event",
		Type:        "bool",
		Default:     strconv.FormatBool(settings.OneMinutesNotify),
		Optional:    true,
	})

	dialog := model.Dialog{
		CallbackId:  rootId,
		Title:       "Settings",
		IconURL:     conf.GetIconUrl(siteURL, s.manifestId),
		SubmitLabel: "Save",
		Elements:    dialogElements,
	}
	return dialog
}

func prepareTimezoneOptions() []*model.PostActionOptions {
	var timezoneOptions []*model.PostActionOptions
	for name, tzinfo := range timezone.New().TzInfos() {
		timezoneOptions = append(timezoneOptions, &model.PostActionOptions{
			Text:  name + " " + tzinfo.StandardOffsetHHMM(),
			Value: name,
		})
	}
	sort.SliceStable(timezoneOptions, func(i, j int) bool {
		return timezoneOptions[i].Value < timezoneOptions[j].Value
	})
	return timezoneOptions
}

func prepareDailyNotifyTimeOptions() []*model.PostActionOptions {
	minutes := []int{0, 15, 30, 45}
	var options []*model.PostActionOptions
	options = append(options, &model.PostActionOptions{
		Text:  conf.DailyNotifyTimeDisableOption,
		Value: conf.DailyNotifyTimeDisableOption,
	})
	for h := 0; h < 24; h++ {
		for _, m := range minutes {
			if h == 0 && m == 0 {
				continue
			}
			hour := strconv.Itoa(h)
			if h < 10 {
				hour = "0" + strconv.Itoa(h)
			}
			min := strconv.Itoa(m)
			if m == 0 {
				min = "00"
			}
			options = append(options, &model.PostActionOptions{
				Text:  hour + ":" + min,
				Value: time.Date(1, 1, 1, h, m, 0, 0, time.UTC).Format(time.RFC3339),
			})
		}
	}
	return options
}

func (s *Sender) SendEvent(userId string, title string, event dto.Event) *model.AppError {
	var attachments []*model.SlackAttachment
	attachments = append(attachments, s.getFormattedEventAttachment(event))
	return s.sendEvents(userId, title, attachments)
}

func (s *Sender) SendEvents(userId string, title string, events []dto.Event) *model.AppError {
	var attachments []*model.SlackAttachment
	for _, event := range events {
		attachments = append(attachments, s.getFormattedEventAttachment(event))
	}
	return s.sendEvents(userId, title, attachments)
}

func (s *Sender) sendEvents(userId string, title string, attachments []*model.SlackAttachment) *model.AppError {
	channel, err := s.pluginAPI.GetDirectChannel(userId, s.botId)
	if err != nil {
		mlog.Error("Couldn't get bot's DM channel", mlog.String("user_id", userId))
	}
	var post *model.Post
	if attachments == nil || len(attachments) == 0 {
		post = &model.Post{
			UserId:    s.botId,
			ChannelId: channel.Id,
			Type:      model.PostTypeDefault,
			Message:   title + "\nNo events",
		}
	} else {
		post = &model.Post{
			UserId:    s.botId,
			ChannelId: channel.Id,
			Type:      model.PostTypeSlackAttachment,
			Message:   title,
		}
		post.AddProp("attachments", attachments)
	}
	return s.sendPost(post)
}

func (s *Sender) getFormattedEventAttachment(event dto.Event) *model.SlackAttachment {
	title := event.GetStartTimeFormatted() + " - " + event.GetEndTimeFormatted()
	title += " [" + event.Name + "](" + event.Url + ")"
	return &model.SlackAttachment{
		Color: "blue",
		Title: title,
		Text:  event.GetDescriptionFormatted(),
	}
}

func (s *Sender) sendPost(post *model.Post) *model.AppError {
	if _, err := s.pluginAPI.CreatePost(post); err != nil {
		mlog.Error(err.Error())
		return err
	}
	return nil
}
