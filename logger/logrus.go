package ranger_logger

import (
	"net"
	"net/http"

	logrustash "github.com/bshuster-repo/logrus-logstash-hook"
	"github.com/sirupsen/logrus"
)

//LoggerInterface ...
type LoggerInterface interface {
	Info(message string, data logrus.Fields)
	Warning(message string, data logrus.Fields)
	Error(message string, data logrus.Fields)
}

//Wrapper - Wrap a logrus logger
type Wrapper struct {
	*logrus.Logger // see promoted methods https://www.goinggo.net/2015/09/composition-with-go.html
}

//NewLoggerWithLogstashHook - LoggerWrapper constructor with logstash hook
func NewLoggerWithLogstashHook(protocol string, addr string, appName string, data logrus.Fields) LoggerInterface {
	log := logrus.New()

	conn, err := net.Dial(protocol, addr)
	if err != nil {
		log.Fatal(err)
	}

	hook := logrustash.New(conn, &logrus.JSONFormatter{})
	log.Hooks.Add(hook)

	return &Wrapper{log}
}

//CreateFieldsFromRequest - Create a logrus.Fields object from a Request
func (logger *Wrapper) CreateFieldsFromRequest(r *http.Request) logrus.Fields {
	return logrus.Fields{
		"client_ip":      r.RemoteAddr,
		"request_method": r.Method,
		"request_uri":    r.RequestURI,
		"request_host":   r.Host,
	}
}

//Info - Wrap Info from logrus logger
func (logger *Wrapper) Info(message string, data logrus.Fields) {
	ctx := logger.WithFields(data)

	ctx.Info(message)
}

//Warning - Wrap Warning from logrus logger
func (logger *Wrapper) Warning(message string, data logrus.Fields) {
	ctx := logger.WithFields(data)

	ctx.Warning(message)
}

//Error - Wrap Error from logrus logger
func (logger *Wrapper) Error(message string, data logrus.Fields) {
	ctx := logger.WithFields(data)

	ctx.Error(message)
}
