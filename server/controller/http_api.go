package controller

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/conf"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/dto"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/repository"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/service"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"io"
	"net/http"
	"path/filepath"
	"time"
)

type HttpController struct {
	pluginAPI     plugin.API
	pluginVersion string
	calendar      *service.Calendar
	user          *service.User
	sender        *service.Sender
	scheduler     *service.Scheduler
	workspace     *service.Workspace
	router        *mux.Router
}

func NewHttpController(
	plugin plugin.API,
	pluginVersion string,
	calendar *service.Calendar,
	user *service.User,
	sender *service.Sender,
	scheduler *service.Scheduler,
	workspace *service.Workspace) *HttpController {
	httpController := &HttpController{
		pluginAPI:     plugin,
		pluginVersion: pluginVersion,
		calendar:      calendar,
		user:          user,
		sender:        sender,
		scheduler:     scheduler,
		workspace:     workspace,
	}
	httpController.router = httpController.newRouter()
	return httpController
}

func (hc *HttpController) newRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/", hc.handleInfo).Methods(http.MethodGet)
	router.HandleFunc("/"+conf.IconImgName, hc.handleIcon).Methods(http.MethodGet)

	apiV1 := router.PathPrefix(conf.ApiV1Prefix).Subrouter()
	apiV1.Use(checkAuthenticity)

	apiV1.HandleFunc(conf.CalendarSettings, hc.handleSetupRequest()).Methods(http.MethodPost)
	return router
}

func (hc *HttpController) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	hc.pluginAPI.LogDebug("New request:", "Host", r.Host, "RequestURI", r.RequestURI, "Method", r.Method)
	hc.router.ServeHTTP(w, r)
}

func checkAuthenticity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Mattermost-User-ID") == "" {
			http.Error(w, "not authorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (hc *HttpController) handleInfo(w http.ResponseWriter, _ *http.Request) {
	_, _ = io.WriteString(w, "Thanks for using Yandex calendar plugin v"+hc.pluginVersion+"\n")
}

func (hc *HttpController) handleSetupRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request := submitDialogRequestFromJson(r.Body)
		if request == nil || request.Submission == nil {
			hc.pluginAPI.LogWarn("Failed to decode DialogSubmission")
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		userId := request.UserId
		if userId != r.Header.Get("Mattermost-User-ID") {
			http.Error(w, "not authorized", http.StatusUnauthorized)
			return
		}

		settings := &dto.Settings{}
		for selector, value := range request.Submission {
			switch selector {
			case conf.SelectCalendarDialogOption:
				settings.Calendar = value.(string)
			case conf.SelectTimezoneDialogOption:
				settings.TimeZone = value.(string)
			case conf.ChangeStatusOnMeetDialogOption:
				settings.ChangeStatusOnMeet = value.(bool)
			case conf.DailyNotifyTimeDialogOption:
				val := value.(string)
				if val == conf.DailyNotifyTimeDisableOption {
					settings.DailyNotifyTime = nil
				} else {
					dt, _ := time.Parse(time.RFC3339, val)
					settings.DailyNotifyTime = &dt
				}
			case conf.TenMinuteNotifyDialogOption:
				settings.TenMinutesNotify = value.(bool)
			case conf.OneMinuteNotifyDialogOption:
				settings.OneMinutesNotify = value.(bool)
			default:
				hc.pluginAPI.LogWarn("Unknown selector: '" + selector + "' in setup dialog")
			}
		}
		repository.SaveSettings(hc.pluginAPI, userId, *settings)

		events, _ := hc.calendar.LoadCalendar(userId)
		hc.sender.SendEvents(userId, conf.GetTodayEventsTitle(settings.GetUserNow()), events)
		hc.workspace.AddUser(userId)
		hc.scheduler.AddCronJobs(userId)
	}
}

func submitDialogRequestFromJson(data io.Reader) *model.SubmitDialogRequest {
	var o *model.SubmitDialogRequest
	err := json.NewDecoder(data).Decode(&o)
	if err != nil {
		return nil
	}
	return o
}

func (hc *HttpController) handleIcon(w http.ResponseWriter, r *http.Request) {
	bundlePath, err := hc.pluginAPI.GetBundlePath()
	if err != nil {
		hc.pluginAPI.LogWarn("failed to get bundle path", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=604800")
	http.ServeFile(w, r, filepath.Join(bundlePath, "assets", conf.IconImgName))
}
