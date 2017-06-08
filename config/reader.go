package config

import (
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/fesposito/go-ranger/http"
)

const (
	defaultTimeout = 5
)

// @todo parse yaml filo into this configs
type config struct {
	AppName           string
	APIRequestTimeout int
	HTTPAddress       string
	Version           string
}

// Reader is the interface for config readers
type Reader interface {
	ReadConfig() (interface{}, error)
	GetConfigPath() string
}

// apiConfigReader is the config reader implementation using an api as source
type remoteConfigReader struct {
	url    string
	client http.APIClientInterface
}

// localConfigReader reader
type localConfigReader struct {
	configPath string
	readFile   func(filename string) ([]byte, error)
}

// newAPIConfigReader is the factory for config readers.
func newRemoteConfigReader(apiClient http.APIClientInterface, configPath string) Reader {
	return &remoteConfigReader{
		url:    configPath,
		client: apiClient,
	}
}

// ReadConfig fetches the config for the app remotely
func (configReader *remoteConfigReader) ReadConfig() (interface{}, error) {
	// @todo define data structure and implement remoteConfigReader
	return nil, nil
}

// GetConfigPath ...
func (configReader *remoteConfigReader) GetConfigPath() string {
	return configReader.url
}

// newLocalConfigReader is the factory for config readers.
func newLocalConfigReader(configPath string) Reader {
	return &localConfigReader{
		configPath: configPath,
		readFile:   ioutil.ReadFile,
	}
}

// ReadConfig fetches the config for the app locally
func (configReader *localConfigReader) ReadConfig() (interface{}, error) {
	// @todo define data structure and implement localConfigReader
	return nil, nil
}

// GetConfigPath ...
func (configReader *localConfigReader) GetConfigPath() string {
	return getLocalPath(configReader.configPath)
}

// GetConfigReader strategy
func GetConfigReader(path string) Reader {
	if isReadConfigurationLocal(path) {
		return newLocalConfigReader(getLocalPath(path))
	}

	return newRemoteConfigReader(http.NewAPIClient(defaultTimeout), path)
}

func isReadConfigurationLocal(path string) bool {
	u, err := url.Parse(path)
	if err != nil {
		panic(err)
	}
	return u.Scheme == "file"
}

func getLocalPath(path string) string {
	u, err := url.Parse(path)
	if err != nil {
		panic(err)
	}
	return strings.Replace(u.String(), "file://", "", -1)
}
