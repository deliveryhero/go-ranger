package ranger_logger

import (
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
)

// LoggerData used to log any data structure
type LoggerData map[string]interface{}

// Hook wrapper
type Hook logrus.Hook

//LoggerInterface ...
type LoggerInterface interface {
	Info(message string, data LoggerData)
	Debug(message string, data LoggerData)
	Warning(message string, data LoggerData)
	Error(message string, data LoggerData)
	Panic(message string, data LoggerData)
}

//Wrapper - Wrap a logrus logger
type Wrapper struct {
	*logrus.Logger            // see promoted methods https://www.goinggo.net/2015/09/composition-with-go.html,
	AppData        LoggerData // default fields
}

// JSONFormatter Wrapper for logrus.JSONFormatter
type JSONFormatter struct {
	logrus.JSONFormatter
}

// GetJSONFormatter https://foodpanda.atlassian.net/browse/DISCOVER-601
func GetJSONFormatter() *JSONFormatter {
	jsonFormatter := logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05-0700",
	}
	return &JSONFormatter{jsonFormatter}
}

// Formatter wrappper
type Formatter logrus.Formatter

//NewLogger - LoggerWrapper constructor that uses the given Formatter and io.Writer like os.Stdout
func NewLogger(out io.Writer, appData LoggerData, f Formatter, hooks ...Hook) LoggerInterface {
	log := logrus.New()

	log.Out = out
	log.Formatter = f
	log.Level = logrus.InfoLevel

	for _, h := range hooks {
		if h != nil {
			log.Hooks.Add(h)
		}
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

//Debug - Wrap Debug from logrus logger
func (logger *Wrapper) Debug(message string, data LoggerData) {
	ctx := logger.WithFields(convertToLogrusFields(logger.GetAllFieldsToLog(data)))

	ctx.Debug(message)
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
