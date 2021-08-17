package daemon

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2021 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"strings"

	"pkg.re/essentialkaos/ek.v12/log"
	"pkg.re/essentialkaos/ek.v12/strutil"

	"pkg.re/essentialkaos/go-badge.v1"

	"github.com/valyala/fasthttp"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	CHECK_UNKNOWN uint8 = 0
	CHECK_STATUS        = 1
	CHECK_UPTIME        = 2
	CHECK_APDEX         = 3
)

const (
	STYLE_PLASTIC     = "plastic"
	STYLE_FLAT        = "flat"
	STYLE_FLAT_SQUARE = "flat-square"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// startHTTPServer start HTTP server
func startHTTPServer(ip, port string) error {
	addr := ip + ":" + port

	log.Info("HTTP server is started on %s", addr)

	server = &fasthttp.Server{
		Handler: fastHTTPHandler,
		Name:    APP + "/" + VER,
	}

	return server.ListenAndServe(addr)
}

// fastHTTPHandler is a handler for http requests
func fastHTTPHandler(ctx *fasthttp.RequestCtx) {
	defer requestRecover(ctx)

	path := string(ctx.Path())

	log.Debug("Got request '%s'", path)

	if !isValidRequestPath(path) {
		if redirectURL == "" {
			ctx.SetStatusCode(404)
		} else {
			ctx.Redirect(redirectURL, 301)
		}

		return
	}

	badgeHandler(ctx, path)
}

// requestRecover recovers panic in request
func requestRecover(ctx *fasthttp.RequestCtx) {
	r := recover()

	if r != nil {
		log.Error("Recovered internal error in HTTP request handler: %v", r)
		ctx.SetStatusCode(501)
	}
}

// badgeHandler handler badge request
func badgeHandler(ctx *fasthttp.RequestCtx, path string) {
	token, checkType := parsePath(path)

	log.Debug("Generating badge for token %s (checkType: %d)", token, checkType)

	ctx.Response.Header.Set("Content-Type", "image/svg+xml")
	ctx.Response.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	ctx.Response.Header.Set("Pragma", "no-cache")
	ctx.Response.Header.Set("Expires", "0")

	ctx.Write(getBadge(token, checkType))
	ctx.SetStatusCode(200)
}

// getBadge returns badge for check with given token and check type
func getBadge(token string, checkType uint8) []byte {
	if checkType == CHECK_UNKNOWN {
		token = "UNKNOWN"
	}

	cacheKey := fmt.Sprintf("%s:%d", token, checkType)

	if badgeCache.Has(cacheKey) {
		return badgeCache.Get(cacheKey).([]byte)
	}

	var badgeData []byte

	switch checkType {
	case CHECK_STATUS:
		badgeData = genStatusBadge(token)
	case CHECK_UPTIME:
		badgeData = genUptimeBadge(token)
	case CHECK_APDEX:
		badgeData = genApdexBadge(token)
	default:
		badgeData = genBadge("status", "unknown", badge.COLOR_INACTIVE)
	}

	badgeCache.Set(cacheKey, badgeData)

	return badgeData
}

// genStatusBadge generates status badge
func genStatusBadge(token string) []byte {
	label := "status"
	status, err := udAPI.GetStatus(token)

	if err != nil {
		log.Error("Can't get status info: %v", err)
	}

	switch {
	case err != nil || status == nil:
		return genBadge(label, "unknown", badge.COLOR_INACTIVE)
	case status.IsDown:
		return genBadge(label, "down", badge.COLOR_CRITICAL)
	}

	return genBadge(label, "up", badge.COLOR_SUCCESS)
}

// genUptimeBadge generates uptime badge
func genUptimeBadge(token string) []byte {
	var value string

	label := "uptime"
	status, err := udAPI.GetStatus(token)

	if err != nil {
		log.Error("Can't get status info: %v", err)
	} else {
		value = fmt.Sprintf("%.2f%%", status.Uptime)
	}

	switch {
	case err != nil || status == nil:
		return genBadge(label, "unknown", badge.COLOR_INACTIVE)
	case status.Uptime < 90:
		return genBadge(label, value, badge.COLOR_RED)
	case status.Uptime < 95:
		return genBadge(label, value, badge.COLOR_ORANGE)
	case status.Uptime < 100:
		return genBadge(label, value, badge.COLOR_YELLOW)
	}

	return genBadge(label, "100%", badge.COLOR_GREEN)
}

// genApdexBadge generates apdex badge
func genApdexBadge(token string) []byte {
	var value string

	label := "apdex"
	apdex, err := udAPI.GetApdex(token)

	if err != nil {
		log.Error("Can't get apdex info: %v", err)
	} else {
		value = fmt.Sprintf("%.2f", apdex.Value)
	}

	switch {
	case err != nil || apdex == nil:
		return genBadge(label, "unknown", badge.COLOR_INACTIVE)
	case apdex.Value < 0.8:
		return genBadge(label, value, badge.COLOR_RED)
	case apdex.Value < 0.9:
		return genBadge(label, value, badge.COLOR_ORANGE)
	case apdex.Value < 0.995:
		return genBadge(label, value, badge.COLOR_YELLOW)
	}

	return genBadge(label, "1.0", badge.COLOR_GREEN)
}

// genBadge generates badge with global style
func genBadge(label, value, color string) []byte {
	switch badgeStyle {
	case STYLE_PLASTIC:
		return badgeGen.GeneratePlastic(label, value, color)

	case STYLE_FLAT:
		return badgeGen.GenerateFlat(label, value, color)

	case STYLE_FLAT_SQUARE:
		return badgeGen.GenerateFlatSquare(label, value, color)
	}

	return nil
}

// isValidRequestPath checks if request path is ok
func isValidRequestPath(path string) bool {
	if strings.Count(path, "/") != 2 {
		return false
	}

	if !strings.HasSuffix(path, ".svg") {
		return false
	}

	return true
}

// parsePath parses request path
func parsePath(path string) (string, uint8) {
	token := strutil.ReadField(path, 0, false, "/")
	badge := strutil.ReadField(path, 1, false, "/")

	switch badge {
	case "status.svg":
		return token, CHECK_STATUS
	case "uptime.svg":
		return token, CHECK_UPTIME
	case "apdex.svg":
		return token, CHECK_APDEX
	}

	return "", 0
}
