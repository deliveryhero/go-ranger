package fddb

import (
	"time"

	"github.com/foodora/go-ranger/fdbackoff"
)

// MaxConnAttempt is the number of times that it'll try to connect to the database
// in case of error. Use -1 to try forever and 0 to not try again
var MaxConnAttempt = 5

// BackoffFunc is a generator of duration that we use to sleep
// between attempt in case we cannot connect to database.
// Default implementation use following formula:
// MinBackoff * pow(2, attempt) will will result in 2s, 4s, 8s, 16s, 32s, 1m4s
// You also can override with something like this:
//
//      fddb.BackoffFunc = func() func(attempt int) time.Duration {
//          d := []time.Duration{
//              2*time.Second,
//              4*time.Second,
//              8*time.Second,
//              16*time.Second,
//              32*time.Second,
//              1*time.Minute + 4*time.Second,
//          }
//          return func(attempt int) time.Duration {
//              if attempt > len(d) {
//                  return d[len(d)-1]
//              }
//              return d[attempt-1]
//          }
//      }
//
// Or you can also check some implemented strategy in the fdbackoff package.
var BackoffFunc = fdbackoff.Exponential(2 * time.Second)

// DefaultMaxOpenConnection call SetMaxOpenConns to limit the number of connection
// because by default no limit is setted, you also can override it, calling:
// db.SetMaxOpenConns(N)
var DefaultMaxOpenConnection = 100

// DB is an interface that will be returned when you call fddb.Open
type DB interface{}

// Open will call the right fddb.OpenXXX for you, based on the driver.
// You need to do a type assert with the result to be able to use it
func Open(c DBConfig) (DB, error) {
	switch c.Driver {
	case "mysql", "postgres":
		return OpenSQL(c)
	}

	return nil, ErrUnknownDriver
}
