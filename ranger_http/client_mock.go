package ranger_http

import (
	"net/http"
)

// ApiClientMock allows mocking the apiClient struct
// Example usage to mock the response of a method:
// newClient := ApiClientMock{
//   GetMock: func (url string) (resp *http.Response, err error) { return &http.Response{}, errors.New("boom")},
//   DoMock: func (req *http.Request) (*http.Response, error) { return &http.Response{}, nil },
//   GetContentByURLMock: func (method string, url string, header http.Header) ([]byte, error) { return []byte{}, nil },
//   HeadMock: func (url string) (resp *http.Response, err error) { return },
// }
// _, err := newClient.Get("http://foo.com") -> boom
type ApiClientMock struct {
	GetMock             func(url string) (resp *http.Response, err error)
	DoMock              func(req *http.Request) (*http.Response, error)
	GetContentByURLMock func(method string, url string, header http.Header) ([]byte, error)
	HeadMock            func(url string) (resp *http.Response, err error)
}

func (d *ApiClientMock) Get(url string) (resp *http.Response, err error) {
	if d.GetMock != nil {
		return d.GetMock(url)
	}

	return &http.Response{}, nil
}

func (d *ApiClientMock) Do(req *http.Request) (*http.Response, error) {
	if d.DoMock != nil {
		return d.DoMock(req)
	}

	return &http.Response{}, nil
}

func (d *ApiClientMock) GetContentByURL(method string, url string, header http.Header) ([]byte, error) {
	if d.GetContentByURLMock != nil {
		return d.GetContentByURLMock(method, url, header)
	}

	return []byte{}, nil
}

func (d *ApiClientMock) Head(url string) (resp *http.Response, err error) {
	if d.HeadMock != nil {
		return d.HeadMock(url)
	}

	return &http.Response{}, nil
}
