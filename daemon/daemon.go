package daemon

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2021 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"os"
	"runtime"
	"strings"
	"time"

	"pkg.re/essentialkaos/ek.v12/cache"
	"pkg.re/essentialkaos/ek.v12/fmtc"
	"pkg.re/essentialkaos/ek.v12/knf"
	"pkg.re/essentialkaos/ek.v12/log"
	"pkg.re/essentialkaos/ek.v12/options"
	"pkg.re/essentialkaos/ek.v12/pid"
	"pkg.re/essentialkaos/ek.v12/signal"
	"pkg.re/essentialkaos/ek.v12/usage"

	knfv "pkg.re/essentialkaos/ek.v12/knf/validators"
	knff "pkg.re/essentialkaos/ek.v12/knf/validators/fs"
	knfn "pkg.re/essentialkaos/ek.v12/knf/validators/network"
	knfr "pkg.re/essentialkaos/ek.v12/knf/validators/regexp"

	"pkg.re/essentialkaos/go-badge.v1"

	"github.com/valyala/fasthttp"

	"github.com/essentialkaos/updown-badge-server/api"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Basic service info
const (
	APP  = "UpDownBadgeServer"
	VER  = "1.1.0"
	DESC = "Service for generating badges for updown.io checks"
)

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
	OPT_VERSION  = "v:version"
)

// Configuration file properties
const (
	MAIN_MAX_PROCS = "main:max-procs"
	UPDOWN_API_KEY = "updown:api-key"
	BADGE_FONT     = "badge:font"
	BADGE_STYLE    = "badge:style"
	CACHE_PERIOD   = "cache:period"
	SERVER_IP      = "server:ip"
	SERVER_PORT    = "server:port"
	LOG_DIR        = "log:dir"
	LOG_FILE       = "log:file"
	LOG_PERMS      = "log:perms"
	LOG_LEVEL      = "log:level"
)

// Pid file info
const (
	PID_DIR  = "/var/run/updown-badge-server"
	PID_FILE = "updown-badge-server.pid"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// optMap contains information about all supported options
var optMap = options.Map{
	OPT_CONFIG:   {Value: "/etc/updown-badge-server.knf"},
	OPT_NO_COLOR: {Type: options.BOOL},
	OPT_HELP:     {Type: options.BOOL, Alias: "u:usage"},
	OPT_VERSION:  {Type: options.BOOL, Alias: "ver"},
}

var udAPI *api.API
var server *fasthttp.Server
var badgeCache *cache.Cache
var badgeGen *badge.Generator
var badgeStyle string

// ////////////////////////////////////////////////////////////////////////////////// //

func Init() {
	_, errs := options.Parse(optMap)

	if len(errs) != 0 {
		for _, err := range errs {
			printError(err.Error())
		}

		os.Exit(1)
	}

	configureUI()

	if options.GetB(OPT_VERSION) {
		os.Exit(showAbout())
	}

	if options.GetB(OPT_HELP) {
		os.Exit(showUsage())
	}

	loadConfig()
	validateConfig()
	configureRuntime()
	registerSignalHandlers()
	setupLogger()
	createPidFile()

	log.Aux(strings.Repeat("-", 80))
	log.Aux("%s %s starting…", APP, VER)

	start()
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
		printErrorAndExit(err.Error())
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

		{LOG_DIR, knff.Perms, "DW"},
		{LOG_DIR, knff.Perms, "DX"},

		{LOG_LEVEL, knfv.NotContains, []string{
			"debug", "info", "warn", "error", "crit",
		}},
	})

	if len(errs) != 0 {
		printError("Error while configuration file validation:")

		for _, err := range errs {
			printError("  %v", err)
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

// setupLogger confugures logger subsystems
func setupLogger() {
	err := log.Set(knf.GetS(LOG_FILE), knf.GetM(LOG_PERMS, 644))

	if err != nil {
		printErrorAndExit(err.Error())
	}

	err = log.MinLevel(knf.GetS(LOG_LEVEL))

	if err != nil {
		printErrorAndExit(err.Error())
	}
}

// createPidFile creates PID file
func createPidFile() {
	pid.Dir = PID_DIR

	err := pid.Create(PID_FILE)

	if err != nil {
		printErrorAndExit(err.Error())
	}
}

// start configures and starts all subsystems
func start() {
	var err error

	badgeStyle = knf.GetS(BADGE_STYLE, "flat")
	badgeGen, err = badge.NewGenerator(knf.GetS(BADGE_FONT), 11)

	if err != nil {
		log.Crit("Can't load font for badges: %v", err)
		shutdown(1)
	}

	udAPI = api.NewClient(knf.GetS(UPDOWN_API_KEY))
	udAPI.SetUserAgent(APP, VER)

	badgeCache = cache.New(knf.GetD(CACHE_PERIOD), time.Minute)

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

// printError prints error message to console
func printError(f string, a ...interface{}) {
	fmtc.Fprintf(os.Stderr, "{r}"+f+"{!}\n", a...)
}

// printError prints warning message to console
func printWarn(f string, a ...interface{}) {
	fmtc.Fprintf(os.Stderr, "{y}"+f+"{!}\n", a...)
}

// printErrorAndExit print error mesage and exit with exit code 1
func printErrorAndExit(f string, a ...interface{}) {
	printError(f, a...)
	os.Exit(1)
}

// shutdown stops deamon
func shutdown(code int) {
	if server != nil {
		err := server.Shutdown()

		if err != nil {
			log.Error("Can't gracefully shut down HTTP server: %v", err)
		}
	}

	pid.Remove(PID_FILE)
	os.Exit(code)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// showUsage prints usage info
func showUsage() int {
	info := usage.NewInfo()

	info.AddOption(OPT_CONFIG, "Path to configuration file", "file")
	info.AddOption(OPT_NO_COLOR, "Disable colors in output")
	info.AddOption(OPT_HELP, "Show this help message")
	info.AddOption(OPT_VERSION, "Show version")

	info.Render()

	return 0
}

// showAbout prints info about version
func showAbout() int {
	usage := &usage.About{
		App:     APP,
		Version: VER,
		Desc:    DESC,
		Year:    2009,
		Owner:   "ESSENTIAL KAOS",
		License: "Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>",
	}

	usage.Render()

	return 0
}