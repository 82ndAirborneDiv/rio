package tui

import (
	"github.com/gdamore/tcell"
	"github.com/rancher/axe/throwing"
	"github.com/rancher/axe/throwing/datafeeder"
	"github.com/rancher/axe/throwing/types"
)

const (
	serviceKind         = "service"
	routeKind           = "router"
	appKind             = "app"
	podKind             = "pod"
	containerKind       = "container"
	configKind          = "config"
	publicdomainKind    = "publicdomain"
	externalServiceKind = "externalservice"
	buildKind           = "build"
)

var (
	defaultBackGroundColor = tcell.ColorBlack

	colorStyles []string

	RootPage = appKind

	Shortcuts = [][]string{
		// CRUD
		{"Key c", "Create"},
		{"Key i", "Inspect"},
		{"Key e", "Edit"},
		{"Key d", "Delete"},

		// exec and log
		{"Key l", "Logs"},
		{"Key x", "Exec"},

		// view pods and revisions
		{"Key p", "View Pods"},
		{"Key v", "View revision"},

		{"Key /", "Search"},
		{"Key Ctrl+h", "Hit Endpoint"},
		{"Key Ctrl+r", "Refresh"},
		{"Key Ctrl+s", "Show system resource"},
	}

	Footers = []types.ResourceView{
		{
			Title: "Apps",
			Kind:  appKind,
			Index: 1,
		},
		{
			Title: "Routes",
			Kind:  routeKind,
			Index: 2,
		},
		{
			Title: "ExternalService",
			Kind:  externalServiceKind,
			Index: 3,
		},
		{
			Title: "PublicDomain",
			Kind:  publicdomainKind,
			Index: 4,
		},
		{
			Title: "Config",
			Kind:  configKind,
			Index: 5,
		},
		{
			Title: "Build",
			Kind:  buildKind,
			Index: 6,
		},
	}

	PageNav = map[rune]string{
		'1': appKind,
		'2': routeKind,
		'3': externalServiceKind,
		'4': publicdomainKind,
		'5': configKind,
		'6': buildKind,
	}

	tableEventHandler = func(t *throwing.TableView) func(event *tcell.EventKey) *tcell.EventKey {
		return func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEscape:
				escape(t)
			case tcell.KeyEnter:
				switch t.GetResourceKind() {
				case appKind:
					revisions(t)
				case containerKind:
					logs("", t)
				case podKind:
					containers(t)
				case serviceKind:
					pods(t)
				}
			case tcell.KeyDelete:
				rm(t)
			case tcell.KeyCtrlH:
				hit(t)
			case tcell.KeyCtrlP:
				promote(t)
			case tcell.KeyCtrlR, tcell.KeyF5:
				t.Refresh()
			case tcell.KeyRune:
				switch event.Rune() {
				case 'q':
					t.GetApplication().Stop()
				case 's':
					showSystem = !showSystem
					t.Refresh()
				case 'i':
					inspect("yaml", t)
				case 'l':
					logs("", t)
				case 'x':
					execute("", t)
				case '/':
					t.ShowSearch()
				default:
					t.Navigate(event.Rune())
				case 'p':
					pods(t)
				case 'e':
					edit(t)
				case 'r':
					revisions(t)
				}
			}
			return event
		}
	}

	App = types.ResourceKind{
		Title: " Apps ",
		Kind:  appKind,
	}

	Route = types.ResourceKind{
		Title: " Routers ",
		Kind:  routeKind,
	}

	Config = types.ResourceKind{
		Title: " Configs ",
		Kind:  configKind,
	}

	PublicDomain = types.ResourceKind{
		Title: " PublicDomains ",
		Kind:  publicdomainKind,
	}

	Service = types.ResourceKind{
		Title: " Services ",
		Kind:  serviceKind,
	}

	ExternalService = types.ResourceKind{
		Title: " ExternalServices ",
		Kind:  externalServiceKind,
	}

	Pod = types.ResourceKind{
		Title: " Pods ",
		Kind:  podKind,
	}

	Build = types.ResourceKind{
		Title: " Builds ",
		Kind:  buildKind,
	}

	Container = types.ResourceKind{
		Title: " Containers ",
		Kind:  containerKind,
	}

	DefaultAction = []types.Action{
		{
			Name:        "Inspect",
			Shortcut:    "I",
			Description: "inspect a resource",
		},
		{
			Name:        "Edit",
			Shortcut:    "E",
			Description: "edit a resource",
		},
		{
			Name:        "Delete",
			Shortcut:    "Del",
			Description: "delete a resource",
		},
		{
			Name:        "Refresh",
			Shortcut:    "Ctrl+R",
			Description: "Refresh Page",
		},
		{
			Name:        "ShowSystem",
			Shortcut:    "S",
			Description: "Show system resource",
		},
		{
			Name:        "Escape",
			Shortcut:    "Esc",
			Description: "Go to the previous level",
		},
		{
			Name:        "Quit",
			Shortcut:    "Q",
			Description: "Quit console",
		},
	}

	ExecAction = types.Action{
		Name:        "Exec",
		Shortcut:    "X",
		Description: "exec into a container or service",
	}

	LogAction = types.Action{
		Name:        "Log",
		Shortcut:    "L",
		Description: "view logs of a service",
	}

	AppAction = types.Action{
		Name:        "Revisions",
		Shortcut:    "R",
		Description: "view revisions of a app",
	}

	HitAction = types.Action{
		Name:        "Hit",
		Shortcut:    "Ctrl+H",
		Description: "hit endpoint of a service(need jq and curl)",
	}

	ServiceAction = types.Action{
		Name:        "Pods",
		Shortcut:    "P",
		Description: "view pods of a service or app",
	}

	ViewMap = map[string]types.View{
		appKind: {
			Actions: append(DefaultAction, AppAction, HitAction, ServiceAction, ExecAction, LogAction),
			Kind:    App,
			Feeder:  datafeeder.NewDataFeeder(AppRefresher),
		},
		routeKind: {
			Actions: append(DefaultAction, HitAction),
			Kind:    Route,
			Feeder:  datafeeder.NewDataFeeder(RouteRefresher),
		},
		externalServiceKind: {
			Actions: DefaultAction,
			Kind:    ExternalService,
			Feeder:  datafeeder.NewDataFeeder(ExternalRefresher),
		},
		configKind: {
			Actions: DefaultAction,
			Kind:    Config,
			Feeder:  datafeeder.NewDataFeeder(ConfigRefresher),
		},
		publicdomainKind: {
			Actions: DefaultAction,
			Kind:    PublicDomain,
			Feeder:  datafeeder.NewDataFeeder(PublicDomainRefresher),
		},
		serviceKind: {
			Actions: append(DefaultAction, HitAction, ServiceAction, ExecAction, LogAction),
			Kind:    Service,
			Feeder:  datafeeder.NewDataFeeder(ServiceRefresher),
		},
		podKind: {
			Actions: append(DefaultAction, ExecAction, LogAction),
			Kind:    Pod,
			Feeder:  datafeeder.NewDataFeeder(PodRefresher),
		},
		buildKind: {
			Actions: append(DefaultAction, LogAction),
			Kind:    Build,
			Feeder:  datafeeder.NewDataFeeder(BuildRefresher),
		},
		containerKind: {
			Actions: append(DefaultAction, ExecAction, LogAction),
			Kind:    Container,
			Feeder:  datafeeder.NewDataFeeder(ContainerRefresher),
		},
	}

	drawer = types.Drawer{
		RootPage:  RootPage,
		Shortcuts: Shortcuts,
		ViewMap:   ViewMap,
		PageNav:   PageNav,
		Footers:   Footers,
		Menu:      ViewMap[RootPage].Actions,
	}
)
