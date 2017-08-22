package ranger_logger

import (
	"fmt"
	"net"

	"github.com/bshuster-repo/logrus-logstash-hook"
	"github.com/johntdyer/slackrus"
	"github.com/sirupsen/logrus"
)

// NewLogstashHook ...
func NewLogstashHook(protocol string, addr string, formatter Formatter) Hook {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		fmt.Printf("Unable to connect to logstash: %s \n", err.Error())
		return nil
	}
	return logrustash.New(conn, formatter)
}

// NewSlackHook constructor
func NewSlackHook(channel, webhook, logLevel string) Hook {
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.DebugLevel
	}
	return &slackrus.SlackrusHook{
		HookURL:        webhook,
		AcceptedLevels: slackrus.LevelThreshold(level),
		Channel:        channel,
	}
}
