package fddb

import (
	"database/sql"
	"fmt"
	"math"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// MaxConnAttempt is the number of times that it'll try to connect to the database
// in case of error. Use -1 to try forever and 0 to not try again
var MaxConnAttempt = 5

// MinBackoff is the first duration that it'll wait, after that we'll
// wait for MinBackoff * pow(2, attempt). The default value will give you:
// 5s, 10s, 20s, 40s, 1m20s, 2m40s
var MinBackoff = 5 * time.Second

// Open a connection with mysql using provide configuration
func OpenSQL(c SQLConfig) (*sql.DB, error) {
	c = c.DefaultConfig()
	defaultLogger.Printf("Connecting to %s...", c.String())

	db, err := sql.Open(c.Driver, c.ConnString())
	if err != nil {
		return db, err
	}

	var attempt int
	for {
		err := db.Ping()
		if err == nil {
			break
		}

		var prefix string

		if MaxConnAttempt > -1 {
			if attempt >= MaxConnAttempt {
				return nil, fmt.Errorf("fddb: cannot connect to '%s' in '%s': %s", c.Driver, c, err)
			}

			prefix = fmt.Sprintf("[%d/%d] ", attempt, MaxConnAttempt)
		}

		sleepFor := time.Duration(float64(MinBackoff) * math.Pow(2, float64(attempt)))
		defaultLogger.Printf("%sCannot connect to database, trying again in %s", prefix, sleepFor)
		time.Sleep(sleepFor)

		attempt++
	}

	return db, nil
}
