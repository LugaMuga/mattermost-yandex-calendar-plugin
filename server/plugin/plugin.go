package plugin

import (
	"github.com/blang/semver/v4"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/conf"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/controller"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/repository"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/service"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/util"
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

	logger     *util.Logger
	repo       *Repo
	service    *Service
	controller *Controller

	botId string
}

type Repo struct {
	workspace   *repository.WorkspaceRepo
	credentials *repository.CredentialsRepo
}

type Service struct {
	calendar  *service.Calendar
	sender    *service.Sender
	workspace *service.Workspace
	user      *service.User
	scheduler *service.Scheduler
}

type Controller struct {
	http *controller.HttpController
	hook *controller.HookController
}

//OnActivate function ensures what bot does when become actived
func (p *Plugin) OnActivate() error {
	p.logger = util.NewLogger(p.API)
	p.client = pluginapi.NewClient(p.API, p.Driver)
	command, err := controller.GetHookCommand(p.API)

	if err != nil {
		return errors.Wrap(err, "Failed to get command")
	}
	err = p.API.RegisterCommand(command)
	if err != nil {
		return errors.Wrap(err, "Failed on register command")
	}

	botId, err := p.client.Bot.EnsureBot(&model.Bot{
		Username:    "yandex.calendar",
		DisplayName: "Yandex Calendar",
		Description: "Created by the Yandex Calendar plugin.",
	})
	if err != nil {
		return errors.Wrap(err, "failed to ensure google calendar bot")
	}
	p.botId = botId

	bundlePath, err := p.API.GetBundlePath()
	if err != nil {
		return errors.Wrap(err, "couldn't get bundle path")
	}

	profileImage, err := ioutil.ReadFile(filepath.Join(bundlePath, "assets", conf.IconImgName))
	if err != nil {
		return errors.Wrap(err, "couldn't read profile image")
	}

	appErr := p.API.SetProfileImage(botId, profileImage)
	if appErr != nil {
		return errors.Wrap(appErr, "couldn't set profile image")
	}

	p.registerRepos()
	p.registerServices()
	p.registerControllers()
	return nil
}

// ServeHTTP method for register custom HTTP handler
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	p.API.LogDebug("New request:", "Host", r.Host, "RequestURI", r.RequestURI, "Method", r.Method)
	p.controller.http.ServeHTTP(c, w, r)
}

// ExecuteCommand handler for hook API
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	return p.controller.hook.ExecuteCommand(c, args)
}

func (p *Plugin) registerRepos() {
	p.repo = &Repo{
		workspace:   repository.NewWorkspaceRepo(p.logger, p.API),
		credentials: repository.NewCredentialsRepo(p.logger, p.API),
	}
}

func (p *Plugin) registerServices() {
	p.service = &Service{}
	p.service.calendar = service.NewCalendarService(p.logger, p.API, p.getConfiguration().ServerUrl, p.repo.credentials)
	p.service.sender = service.NewSenderService(manifest.ID, p.botId, p.logger, p.API, p.supportedUserCustomStatus(), p.serverConfig)
	p.service.workspace = service.NewWorkspaceService(p.repo.workspace)
	p.service.user = service.NewUserService(p.logger, p.API, p.supportedUserCustomStatus(), p.repo.credentials, p.service.sender, p.service.calendar)
	p.service.scheduler = service.NewSchedulerService(p.logger, p.API, p.service.workspace, p.service.user)

	p.service.scheduler.InitCronJobs()
}

func (p *Plugin) registerControllers() {
	p.controller = &Controller{}
	p.controller.http = controller.NewHttpController(p.API, manifest.Version,
		p.service.calendar, p.service.user, p.service.sender, p.service.scheduler, p.service.workspace)
	p.controller.hook = controller.NewHookController(p.API, p.botId, p.service.calendar, p.service.user, p.service.sender, p.service.scheduler, p.service.workspace)
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

func (p *Plugin) supportedUserCustomStatus() bool {
	return p.getServerVersion().GTE(semver.Version{
		Major: 6,
		Minor: 2,
	})
}
