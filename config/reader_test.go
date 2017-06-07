package config

import (
	"testing"

	"github.com/fesposito/go-ranger/http"
)

func TestNewRemoteConfigReader(t *testing.T) {
	configReader := newRemoteConfigReader(http.NewAPIClient(int64(5)), "http://url")

	if configReader.GetConfigPath() != "http://url" {
		t.Error("invalid url set to configReader")
	}
}

func TestNewLocalConfigReader(t *testing.T) {
	configReader := newLocalConfigReader("file:///my/path")

	if configReader.GetConfigPath() != "/my/path" {
		t.Error("invalid url set to configReader")
	}
}
