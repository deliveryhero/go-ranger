package ranger_config

import (
	"testing"

	"github.com/foodora/go-ranger/ranger_http"
)

func TestNewRemoteConfigReader(t *testing.T) {
	configReader := newRemoteConfigReader(ranger_http.NewAPIClient(5), "http://url")

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

func TestParseLocalConfig(t *testing.T) {
	t.Error("@todo TestParseLocalConfig")
}

func TestParseRemoteConfig(t *testing.T) {
	t.Error("@todo TestParseLocalConfig")
}

func TestParseConfig_InvalidConfig(t *testing.T) {
	t.Error("@todo TestParseConfig_InvalidConfig")
}
