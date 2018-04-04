package ranger_http

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/foodora/go-ranger/ranger_logger"
)

// APIClientInterface is an interface for api clients. It allows us to mock the basic http client.
type APIClientInterface interface {
	Get(url string) (resp *http.Response, err error)
	Do(req *http.Request) (*http.Response, error)
	Head(url string) (resp *http.Response, err error)
}

// APIClient ...
type APIClient struct {
	*http.Client
	ranger_logger.LoggerInterface
}

// NewAPIClient is the factory method for api clients.
func NewAPIClient(requestTimeout int) *APIClient {
	return &APIClient{
		Client: &http.Client{
			Timeout: time.Second * time.Duration(requestTimeout),
		},
	}
}

// Get is issueing a GET request to the given url
func (c *APIClient) Get(url string) (*http.Response, error) {
	c.Debug("ApiClient.Get", ranger_logger.LoggerData{"url": url})
	res, err := c.Client.Get(url)
	if err != nil && res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"ApiClient.Get=Bad request,StatusCode=%d, URL=%s, Header: %+v", res.StatusCode, url, res.Header,
		)
	}

	return res, err
}

// Do sends an HTTP request and returns an HTTP response, following
// policy (such as redirects, cookies, auth) as configured on the
// client.
func (c *APIClient) Do(req *http.Request) (*http.Response, error) {
	c.Debug("ApiClient.Do", ranger_logger.CreateFieldsFromRequest(req))
	res, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(
			"ApiClient.Do=Cannot execute request, URL=%s, Header=%+v", req.URL, req.Header,
		)
	}

	return res, err
}

// GetContentByURL execute a GET request and return
func (c *APIClient) GetContentByURL(method string, url string, header http.Header) ([]byte, error) {
	req, err := http.NewRequest(method, url, nil)
	c.Debug("ApiClient.GetContentByURL", ranger_logger.CreateFieldsFromRequest(req))
	if err != nil {
		return nil, err
	}

	if header != nil {
		req.Header = header
	}

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// Head issues a HEAD to the specified URL.
func (c *APIClient) Head(url string) (resp *http.Response, err error) {
	response, err := c.Client.Head(url)

	if err != nil {
		return nil, err
	}

	return response, nil
}
