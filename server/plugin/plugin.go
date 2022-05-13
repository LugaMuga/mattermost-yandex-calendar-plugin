package plugin

import (
	"github.com/blang/semver/v4"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/conf"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/controller"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/service"
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/mattermost/mattermost-server/v6/plugin"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin
	client *pluginapi.Client

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
	serverConfig  *model.Config

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
	p.client = pluginapi.NewClient(p.API, p.Driver)
	command, err := controller.GetHookCommand(p.API)

	if err != nil {
		return errors.Wrap(err, "Failed to get command")
	}
	err = p.API.RegisterCommand(command)
	if err != nil {
		return errors.Wrap(err, "Failed on register command")
	}

	botID, err := p.client.Bot.EnsureBot(&model.Bot{
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
	p.API.GetServerVersion()

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
	p.sender = service.NewSenderService(manifest.ID, p.botID, p.API, p.serverConfig)
	p.workspace = service.NewWorkspaceService(p.API)
	p.user = service.NewUserService(p.API, p.getServerVersion(), p.sender, p.calendar)
	p.scheduler = service.NewSchedulerService(p.API, p.workspace, p.user)
}

func (p *Plugin) registerControllers() {
	p.http = controller.NewHttpController(p.API, manifest.Version, p.calendar, p.user, p.sender, p.scheduler, p.workspace)
	p.hook = controller.NewHookController(p.API, p.botID, p.calendar, p.user, p.sender, p.scheduler, p.workspace)
}

func (p *Plugin) getServerVersion() *semver.Version {
	serverVersion, err := semver.Parse(p.API.GetServerVersion())
	if err != nil {
		p.API.LogError("Can't parse server version")
		// Should be synchronized with plugin json
		return &semver.Version{
			Major: 5,
			Minor: 10,
			Patch: 0,
		}
	}
	return &serverVersion
}
