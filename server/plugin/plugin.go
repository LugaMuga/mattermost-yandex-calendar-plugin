package plugin

import (
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/conf"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/controller"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/service"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/mattermost/mattermost-server/v5/plugin"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
	ServerConfig  *model.Config

	http *controller.HttpController
	hook *controller.HookController

	calendar  *service.Calendar
	sender    *service.Sender
	workspace *service.Workspace
	user      *service.User
	scheduler *service.Scheduler

	botID string
}

//OnActivate function ensures what bot does when become actived
func (p *Plugin) OnActivate() error {
	command, err := controller.GetHookCommand(p.API)

	if err != nil {
		return errors.Wrap(err, "failed to get command")
	}
	p.API.RegisterCommand(command)

	botID, err := p.Helpers.EnsureBot(&model.Bot{
		Username:    "yandex.calendar",
		DisplayName: "Yandex Calendar",
		Description: "Created by the Yandex Calendar plugin.",
	})
	if err != nil {
		return errors.Wrap(err, "failed to ensure google calendar bot")
	}
	p.botID = botID

	bundlePath, err := p.API.GetBundlePath()
	if err != nil {
		return errors.Wrap(err, "couldn't get bundle path")
	}

	profileImage, err := ioutil.ReadFile(filepath.Join(bundlePath, "assets", conf.IconImgName))
	if err != nil {
		return errors.Wrap(err, "couldn't read profile image")
	}

	appErr := p.API.SetProfileImage(botID, profileImage)
	if appErr != nil {
		return errors.Wrap(appErr, "couldn't set profile image")
	}

	p.registerServices()
	p.scheduler.InitCronJobs()
	p.registerControllers()

	return nil
}

// ServeHTTP method for register custom HTTP handler
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	p.API.LogDebug("New request:", "Host", r.Host, "RequestURI", r.RequestURI, "Method", r.Method)
	p.http.ServeHTTP(c, w, r)
}

// ExecuteCommand handler for hook API
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	return p.hook.ExecuteCommand(c, args)
}

func (p *Plugin) registerServices() {
	p.calendar = service.NewCalendarService(p.API)
	p.sender = service.NewSenderService(manifest.Id, p.botID, p.API, p.ServerConfig)
	p.workspace = service.NewWorkspaceService(p.API)
	p.user = service.NewUserService(p.API, p.sender, p.calendar)
	p.scheduler = service.NewSchedulerService(p.API, p.workspace, p.user)
}

func (p *Plugin) registerControllers() {
	p.http = controller.NewHttpController(p.API, manifest.Version, p.calendar, p.user, p.sender, p.scheduler, p.workspace)
	p.hook = controller.NewHookController(p.API, p.botID, p.calendar, p.user, p.sender, p.scheduler, p.workspace)
}
