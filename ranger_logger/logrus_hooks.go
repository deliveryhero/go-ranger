package ranger_logger

import (
	"fmt"
	"net"

	"github.com/bshuster-repo/logrus-logstash-hook"
	"github.com/johntdyer/slackrus"
	"github.com/sirupsen/logrus"
)

// NewLogstashHook ...
func NewLogstashHook(protocol string, addr string, f Formatter) Hook {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		fmt.Printf("Unable to connect to logstash: %s \n", err.Error())
		return nil
	}
	return logrustash.New(conn, f)
}

// NewSlackHook constructor
func NewSlackHook(channel string, webhook string, logNotificationLevel string) Hook {
	logLevel, err := logrus.ParseLevel(logNotificationLevel)
	if err != nil {
		logLevel = logrus.DebugLevel
	}
	return &slackrus.SlackrusHook{
		HookURL:        webhook,
		AcceptedLevels: slackrus.LevelThreshold(logLevel),
		Channel:        channel,
	}
}
