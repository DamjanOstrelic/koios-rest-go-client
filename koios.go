// Copyright 2022 The Howijd.Network Authors
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//   http://www.apache.org/licenses/LICENSE-2.0
//   or LICENSE file in repository root.
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package koios provides api client library to interact with Koios API
// endpoints and Cardano Blockchain. Sub package ./cmd/koios-rest
// provides cli application.
//
// Koios is best described as a Decentralized and Elastic RESTful query layer
// for exploring data on Cardano blockchain to consume within
// applications/wallets/explorers/etc.
package koios

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// MainnetHost       : is primay and default api host.
// GuildHost         : is Guild network host.
// TestnetHost       : is api host for testnet.
// DefaultAPIVersion : is openapi spec version e.g. /v0.
// DefaultPort       : default port used by api client.
// DefaultSchema     : default schema used by api client.
// LibraryVersion    : koios go library version.
// DefaultRateLimit  : is default rate limit used by api client.
// DefaultOrigin     : is default origin header used by api client.
const (
	MainnetHost              = "api.koios.rest"
	GuildHost                = "guild.koios.rest"
	TestnetHost              = "testnet.koios.rest"
	DefaultAPIVersion        = "v0"
	DefaultPort       uint16 = 443
	DefaultSchema            = "https"
	LibraryVersion           = "v0"
	DefaultRateLimit  uint8  = 5
	DefaultOrigin            = "https://github.com/howijd/koios-rest-go-client"
)

// Predefined errors used by the library.
var (
	ErrURLValuesLenght          = errors.New("if presenent then only single url.Values should be provided")
	ErrHTTPClientTimeoutSetting = errors.New("http.Client.Timeout should never be 0 in production")
	ErrHTTPClientChange         = errors.New("http.Client can only be set as option to koios.New")
	ErrOriginSet                = errors.New("origin can only be set as option to koios.New")
	ErrRateLimitRange           = errors.New("rate limit must be between 1-255 requests per sec")
	ErrResponseIsNotJSON        = errors.New("go non json response")
	ErrNoTxHash                 = errors.New("missing transaxtion hash(es)")
	ErrNoAddress                = errors.New("missing address")
	ErrNoPoolID                 = errors.New("missing pool id")
)

type (
	// Client is api client instance.
	Client struct {
		mux             sync.RWMutex
		host            string
		version         string
		port            uint16
		schema          string
		origin          string
		url             *url.URL
		client          *http.Client
		commonHeaders   http.Header
		reqInterval     time.Duration
		lastRequest     time.Time
		totalReq        uint64
		reqStatsEnabled bool
	}

	// Option is callback function which can be implemented
	// to change configurations options of API Client.
	Option func(*Client) error

	// Address defines type for _address.
	Address string

	// PaymentCredential.
	PaymentCredential string

	// AssetName defines type for _asset_name.
	AssetName string

	// BlockHash defines type for _block_hash.
	BlockHash string

	// TxHash defines type for tx_hash.
	TxHash string

	// EpochNo defines type for _epoch_no.
	EpochNo uint64

	// PoolID.
	PoolID string

	// PolicyID.
	PolicyID string

	// ScriptHash defines type for _script_hash.
	ScriptHash string

	// StakeAddress is Cardano staking address (reward account, bech32 encoded).
	StakeAddress string

	// Lovelace defines type for ADA lovelaces. This library uses forked snapshot
	// of github.com/shopspring/decimal package to provide. JSON and XML
	// serialization/deserialization and make it ease to work with calculations
	// and deciimal precisions of ADA lovelace and native assets.
	//
	// For API of decimal package see
	// https://pkg.go.dev/github.com/shopspring/decimal
	//
	// SEE: https://github.com/howijd/decimal
	// issues and bug reports are welcome to:
	// https://github.com/howijd/decimal/issues.
	Lovelace struct {
		decimal.Decimal
	}

	// PaymentAddr info.
	PaymentAddr struct {
		// Bech32 is Cardano payment/base address (bech32 encoded)
		// for transaction's or change to be returned.
		Bech32 string `json:"bech32"`

		// Payment credential.
		Cred PaymentCredential `json:"cred"`
	}

	// Certificate information.
	Certificate struct {
		// Index of the certificate
		Index int `json:"index"`

		// Info is A JSON object containing information from the certificate.
		Info map[string]interface{} `json:"info"`

		// Type of certificate could be:
		// delegation, stake_registration, stake_deregistraion, pool_update,
		// pool_retire, param_proposal, reserve_MIR, treasury_MIR).
		Type string `json:"type"`
	}

	// Response wraps API responses.
	Response struct {
		// RequestURL is full request url.
		RequestURL string `json:"request_url"`

		// RequestMethod is HTTP method used for request.
		RequestMethod string `json:"request_method"`

		// StatusCode of the HTTP response.
		StatusCode int `json:"status_code"`

		// Status of the HTTP response header if present.
		Status string `json:"status"`

		// Date response header.
		Date string `json:"date,omitempty"`

		// ContentLocation response header if present.
		ContentLocation string `json:"content_location,omitempty"`

		// ContentRange response header if present.
		ContentRange string `json:"content_range,omitempty"`

		// Error response body if present.
		Error *ResponseError `json:"error,omitempty"`

		// Stats of the request if stats are enabled.
		Stats *RequestStats `json:"stats,omitempty"`
	}

	// RequestStats represent collected request stats if collecting
	// request stats is enabled.
	RequestStats struct {
		// ReqStartedAt time when request was started.
		ReqStartedAt time.Time `json:"req_started_at,omitempty"`

		// DNSLookupDur DNS lookup duration.
		DNSLookupDur time.Duration `json:"req_dns_lookup_dur,omitempty"`

		// TLSHSDur time it took to perform TLS handshake.
		TLSHSDur time.Duration `json:"tls_hs_dur,omitempty"`

		// ESTCXNDur time it took to establish connection.
		ESTCXNDur time.Duration `json:"est_cxn_dur,omitempty"`

		// TTFB time it took to get the first byte of the response
		// after connextion was established.
		TTFB time.Duration `json:"ttfb,omitempty"`

		// ReqDur total time it took to peform the request.
		ReqDur time.Duration `json:"req_dur,omitempty"`

		// ReqDurStr String representation of ReqDur.
		ReqDurStr string `json:"req_dur_str,omitempty"`
	}

	// ResponseError represents api error messages.
	ResponseError struct {
		// Hint of the error reported by server.
		Hint string `json:"hint,omitempty"`

		// Details of the error reported by server.
		Details string `json:"details,omitempty"`

		// Code is error code reported by server.
		Code string `json:"code,omitempty"`

		// Message is error message reported by server.
		Message string `json:"message,omitempty"`
	}
)

// New creates thread-safe API client you can freerly pass this
// client to multiple go routines.
//
// Call to New without options is same as call with default options.
// e.g.
// api, err := koios.New(
// 	koios.Host(koios.MainnetHost),
// 	koios.APIVersion(koios.DefaultAPIVersion),
// 	koios.Port(koios.DefaultPort),
// 	koios.Schema(koios.DefaultSchema),
// 	koios.HttpClient(koios.DefaultHttpClient),
// ).
func New(opts ...Option) (*Client, error) {
	c := &Client{
		host:          MainnetHost,
		version:       DefaultAPIVersion,
		port:          DefaultPort,
		schema:        DefaultSchema,
		commonHeaders: make(http.Header),
	}
	// set default base url
	_ = c.updateBaseURL()
	// set default rate limit for outgoing requests.
	_ = RateLimit(DefaultRateLimit)(c)

	// set default common headers
	c.commonHeaders.Set("Accept", "application/json")
	c.commonHeaders.Set("Accept-Encoding", "gzip, deflate")
	c.commonHeaders.Set(
		"User-Agent",
		fmt.Sprintf(
			"go-koios/%s (%s %s) %s/%s https://github.com/howijd/go-koios",
			LibraryVersion,
			cases.Title(language.English).String(runtime.GOOS),
			runtime.GOARCH,
			runtime.GOOS,
			runtime.GOARCH,
		),
	)

	// Apply provided options
	for _, setOpt := range opts {
		if err := setOpt(c); err != nil {
			return nil, err
		}
	}

	// Sets default origin if option was not provided.
	_ = Origin(DefaultOrigin)(c)

	// If HttpClient option was not provided
	// use default http.Client
	if c.client == nil {
		if err := HTTPClient(nil)(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// Host returns option apply func which can be used to change the
// baseurl hostname https://<host>/api/v0/
func Host(host string) Option {
	return func(c *Client) error {
		c.mux.Lock()
		c.host = host
		c.mux.Unlock()
		return c.updateBaseURL()
	}
}

// APIVersion returns option apply func which can be used to change the
// baseurl api version https://api.koios.rest/api/<version>/
func APIVersion(version string) Option {
	return func(c *Client) error {
		c.mux.Lock()
		c.version = version
		c.mux.Unlock()
		return c.updateBaseURL()
	}
}

// Port returns option apply func which can be used to change the
// baseurl port https://api.koios.rest:<port>/api/v0/
func Port(port uint16) Option {
	return func(c *Client) error {
		c.mux.Lock()
		c.port = port
		c.mux.Unlock()
		return c.updateBaseURL()
	}
}

// Schema returns option apply func which can be used to change the
// baseurl schema <schema>://api.koios.rest/api/v0/.
func Schema(schema string) Option {
	return func(c *Client) error {
		c.mux.Lock()
		c.schema = schema
		c.mux.Unlock()
		return c.updateBaseURL()
	}
}

// HTTPClient enables to set htt.Client to be used for requests.
// http.Client can only be set once.
func HTTPClient(client *http.Client) Option {
	return func(c *Client) error {
		if c.client != nil {
			return ErrHTTPClientChange
		}
		c.mux.Lock()
		defer c.mux.Unlock()
		if client == nil {
			client = &http.Client{
				Timeout: time.Second * 60,
			}
		}
		if client.Timeout == 0 {
			return ErrHTTPClientTimeoutSetting
		}
		c.client = client
		if c.client.Transport == nil {
			c.client.Transport = http.DefaultTransport
		}
		return nil
	}
}

// RateLimit sets requests per second this client is allowed to create
// and effectievely rate limits outgoing requests.
// Let's respect usage of the community provided resources.
func RateLimit(reqps uint8) Option {
	return func(c *Client) error {
		if reqps == 0 {
			return ErrRateLimitRange
		}
		c.mux.Lock()
		defer c.mux.Unlock()
		c.reqInterval = time.Second / time.Duration(reqps)
		return nil
	}
}

// Origin sets Origin header for outgoing api requests.
// Recomoended is to set it to URL or FQDN of your project using this library.
//
// In case you appliation goes rouge this could help to keep api.koios.rest
// service stable and up and running while temporary limiting requests
// it accepts from your application.
//
// It's not required, but considered as good practice so that Cardano Community
// can provide HA services for Cardano ecosystem.
func Origin(origin string) Option {
	return func(c *Client) error {
		u, err := url.ParseRequestURI(origin)
		if err != nil {
			return err
		}

		c.mux.Lock()
		defer c.mux.Unlock()
		c.origin = u.String()
		return nil
	}
}

// CollectRequestsStats when enabled uses httptrace is used
// to collect detailed timing information about the request.
func CollectRequestsStats(enabled bool) Option {
	return func(c *Client) error {
		c.mux.Lock()
		defer c.mux.Unlock()
		c.reqStatsEnabled = enabled
		return nil
	}
}

func readResponseBody(rsp *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}
	if !strings.Contains(rsp.Header.Get("Content-Type"), "json") {
		return nil, fmt.Errorf("%w: %s", ErrResponseIsNotJSON, string(body))
	}
	return body, nil
}

func (r *Response) applyError(body []byte, err error) {
	r.Error = &ResponseError{}
	_ = json.Unmarshal(body, r.Error)
	if err != nil && len(r.Error.Message) == 0 {
		r.Error.Message = err.Error()
	}
	r.ready()
}

func (r *Response) ready() {
	if r.Stats == nil {
		return
	}
	r.Stats.ReqDur = time.Since(r.Stats.ReqStartedAt)
	r.Stats.ReqDurStr = fmt.Sprint(r.Stats.ReqDur)
}

func (r *Response) applyRsp(rsp *http.Response) {
	r.StatusCode = rsp.StatusCode
	r.RequestMethod = rsp.Request.Method
	r.Status = rsp.Status
	r.Date = rsp.Header.Get("date")
	r.ContentRange = rsp.Header.Get("content-range")
	r.ContentLocation = rsp.Header.Get("content-location")
}
