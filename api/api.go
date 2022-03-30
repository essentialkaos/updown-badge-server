package api

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2022 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"github.com/valyala/fasthttp"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// API is API URL
const API_URL = "https://updown.io/api/checks"

// ////////////////////////////////////////////////////////////////////////////////// //

// Status contains info about check
type Status struct {
	Uptime float64 `json:"uptime"`
	IsDown bool    `json:"down"`
}

// Apdex contains apdex (Application Performance Index) info
type Apdex struct {
	Value float64 `json:"apdex"`
}

// API is API client struct
type API struct {
	client *fasthttp.Client
	key    string
}

// ////////////////////////////////////////////////////////////////////////////////// //

func NewClient(apiKey string) *API {
	return &API{
		key: apiKey,
		client: &fasthttp.Client{
			MaxIdleConnDuration: 5 * time.Second,
			ReadTimeout:         3 * time.Second,
			WriteTimeout:        3 * time.Second,
			MaxConnsPerHost:     150,
		},
	}
}

// ////////////////////////////////////////////////////////////////////////////////// //

// SetUserAgent set user-agent string based on app name and version
func (api *API) SetUserAgent(app, version string) {
	api.client.Name = formatUserAgent(app, version)
}

// GetStatus fetches status info from updown API
func (api *API) GetStatus(token string) (*Status, error) {
	data, err := api.doRequest(API_URL + "/" + token)

	if err != nil {
		return nil, err
	}

	status := &Status{}
	err = json.Unmarshal(data, status)

	if err != nil {
		return nil, fmt.Errorf("Can't decode status info: %v", err)
	}

	return status, nil
}

// GetApdex fetches apdex info from updown API
func (api *API) GetApdex(token string) (*Apdex, error) {
	data, err := api.doRequest(API_URL + "/" + token + "/metrics")

	if err != nil {
		return nil, err
	}

	apdex := &Apdex{}
	err = json.Unmarshal(data, apdex)

	if err != nil {
		return nil, fmt.Errorf("Can't decode apdex info: %v", err)
	}

	return apdex, nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// doRequest sends request to updown API
func (api *API) doRequest(url string) ([]byte, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url)
	req.URI().SetQueryString("api-key=" + api.key)

	err := api.client.Do(req, resp)

	if err != nil {
		return nil, err
	}

	statusCode := resp.StatusCode()

	if statusCode != 200 {
		return nil, fmt.Errorf("Server return status code %d", statusCode)
	}

	return resp.Body(), nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// formatUserAgent generate user-agent string for client
func formatUserAgent(app, version string) string {
	return fmt.Sprintf(
		"%s/%s (go; %s; %s-%s)",
		app, version, runtime.Version(),
		runtime.GOARCH, runtime.GOOS,
	)
}
