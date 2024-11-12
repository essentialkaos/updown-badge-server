package daemon

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2024 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"strings"

	"github.com/essentialkaos/ek/v13/color"
	"github.com/essentialkaos/ek/v13/easing"
	"github.com/essentialkaos/ek/v13/log"
	"github.com/essentialkaos/ek/v13/mathutil"
	"github.com/essentialkaos/ek/v13/strutil"

	"github.com/essentialkaos/go-badge"

	"github.com/valyala/fasthttp"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	CHECK_UNKNOWN uint8 = 0
	CHECK_STATUS  uint8 = 1
	CHECK_UPTIME  uint8 = 2
	CHECK_APDEX   uint8 = 3
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

	if err != nil || status == nil {
		return genBadge(label, "unknown", badge.COLOR_INACTIVE)
	}

	if status.Uptime >= 100 {
		return genBadge(label, "100%", getColorForStatus(1))
	}

	v := mathutil.Between((status.Uptime-70)/30, 0, 1)

	return genBadge(label, value, getColorForStatus(v))
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

	if err != nil || apdex == nil {
		return genBadge(label, "unknown", badge.COLOR_INACTIVE)
	}

	if apdex.Value > 0.995 {
		return genBadge(label, "1.0", getColorForStatus(1))
	}

	v := mathutil.Between((apdex.Value-0.7)/0.3, 0, 1)

	return genBadge(label, value, getColorForStatus(v))
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

// getColorForStatus generates color from green to red
func getColorForStatus(p float64) string {
	h := easing.CircIn(p, 0, 0.287, 1.0)
	k := color.HSV{H: h, S: 0.916, V: 0.80}
	return k.ToRGB().ToHex().ToWeb(true, true)
}

// parsePath parses request path
func parsePath(path string) (string, uint8) {
	token := strutil.ReadField(path, 0, false, '/')
	badge := strutil.ReadField(path, 1, false, '/')

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
