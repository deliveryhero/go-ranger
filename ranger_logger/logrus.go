package ranger_logger

import (
	"io"
	"net"
	"net/http"

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
	*logrus.Logger            // see promoted methods https://www.goinggo.net/2015/09/composition-with-go.html,
	AppData        LoggerData // default fields
}

//NewLoggerWithLogstashHook - LoggerWrapper constructor with logstash hook
func NewLoggerWithLogstashHook(protocol string, addr string, appName string, appData LoggerData) LoggerInterface {
	log := logrus.New()

	if conn, err := net.Dial(protocol, addr); err == nil {
		hook := logrustash.New(conn, &logrus.JSONFormatter{})
		log.Hooks.Add(hook)
	} else {
		log.Warn("unable to connect to logstash")
	}

	return &Wrapper{log, appData}
}

//NewLoggerStdout - LoggerWrapper constructor that uses the given io.Writer like os.Stdout
func NewLoggerIoWriter(out io.Writer, appData LoggerData) LoggerInterface {
	log := &logrus.Logger{
		Out:       out,
		Formatter: &logrus.JSONFormatter{},
		Level:     logrus.InfoLevel,
	}

	return &Wrapper{log, appData}
}

//CreateFieldsFromRequest - Create a logrus.Fields object from a Request
func CreateFieldsFromRequest(r *http.Request) LoggerData {
	return LoggerData{
		"client_ip":      r.RemoteAddr,
		"request_method": r.Method,
		"request_uri":    r.RequestURI,
		"request_host":   r.Host,
	}
}

//Info - Wrap Info from logrus logger
func (logger *Wrapper) Info(message string, data LoggerData) {
	ctx := logger.WithFields(convertToLogrusFields(logger.GetAllFieldsToLog(data)))

	ctx.Info(message)
}

//Warning - Wrap Warning from logrus logger
func (logger *Wrapper) Warning(message string, data LoggerData) {
	ctx := logger.WithFields(convertToLogrusFields(logger.GetAllFieldsToLog(data)))

	ctx.Warning(message)
}

//Error - Wrap Error from logrus logger
func (logger *Wrapper) Error(message string, data LoggerData) {
	ctx := logger.WithFields(convertToLogrusFields(logger.GetAllFieldsToLog(data)))

	ctx.Error(message)
}

//Panic - Wrap Panic from logrus logger
func (logger *Wrapper) Panic(message string, data LoggerData) {
	ctx := logger.WithFields(convertToLogrusFields(logger.GetAllFieldsToLog(data)))

	ctx.Panic(message)
}

func convertToLogrusFields(loggerData LoggerData) logrus.Fields {
	fields := logrus.Fields{}
	for k, v := range loggerData {
		fields[k] = v
	}
	return fields
}

//GetAllFieldsToLog – merges default fields with the given ones
func (logger *Wrapper) GetAllFieldsToLog(data LoggerData) LoggerData {
	result := make(LoggerData)

	for k, v := range logger.AppData {
		result[k] = v
	}

	for k, v := range data {
		result[k] = v
	}

	return result
}
