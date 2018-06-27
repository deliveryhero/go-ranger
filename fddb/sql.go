package fddb

import (
	"database/sql"
	"fmt"
	"time"
)

// Open a connection with sql database using provide configuration
func OpenSQL(c DBConfig) (*sql.DB, error) {
	c = c.init()
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
				return nil, fmt.Errorf("fddb: %s: unable to connect to '%s': %s", c.Driver, c, err)
			}

			prefix = fmt.Sprintf("[%d/%d]", attempt, MaxConnAttempt)
		}

		sleepFor := sleepFn(attempt)
		defaultLogger.Printf("%s Unable to connect to database, trying again in %s: %s", prefix, sleepFor, err)
		time.Sleep(sleepFor)

	}

	return db, nil
}
