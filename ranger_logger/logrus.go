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
	WithData(data LoggerData) LoggerInterface
}

//Wrapper - Wrap a logrus logger
type Wrapper struct {
	*logrus.Logger             // see promoted methods https://www.goinggo.net/2015/09/composition-with-go.html,
	DefaultData     LoggerData // default fields
	ExtraDataPrefix string
}

// JSONFormatter Wrapper for logrus.JSONFormatter
type JSONFormatter struct {
	logrus.JSONFormatter
}

// GetJSONFormatter https://foodpanda.atlassian.net/browse/DISCOVER-601
func GetJSONFormatter() *JSONFormatter {
	jsonFormatter := logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05-0700",
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime: "@timestamp",
		},
	}
	return &JSONFormatter{jsonFormatter}
}

// Formatter wrappper
type Formatter logrus.Formatter

//NewLogger - LoggerWrapper constructor that uses the given Formatter and io.Writer like os.Stdout
func NewLogger(out io.Writer, appData LoggerData, f Formatter, logLevel string, hooks ...Hook) LoggerInterface {
	log := logrus.New()

	log.Out = out
	log.Formatter = f
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.DebugLevel
	}
	log.Level = level

	for _, h := range hooks {
		if h != nil {
			log.Hooks.Add(h)
		}
	}

	return &Wrapper{log, appData, ""}
}

//CreateFieldsFromRequest - Create a logrus.Fields object from a Request
func CreateFieldsFromRequest(r *http.Request) LoggerData {
	return LoggerData{
		"client_ip":      r.Header.Get("X-Forwarded-For"),
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

// WithData returns a copy of the logger with new extra data
func (logger *Wrapper) WithData(data LoggerData) LoggerInterface {
	return &Wrapper{logger.Logger, logger.GetAllFieldsToLog(data), ""}
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

	for k, v := range logger.DefaultData {
		result[k] = v
	}

	ctx := LoggerData{}
	for k, v := range data {
		if _, ok := result[k]; !ok && logger.ExtraDataPrefix != "" {
			ctx[k] = v
		} else {
			result[k] = v
		}
	}

	if len(ctx) > 0 {
		result[logger.ExtraDataPrefix] = ctx
	}

	return result
}

func (logger *Wrapper) SetPrefix(p string) {
	logger.ExtraDataPrefix = p
}
