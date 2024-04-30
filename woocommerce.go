package woocommerce

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// API represents the WooCommerce API client
type API struct {
	URL             string
	ConsumerKey     string
	ConsumerSecret  string
	WPAPI           bool
	Version         string
	IsSSL           bool
	Timeout         time.Duration
	VerifySSL       bool
	QueryStringAuth bool
	UserAgent       string
}

// NewAPI creates a new instance of the WooCommerce API client
func NewAPI(url, consumerKey, consumerSecret string) *API {
	return &API{
		URL:            url,
		ConsumerKey:    consumerKey,
		ConsumerSecret: consumerSecret,
		WPAPI:          true,
		Version:        "wc/v3",
		Timeout:        5 * time.Second,
		VerifySSL:      true,
		UserAgent:      fmt.Sprintf("WooCommerce-Go-REST-API/%s", "3.0.0"),
	}
}

func (api *API) isSSL() bool {
	return strings.HasPrefix(api.URL, "https")
}

func (api *API) getURL(endpoint string) string {
	apiPath := "wc-api"
	if !api.WPAPI {
		apiPath = "wp-json"
	}
	if !strings.HasSuffix(api.URL, "/") {
		api.URL += "/"
	}
	return fmt.Sprintf("%s%s/%s/%s", api.URL, apiPath, api.Version, endpoint)
}

func (api *API) getOAuthURL(url, method string, oauthTimestamp int64) string {
	oauth := NewOAuth(url, api.ConsumerKey, api.ConsumerSecret, api.Version, method, oauthTimestamp)
	return oauth.GetOAuthURL()
}

func (api *API) request(method, endpoint string, data interface{}, params url.Values) (*http.Response, error) {
	url := api.getURL(endpoint)
	headers := make(map[string]string)
	headers["user-agent"] = api.UserAgent
	headers["accept"] = "application/json"

	var reqBody []byte
	if data != nil {
		reqBody, _ = json.Marshal(data)
		headers["content-type"] = "application/json;charset=utf-8"
	}

	var req *http.Request
	var err error
	if api.IsSSL && !api.QueryStringAuth {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(reqBody))
		if err != nil {
			return nil, err
		}
		req.SetBasicAuth(api.ConsumerKey, api.ConsumerSecret)
	} else if api.IsSSL && api.QueryStringAuth {
		params.Set("consumer_key", api.ConsumerKey)
		params.Set("consumer_secret", api.ConsumerSecret)
		url += "?" + params.Encode()
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return nil, err
		}
	} else {
		url += "?" + params.Encode()
		url = api.getOAuthURL(url, method, time.Now().Unix())
		req, err = http.NewRequest(method, url, bytes.NewBuffer(reqBody))
		if err != nil {
			return nil, err
		}
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{Timeout: api.Timeout}
	return client.Do(req)
}

func (api *API) Get(endpoint string, params url.Values) (*http.Response, error) {
	return api.request("GET", endpoint, nil, params)
}

func (api *API) Post(endpoint string, data interface{}, params url.Values) (*http.Response, error) {
	return api.request("POST", endpoint, data, params)
}

func (api *API) Put(endpoint string, data interface{}, params url.Values) (*http.Response, error) {
	return api.request("PUT", endpoint, data, params)
}

func (api *API) Delete(endpoint string, params url.Values) (*http.Response, error) {
	return api.request("DELETE", endpoint, nil, params)
}

func (api *API) Options(endpoint string, params url.Values) (*http.Response, error) {
	return api.request("OPTIONS", endpoint, nil, params)
}
