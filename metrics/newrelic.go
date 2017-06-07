package metrics

import newrelic "github.com/newrelic/go-agent"

// InitNewRelic ...
func InitNewRelic(appName string, license string) (newrelic.Application, error) {
	return newrelic.NewApplication(newrelic.NewConfig(
		appName,
		license),
	)
}
