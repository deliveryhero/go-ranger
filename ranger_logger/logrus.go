package ranger_logger

import (
	"net"

	logrustash "github.com/bshuster-repo/logrus-logstash-hook"
	"github.com/sirupsen/logrus"
)

type LoggerData map[string]interface{}

//LoggerInterface ...
type LoggerInterface interface {
	Info(message string, data LoggerData)
	Warning(message string, data LoggerData)
	Error(message string, data LoggerData)
	Panic(message string, data LoggerData)
}

//Wrapper - Wrap a logrus logger
type Wrapper struct {
	*logrus.Logger // see promoted methods https://www.goinggo.net/2015/09/composition-with-go.html
}

//NewLoggerWithLogstashHook - LoggerWrapper constructor with logstash hook
func NewLoggerWithLogstashHook(protocol string, addr string, appName string) LoggerInterface {
	log := logrus.New()

	if conn, err := net.Dial(protocol, addr); err == nil {
		hook := logrustash.New(conn, &logrus.JSONFormatter{})
		log.Hooks.Add(hook)
	} else {
		log.Warn("unable to connect to logstash")
	}

	return &Wrapper{log}
}

//CreateFieldsFromRequest - Create a logrus.Fields object from a Request
/*func (logger *Wrapper) CreateFieldsFromRequest(r *http.Request) LoggerData {
	return logrus.Fields{
		"client_ip":      r.RemoteAddr,
		"request_method": r.Method,
		"request_uri":    r.RequestURI,
		"request_host":   r.Host,
	}
}*/

//Info - Wrap Info from logrus logger
func (logger *Wrapper) Info(message string, data LoggerData) {
	ctx := logger.WithFields(convertToLogrusFields(data))

	ctx.Info(message)
}

//Warning - Wrap Warning from logrus logger
func (logger *Wrapper) Warning(message string, data LoggerData) {
	ctx := logger.WithFields(convertToLogrusFields(data))

	ctx.Warning(message)
}

//Error - Wrap Error from logrus logger
func (logger *Wrapper) Error(message string, data LoggerData) {
	ctx := logger.WithFields(convertToLogrusFields(data))

	ctx.Error(message)
}

//Panic - Wrap Panic from logrus logger
func (logger *Wrapper) Panic(message string, data LoggerData) {
	ctx := logger.WithFields(convertToLogrusFields(data))

	ctx.Panic(message)
}

func convertToLogrusFields(loggerData LoggerData) logrus.Fields {
	fields := logrus.Fields{}
	for k, v := range loggerData {
		fields[k] = v
	}
	return fields
}
