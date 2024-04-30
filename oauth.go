package woocommerce

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// OAuth represents the OAuth1.0a client
type OAuth struct {
	URL            string
	ConsumerKey    string
	ConsumerSecret string
	Version        string
	Method         string
	Timestamp      int64
}

// NewOAuth creates a new instance of the OAuth1.0a client
func NewOAuth(url, consumerKey, consumerSecret, version, method string, timestamp int64) *OAuth {
	return &OAuth{
		URL:            url,
		ConsumerKey:    consumerKey,
		ConsumerSecret: consumerSecret,
		Version:        version,
		Method:         method,
		Timestamp:      timestamp,
	}
}

func (o *OAuth) GetOAuthURL() string {
	params := make(map[string]string)

	// Parse existing query parameters
	urlParts, err := url.Parse(o.URL)
	if err != nil {
		return ""
	}
	for key, value := range urlParts.Query() {
		params[key] = value[0]
	}

	baseURL := strings.Split(o.URL, "?")[0]

	params["oauth_consumer_key"] = o.ConsumerKey
	params["oauth_timestamp"] = strconv.FormatInt(o.Timestamp, 10)
	params["oauth_nonce"] = o.generateNonce()
	params["oauth_signature_method"] = "HMAC-SHA256"
	params["oauth_signature"] = o.generateOAuthSignature(params, baseURL)

	query := url.Values{}
	for key, value := range params {
		query.Set(key, value)
	}
	queryString := query.Encode()

	return fmt.Sprintf("%s?%s", baseURL, queryString)
}

func (o *OAuth) generateOAuthSignature(params map[string]string, baseURL string) string {
	// Remove existing signature
	delete(params, "oauth_signature")

	// Sort parameters
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Create parameter string
	paramStrings := make([]string, len(keys))
	for i, key := range keys {
		paramStrings[i] = fmt.Sprintf("%s=%s", url.QueryEscape(key), url.QueryEscape(params[key]))
	}
	paramString := strings.Join(paramStrings, "&")

	// Create base string
	baseString := fmt.Sprintf("%s&%s&%s", o.Method, url.QueryEscape(baseURL), url.QueryEscape(paramString))

	// Create signing key
	signingKey := []byte(o.ConsumerSecret)
	if o.Version != "v1" && o.Version != "v2" {
		signingKey = append(signingKey, '&')
	}

	// Calculate signature
	hash := hmac.New(sha256.New, signingKey)
	hash.Write([]byte(baseString))
	signature := hash.Sum(nil)

	// Encode and return
	return base64.StdEncoding.EncodeToString(signature)
}

func (o *OAuth) generateNonce() string {
	rand.Seed(time.Now().UnixNano())
	nonce := make([]byte, 32)
	for i := range nonce {
		nonce[i] = byte(rand.Intn(10))
	}
	return fmt.Sprintf("%x", nonce)
}
