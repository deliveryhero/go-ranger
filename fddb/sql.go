package fddb

import (
	"database/sql"
	"fmt"
	"math"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// MaxConnAttempt is the number of times that it'll try to connect to the database
// in case of error. Use -1 to try forever and 0 to not try again
var MaxConnAttempt = 5

// BackoffFunc is a generator of duration that we use to sleep
// between attempt in case we cannot connect to database.
// Default implementation use following formula:
// MinBackoff * pow(2, attempt) will will result in 5s, 10s, 20s, 40s, 1m20s, 2m40s
// You also can override with something like this:
//
// 		fddb.BackoffFunc = func() func(attempt int) time.Duration {
// 			d := []time.Duration{
// 				5*time.Second,
// 				10*time.Second,
// 				20*time.Second,
// 				1*time.Minute + 20*time.Second,
// 				2*time.Minute + 40*time.Second,
// 			}
// 			return func(attempt int) time.Duration {
// 				if attempt > len(d) {
// 					return d[len(d)-1]
//				}
// 				return d[attempt-1]
// 			}
// 		}
var BackoffFunc = func() func(attempt int) time.Duration {
	startBackoff := float64(5 * time.Second)

	return func(attempt int) time.Duration {
		return time.Duration(startBackoff * math.Pow(2, float64(attempt-1)))
	}
}

// DefaultMaxOpenConnection call SetMaxOpenConns to limit the number of connection
// because by default no limit is setted, you also can override it, calling:
// db.SetMaxOpenConns(N)
var DefaultMaxOpenConnection = 100

// Open a connection with sql database using provide configuration
func OpenSQL(c SQLConfig) (*sql.DB, error) {
	c = c.DefaultConfig()
	defaultLogger.Printf("Connecting to '%s'...", c.String())

	db, err := sql.Open(c.Driver, c.ConnString())
	if err != nil {
		return db, err
	}

	db.SetMaxOpenConns(DefaultMaxOpenConnection)
	db.SetConnMaxLifetime(1 * time.Hour)

	var attempt int
	sleepFn := BackoffFunc()

	for {
		err := db.Ping()
		if err == nil {
			break
		}

		attempt++

		var prefix string
		if MaxConnAttempt > -1 {
			if attempt > MaxConnAttempt {
				return nil, fmt.Errorf("fddb: %s: cannot connect to '%s': %s", c.Driver, c, err)
			}

			prefix = fmt.Sprintf("[%d/%d] ", attempt, MaxConnAttempt)
		}

		sleepFor := sleepFn(attempt)
		defaultLogger.Printf("%sCannot connect to database, trying again in %s: %s", prefix, sleepFor, err)
		time.Sleep(sleepFor)

	}

	return db, nil
}
