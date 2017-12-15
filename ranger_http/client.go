package ranger_http

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// APIClientInterface is an interface for api clients. It allows us to mock the basic http client.
type APIClientInterface interface {
	Get(url string) (resp *http.Response, err error)
	Do(req *http.Request) (*http.Response, error)
	GetContentByURL(method string, url string, header http.Header) ([]byte, error)
	Head(url string) (resp *http.Response, err error)
}

type apiClient struct {
	client *http.Client
}

// NewAPIClient is the factory method for api clients.
func NewAPIClient(requestTimeout int) APIClientInterface {
	return &apiClient{
		client: &http.Client{
			Timeout: time.Second * time.Duration(requestTimeout),
		},
	}
}

// Get is issueing a GET request to the given url
func (client *apiClient) Get(url string) (*http.Response, error) {
	res, err := client.client.Get(url)
	if err != nil {
		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf(
				"ApiClient.Get=Bad request,StatusCode=%d, URL=%s, Header: %+v", res.StatusCode, url, res.Header,
			)
		}
	}

	return res, err
}

// Do sends an HTTP request and returns an HTTP response, following
// policy (such as redirects, cookies, auth) as configured on the
// client.
func (client *apiClient) Do(req *http.Request) (*http.Response, error) {
	res, err := client.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(
			"ApiClient.Do=Cannot execute request, URL=%s, Header=%+v", req.URL, req.Header,
		)
	}

	return res, err
}

// GetContentByURL execute a GET request and return
func (client *apiClient) GetContentByURL(method string, url string, header http.Header) ([]byte, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	if header != nil {
		req.Header = header
	}

	res, err := client.Do(req)
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
func (client *apiClient) Head(url string) (resp *http.Response, err error) {
	response, err := client.client.Head(url)

	if err != nil {
		return nil, err
	}

	return response, nil
}
