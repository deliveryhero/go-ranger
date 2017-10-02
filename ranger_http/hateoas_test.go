package ranger_http

import (
	"errors"
	"net/http"
	"net/url"
	"testing"
)

// TestGetHeaderByKey ...
func TestGetHeaderByKey(t *testing.T) {

	testCases := map[string]struct {
		key   string
		value string
		err   error
	}{
		"valid response": {
			key:   "Cool-Header",
			value: "coolValue",
		},
		"empty response": {
			key:   "Cool-Header",
			value: "",
			err:   errors.New("header not found or it's empty"),
		},
		"error response": {
			key:   "",
			value: "",
			err:   errors.New("header not found or it's empty"),
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			testClient := ApiClientMock{
				GetMock:             func(url string) (resp *http.Response, err error) { return &http.Response{}, nil },
				DoMock:              func(req *http.Request) (*http.Response, error) { return &http.Response{}, nil },
				GetContentByURLMock: func(method string, url string, header http.Header) ([]byte, error) { return []byte{}, nil },
				HeadMock: func(url string) (resp *http.Response, err error) {

					response := &http.Response{}

					if testCase.value != "" {
						header := &http.Header{}
						header.Add(testCase.key, testCase.value)
						response.Header = *header
					}

					return response, testCase.err
				},
			}

			result, err := GetHeaderByKey("http://dummy.url", testCase.key, &testClient)

			if result != testCase.value {
				t.Errorf("Unexpected value \"%s\". The expeted value was \"%s\"", result, testCase.value)
			}

			if err != testCase.err {
				t.Errorf("Unexpected error \"%v\". The expeted one was \"%v\"", err, testCase.err)
			}
		})
	}
}

// TestGetLastPageFromLinksHeader ...
func TestGetLastPageFromLinksHeader(t *testing.T) {

	testCases := map[string]struct {
		linksHeader   string
		expectedPages int
		err           error
	}{
		"response OK": {
			expectedPages: 10,
			linksHeader: "<https://webtranslateit.com/api/projects/proj_pvt_dPvsfFzNREHB5rhp9B7V8Q/strings.json?page=2>; rel=\"next\", " +
				"<https://webtranslateit.com/api/projects/proj_pvt_dPvsfFzNREHB5rhp9B7V8Q/strings.json?page=10>; rel=\"last\", " +
				"<https://webtranslateit.com/api/projects/proj_pvt_dPvsfFzNREHB5rhp9B7V8Q/strings.json?page=1>; rel=\"first\"",
		},
		"response OK - switch tags places": {
			expectedPages: 4,
			linksHeader: "<https://webtranslateit.com/api/projects/proj_pvt_dPvsfFzNREHB5rhp9B7V8Q/strings.json?page=2>; rel=\"next\", " +
				"<https://webtranslateit.com/api/projects/proj_pvt_dPvsfFzNREHB5rhp9B7V8Q/strings.json?page=1>; rel=\"first\", " +
				"<https://webtranslateit.com/api/projects/proj_pvt_dPvsfFzNREHB5rhp9B7V8Q/strings.json?page=4>; rel=\"last\"",
		},
		"wrong page param data": {
			expectedPages: 0,
			linksHeader: "<https://webtranslateit.com/api/projects/proj_pvt_dPvsfFzNREHB5rhp9B7V8Q/strings.json?page=2>; rel=\"next\", " +
				"<https://webtranslateit.com/api/projects/proj_pvt_dPvsfFzNREHB5rhp9B7V8Q/strings.json?page=1>; rel=\"first\", " +
				"<https://webtranslateit.com/api/projects/proj_pvt_dPvsfFzNREHB5rhp9B7V8Q/strings.json?page=WRONG_DATA>; rel=\"last\"",
			err: errors.New("page is empty or contains invalid data"),
		},
		"malformed url": {
			expectedPages: 0,
			linksHeader: "<https://webtranslateit.com/api/projects/proj_pvt_dPvsfFzNREHB5rhp9B7V8Q/strings.json?page=2>; rel=\"next\", " +
				"<https://webtranslateit.com/api/projects/proj_pvt_dPvsfFzNREHB5rhp9B7V8Q/strings.json?page=1>; rel=\"first\", " +
				"<^n0tVal!dURL%>; rel=\"last\"",
			err: errors.New("parse ^n0tVal!dURL%: invalid URL escape \"%\""),
		},
		"missing page param": {
			expectedPages: 0,
			linksHeader: "<https://webtranslateit.com/api/projects/proj_pvt_dPvsfFzNREHB5rhp9B7V8Q/strings.json>; rel=\"next\", " +
				"<https://webtranslateit.com/api/projects/proj_pvt_dPvsfFzNREHB5rhp9B7V8Q/strings.json>; rel=\"last\", " +
				"<https://webtranslateit.com/api/projects/proj_pvt_dPvsfFzNREHB5rhp9B7V8Q/strings.json>; rel=\"first\"",
			err: errors.New("page is empty or contains invalid data"),
		},
		"empty result": {
			expectedPages: 0,
			linksHeader:   "",
			err:           errors.New("rel property not found"),
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			result, err := GetLastPageFromLinksHeader(testCase.linksHeader)

			if result != testCase.expectedPages {
				t.Errorf("unexpected result:\n%d\nthe expected was:\n%d\n", result, testCase.expectedPages)
			}

			if testCase.err != nil && err.Error() != testCase.err.Error() {
				t.Errorf("unexpected error:\n%v\nthe expected was:\n%v\n", err, testCase.err)
			}
		})
	}
}

// TestGetNumPagesForURL ...
func TestGetNumPagesForURL(t *testing.T) {

	testCases := map[string]struct {
		linksHeader   string
		expectedPages int
	}{
		"non empty response": {
			linksHeader: "<https://webtranslateit.com/api/projects/proj_pvt_dPvsfFzNREHB5rhp9B7V8Q/strings.json?page=2>; rel=\"next\", " +
				"<https://webtranslateit.com/api/projects/proj_pvt_dPvsfFzNREHB5rhp9B7V8Q/strings.json?page=10>; rel=\"last\", " +
				"<https://webtranslateit.com/api/projects/proj_pvt_dPvsfFzNREHB5rhp9B7V8Q/strings.json?page=1>; rel=\"first\"",
			expectedPages: 10,
		},
		"empty response": {
			linksHeader:   "",
			expectedPages: 0,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			testClient := ApiClientMock{
				GetMock:             func(url string) (resp *http.Response, err error) { return &http.Response{}, nil },
				DoMock:              func(req *http.Request) (*http.Response, error) { return &http.Response{}, nil },
				GetContentByURLMock: func(method string, url string, header http.Header) ([]byte, error) { return []byte{}, nil },
				HeadMock: func(url string) (resp *http.Response, err error) {
					response := &http.Response{}
					header := &http.Header{}
					header.Add("Link", testCase.linksHeader)
					response.Header = *header

					return response, nil
				},
			}

			testURL, _ := url.Parse("http://google.com")
			result, err := GetNumPagesForURL(testURL, &testClient)

			if err != nil && 0 != result {
				t.Error("Default result shoud be 0")
			}

			if result != testCase.expectedPages {
				t.Errorf("Unexpected result \"%d\". The expected one was \"%d\".", result, testCase.expectedPages)
			}
		})
	}
}
