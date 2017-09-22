package ranger_http

import (
	"errors"
	"net/url"
	"strconv"

	"github.com/tomnomnom/linkheader"
)

// GetHeaderByKey retrieve header by specific key
func GetHeaderByKey(url string, key string, apiClient APIClientInterface) (string, error) {

	response, err := apiClient.Head(url)

	if err != nil {
		return "", err
	}

	header := response.Header.Get(key)

	if header == "" {
		return "", errors.New("header not found or it's empty")
	}

	return header, nil
}

// GetLastPageFromLinksHeader extract the last page from a Link header
func GetLastPageFromLinksHeader(linksHeader string) (int, error) {

	parsedLinks := linkheader.Parse(linksHeader)
	lastPage := 0

	if rel := parsedLinks.FilterByRel("last"); rel != nil && len(rel) > 0 {
		parsedURL, err := url.Parse(rel[0].URL)

		if err != nil {
			return 0, err
		}

		params := parsedURL.Query()

		if lastPage, err = strconv.Atoi(params["page"][0]); err != nil {
			return 0, err
		}
	} else {
		return 0, errors.New("rel property not found")
	}

	return lastPage, nil
}

// GetNumPagesForURL Obtain the number of pages for a specific URL
func GetNumPagesForURL(url *url.URL, apiClient APIClientInterface) (int, error) {

	linksHeader, _ := GetHeaderByKey(url.String(), "Link", apiClient)

	res, err := GetLastPageFromLinksHeader(linksHeader)

	if err != nil {
		return 1, err // there is no Link header or no info about paging, so we assume it's 1 page
	} else {
		return res, nil
	}
}
