package daemon

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2024 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/essentialkaos/ek/v12/cache"
	"github.com/essentialkaos/ek/v12/fmtc"
	"github.com/essentialkaos/ek/v12/knf"
	"github.com/essentialkaos/ek/v12/log"
	"github.com/essentialkaos/ek/v12/options"
	"github.com/essentialkaos/ek/v12/signal"
	"github.com/essentialkaos/ek/v12/support"
	"github.com/essentialkaos/ek/v12/support/deps"
	"github.com/essentialkaos/ek/v12/terminal"
	"github.com/essentialkaos/ek/v12/terminal/tty"
	"github.com/essentialkaos/ek/v12/usage"

	knfv "github.com/essentialkaos/ek/v12/knf/validators"
	knff "github.com/essentialkaos/ek/v12/knf/validators/fs"
	knfn "github.com/essentialkaos/ek/v12/knf/validators/network"
	knfr "github.com/essentialkaos/ek/v12/knf/validators/regexp"

	"github.com/essentialkaos/go-badge"

	"github.com/valyala/fasthttp"

	"github.com/essentialkaos/updown-badge-server/api"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Basic service info
const (
	APP  = "UpDownBadgeServer"
	VER  = "1.3.2"
	DESC = "Service for generating badges for updown.io checks"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	MIN_PORT         = 1025
	MAX_PORT         = 65535
	MIN_CACHE_PERIOD = 60   // 1 min
	MAX_CACHE_PERIOD = 3600 // 1 hour
	MIN_PROCS        = 1
	MAX_PROCS        = 256
)

// Options
const (
	OPT_CONFIG   = "c:config"
	OPT_NO_COLOR = "nc:no-color"
	OPT_HELP     = "h:help"
	OPT_VER      = "v:version"

	OPT_VERB_VER = "vv:verbose-version"
)

// Configuration file properties
const (
	MAIN_MAX_PROCS  = "main:max-procs"
	UPDOWN_API_KEY  = "updown:api-key"
	BADGE_FONT      = "badge:font"
	BADGE_STYLE     = "badge:style"
	CACHE_PERIOD    = "cache:period"
	SERVER_IP       = "server:ip"
	SERVER_PORT     = "server:port"
	SERVER_REDIRECT = "server:redirect"
	LOG_DIR         = "log:dir"
	LOG_FILE        = "log:file"
	LOG_MODE        = "log:mode"
	LOG_LEVEL       = "log:level"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// optMap contains information about all supported options
var optMap = options.Map{
	OPT_CONFIG:   {Value: "/etc/updown-badge-server.knf"},
	OPT_NO_COLOR: {Type: options.BOOL},
	OPT_HELP:     {Type: options.BOOL},
	OPT_VER:      {Type: options.MIXED},

	OPT_VERB_VER: {Type: options.BOOL},
}

var udAPI *api.API
var server *fasthttp.Server
var badgeCache *cache.Cache
var badgeGen *badge.Generator
var badgeStyle string
var redirectURL string

// ////////////////////////////////////////////////////////////////////////////////// //

// Run is main utility function
func Run(gomod []byte) {
	preConfigureUI()

	_, errs := options.Parse(optMap)

	if !errs.IsEmpty() {
		terminal.Error("Options parsing errors:")
		terminal.Error(errs.String())
		os.Exit(1)
	}

	configureUI()

	switch {
	case options.GetB(OPT_VER):
		genAbout().Print(options.GetS(OPT_VER))
		os.Exit(0)
	case options.GetB(OPT_VERB_VER):
		support.Collect(APP, VER).
			WithDeps(deps.Extract(gomod)).
			Print()
		os.Exit(0)
	case options.GetB(OPT_HELP):
		genUsage().Print()
		os.Exit(0)
	}

	loadConfig()
	validateConfig()
	configureRuntime()
	registerSignalHandlers()
	setupLogger()

	log.Aux(strings.Repeat("-", 80))
	log.Aux("%s %s starting…", APP, VER)

	start()
}

// preConfigureUI preconfigures user interface
func preConfigureUI() {
	if !tty.IsTTY() || tty.IsSystemd() {
		fmtc.DisableColors = true
	}
}

// configureUI configures user interface
func configureUI() {
	if options.GetB(OPT_NO_COLOR) {
		fmtc.DisableColors = true
	}
}

// loadConfig reads and parses configuration file
func loadConfig() {
	err := knf.Global(options.GetS(OPT_CONFIG))

	if err != nil {
		log.Crit(err.Error())
		os.Exit(1)
	}
}

// validateConfig validates configuration file values
func validateConfig() {
	errs := knf.Validate([]*knf.Validator{
		{UPDOWN_API_KEY, knfv.Empty, nil},
		{SERVER_PORT, knfv.Empty, nil},

		{MAIN_MAX_PROCS, knfv.TypeNum, nil},
		{CACHE_PERIOD, knfv.TypeNum, nil},
		{SERVER_PORT, knfv.TypeNum, nil},

		{BADGE_FONT, knff.Perms, "FRS"},
		{BADGE_STYLE, knfv.NotContains, []string{
			STYLE_PLASTIC, STYLE_FLAT, STYLE_FLAT_SQUARE,
		}},

		{UPDOWN_API_KEY, knfv.NotLen, 23},
		{UPDOWN_API_KEY, knfr.Regexp, "^ro-[0-9A-Za-z]{20}$"},

		{MAIN_MAX_PROCS, knfv.Less, MIN_PROCS},
		{MAIN_MAX_PROCS, knfv.Greater, MAX_PROCS},

		{CACHE_PERIOD, knfv.Less, MIN_CACHE_PERIOD},
		{CACHE_PERIOD, knfv.Greater, MAX_CACHE_PERIOD},

		{SERVER_PORT, knfv.Less, MIN_PORT},
		{SERVER_PORT, knfv.Greater, MAX_PORT},

		{SERVER_IP, knfn.IP, nil},
		{SERVER_REDIRECT, knfn.URL, nil},

		{LOG_DIR, knff.Perms, "DW"},
		{LOG_DIR, knff.Perms, "DX"},

		{LOG_LEVEL, knfv.NotContains, []string{
			"debug", "info", "warn", "error", "crit",
		}},
	})

	if len(errs) != 0 {
		for _, err := range errs {
			log.Crit(err.Error())
		}

		os.Exit(1)
	}
}

// configureRuntime configures runtime
func configureRuntime() {
	if !knf.HasProp(MAIN_MAX_PROCS) {
		return
	}

	runtime.GOMAXPROCS(knf.GetI(MAIN_MAX_PROCS))
}

// registerSignalHandlers registers signal handlers
func registerSignalHandlers() {
	signal.Handlers{
		signal.TERM: termSignalHandler,
		signal.INT:  intSignalHandler,
		signal.HUP:  hupSignalHandler,
	}.TrackAsync()
}

// setupLogger configures logger subsystems
func setupLogger() {
	err := log.Set(knf.GetS(LOG_FILE), knf.GetM(LOG_MODE, 0644))

	if err != nil {
		log.Crit(err.Error())
		os.Exit(1)
	}

	err = log.MinLevel(knf.GetS(LOG_LEVEL))

	if err != nil {
		log.Crit(err.Error())
		os.Exit(1)
	}
}

// start configures and starts all subsystems
func start() {
	var err error

	badgeStyle = knf.GetS(BADGE_STYLE, "flat")
	redirectURL = knf.GetS(SERVER_REDIRECT)

	badgeGen, err = badge.NewGenerator(knf.GetS(BADGE_FONT), 11)

	if err != nil {
		log.Crit("Can't load font for badges: %v", err)
		shutdown(1)
	}

	udAPI = api.NewClient(knf.GetS(UPDOWN_API_KEY))
	udAPI.SetUserAgent(APP, VER)

	badgeCache = cache.New(knf.GetD(CACHE_PERIOD, knf.Second), time.Minute)

	err = startHTTPServer(knf.GetS(SERVER_IP), knf.GetS(SERVER_PORT))

	if err != nil {
		log.Crit("Can't start HTTP server: %v", err)
		shutdown(1)
	}
}

// intSignalHandler is INT signal handler
func intSignalHandler() {
	log.Aux("Received INT signal, shutdown…")
	shutdown(0)
}

// termSignalHandler is TERM signal handler
func termSignalHandler() {
	log.Aux("Received TERM signal, shutdown…")
	shutdown(0)
}

// hupSignalHandler is HUP signal handler
func hupSignalHandler() {
	log.Info("Received HUP signal, log will be reopened…")
	log.Reopen()
	log.Info("Log reopened by HUP signal")
}

// shutdown stops daemon
func shutdown(code int) {
	if server != nil {
		err := server.Shutdown()

		if err != nil {
			log.Error("Can't gracefully shut down HTTP server: %v", err)
		}
	}

	os.Exit(code)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// genUsage generates usage info
func genUsage() *usage.Info {
	info := usage.NewInfo()

	info.AddOption(OPT_CONFIG, "Path to configuration file", "file")
	info.AddOption(OPT_NO_COLOR, "Disable colors in output")
	info.AddOption(OPT_HELP, "Show this help message")
	info.AddOption(OPT_VER, "Show version")

	return info
}

// genAbout generates info about version
func genAbout() *usage.About {
	return &usage.About{
		App:     APP,
		Version: VER,
		Desc:    DESC,
		Year:    2009,
		Owner:   "ESSENTIAL KAOS",
		License: "Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>",
	}
}
