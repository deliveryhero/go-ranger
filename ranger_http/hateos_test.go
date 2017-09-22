package ranger_http

import (
	"net/http"
	"net/url"
	"testing"
)

func TestGetHeaderByKey(t *testing.T) {

	testCases := []map[string]string{
		{
			"key":   "Cool-Header",
			"value": "coolValue",
		},
		{
			"key":   "Cool-Header",
			"value": "",
		},
	}

	for _, testCase := range testCases {
		testClient := ApiClientMock{
			GetMock:             func(url string) (resp *http.Response, err error) { return &http.Response{}, nil },
			DoMock:              func(req *http.Request) (*http.Response, error) { return &http.Response{}, nil },
			GetContentByURLMock: func(method string, url string, header http.Header) ([]byte, error) { return []byte{}, nil },
			HeadMock: func(url string) (resp *http.Response, err error) {

				response := &http.Response{}

				if testCase["value"] != "" {
					header := &http.Header{}
					header.Add(testCase["key"], testCase["value"])
					response.Header = *header
				}

				return response, nil
			},
		}

		result, _ := GetHeaderByKey("http://dummy.url", testCase["key"], &testClient)

		if result != testCase["value"] {
			t.Errorf("Unexpected value \"%s\". The expeted value was \"%s\"", result, testCase["value"])
		}
	}
}

func TestGetLastPageFromLinksHeader(t *testing.T) {

	testCases := []map[string]interface{}{
		{ // regular response
			"expected": 10,
			"linksHeader": "<https://webtranslateit.com/api/projects/proj_pvt_dPvsfFzNREHB5rhp9B7V8Q/strings.json?page=2>; rel=\"next\", " +
				"<https://webtranslateit.com/api/projects/proj_pvt_dPvsfFzNREHB5rhp9B7V8Q/strings.json?page=10>; rel=\"last\", " +
				"<https://webtranslateit.com/api/projects/proj_pvt_dPvsfFzNREHB5rhp9B7V8Q/strings.json?page=1>; rel=\"first\"",
		},
		{ // switch places
			"expected": 4,
			"linksHeader": "<https://webtranslateit.com/api/projects/proj_pvt_dPvsfFzNREHB5rhp9B7V8Q/strings.json?page=2>; rel=\"next\", " +
				"<https://webtranslateit.com/api/projects/proj_pvt_dPvsfFzNREHB5rhp9B7V8Q/strings.json?page=1>; rel=\"first\", " +
				"<https://webtranslateit.com/api/projects/proj_pvt_dPvsfFzNREHB5rhp9B7V8Q/strings.json?page=4>; rel=\"last\"",
		},
		{ // wrong data
			"expected": 0,
			"linksHeader": "<https://webtranslateit.com/api/projects/proj_pvt_dPvsfFzNREHB5rhp9B7V8Q/strings.json?page=2>; rel=\"next\", " +
				"<https://webtranslateit.com/api/projects/proj_pvt_dPvsfFzNREHB5rhp9B7V8Q/strings.json?page=1>; rel=\"first\", " +
				"<https://webtranslateit.com/api/projects/proj_pvt_dPvsfFzNREHB5rhp9B7V8Q/strings.json?page=WRONG_DATA>; rel=\"last\"",
		},
		{ // empty result
			"expected":    0,
			"linksHeader": "",
		},
	}

	for _, testCase := range testCases {

		linksHeader, ok := testCase["linksHeader"].(string)

		if !ok {
			t.Error("uerror while converting linksHeader to string!")
		}

		result, _ := GetLastPageFromLinksHeader(linksHeader)

		if result != testCase["expected"] {
			t.Errorf("unexpected result:\n%d\nthe expected was:\n%d\n", result, testCase["expected"])
		}

	}
}

func TestGetNumPagesForURL(t *testing.T) {

	testCases := []map[string]interface{}{
		{
			"linksHeader": "<https://webtranslateit.com/api/projects/proj_pvt_dPvsfFzNREHB5rhp9B7V8Q/strings.json?page=2>; rel=\"next\", " +
				"<https://webtranslateit.com/api/projects/proj_pvt_dPvsfFzNREHB5rhp9B7V8Q/strings.json?page=10>; rel=\"last\", " +
				"<https://webtranslateit.com/api/projects/proj_pvt_dPvsfFzNREHB5rhp9B7V8Q/strings.json?page=1>; rel=\"first\"",
			"expected": 10,
		},
		{
			"linksHeader": "",
			"expected":    1,
		},
	}

	for _, testCase := range testCases {

		linksHeader := testCase["linksHeader"].(string)

		testClient := ApiClientMock{
			GetMock:             func(url string) (resp *http.Response, err error) { return &http.Response{}, nil },
			DoMock:              func(req *http.Request) (*http.Response, error) { return &http.Response{}, nil },
			GetContentByURLMock: func(method string, url string, header http.Header) ([]byte, error) { return []byte{}, nil },
			HeadMock: func(url string) (resp *http.Response, err error) {
				response := &http.Response{}
				header := &http.Header{}
				header.Add("Link", linksHeader)
				response.Header = *header

				return response, nil
			},
		}

		testUrl, _ := url.Parse("http://google.com")
		result, err := GetNumPagesForURL(testUrl, &testClient)

		if err != nil && 1 != result {
			t.Error("Default result shoud be 1")
		}

		if result != testCase["expected"] {
			t.Errorf("Unexpected result \"%d\". The expected one was \"%d\".", result, testCase["expected"])
		}
	}
}
