package client

import (
	"crypto/tls"
	"github.com/olivere/elastic/v7"
	"monitor/config"
	"net/http"
	"net/url"
	"time"
)

type BasicAuthTransport struct {
	username string
	password string
}

// RoundTrip implements the RoundTripper interface
func (tr *BasicAuthTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.SetBasicAuth(tr.username, tr.password)
	tp := http.DefaultTransport
	tp.(*http.Transport).MaxIdleConnsPerHost = 100
	tp.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return tp.RoundTrip(r)
}

// newBearerTransport returns a new Transport with basic auth
func newBasicTransport(username, password string) *BasicAuthTransport {
	return &BasicAuthTransport{username: username, password: password}
}

// NewClient returns a new elasticsearch client instance
func NewEsClient(config config.ESConfig) (*ESClient, error) {
	httpClient, err := newHttpClient(config)
	if err != nil {
		return nil, err
	}

	clusterURL, err := url.Parse(config.URL)
	if err != nil {
		return nil, err
	}

	client, err := elastic.NewSimpleClient(
		elastic.SetHttpClient(httpClient),
		elastic.SetURL(config.URL),
		elastic.SetSniff(false),
		elastic.SetScheme(clusterURL.Scheme),
		elastic.SetHealthcheckTimeoutStartup(10*time.Second),
		elastic.SetHealthcheckTimeout(10*time.Second),
		// elastic.SetTraceLog(&debugLogger{}),
		// elastic.SetErrorLog(&errorLogger{}),
		// elastic.SetInfoLog(&infoLogger{}),
	)
	if err != nil {
		return nil, err
	}

	return &ESClient{
		Client: client,
		Mock:   config.Mock,
	}, nil
}

// newHttpClient returns a http client
func newHttpClient(config config.ESConfig) (*http.Client, error) {
	return &http.Client{
		Transport: newBasicTransport(config.Username, config.Password),
		Timeout:   time.Second * 10,
	}, nil
}

type ESClient struct {
	Client *elastic.Client
	Mock   bool
}
