package fddb

import (
	"errors"
	"fmt"
)

type SQLConfig struct {
	Driver   string
	Host     string
	Port     string
	User     string
	Password string
	DB       string
}

var availableDrivers = map[string]struct{}{
	"mysql":    {},
	"postgres": {},
}

// ErrNoDriverSpecified is returned when you call fddb.OpenSQL without
// specificy the driver name.
var ErrNoDriverSpecified = errors.New("fddb: sql driver was not specified")

// ErrUnknownDriver is returned when you call fddb.OpenSQL specificing
// a unknown driver name.
var ErrUnknownDriver = errors.New("fddb: sql driver unknown")

// DefaultConfig return a new config filling with default values
// field that was not provided.
func (c SQLConfig) DefaultConfig() SQLConfig {
	if c.Driver == "" {
		panic(ErrNoDriverSpecified)
	}
	if _, ok := availableDrivers[c.Driver]; !ok {
		panic(ErrUnknownDriver)
	}

	if c.User == "" {
		c.User = "root"
	}
	if c.Host == "" {
		c.Host = "127.0.0.1"
	}
	if c.Port == "" {
		switch c.Driver {
		case "mysql":
			c.Port = "3306"
		case "postgres":
			c.Port = "5432"
		}
	}

	return c
}

func (c SQLConfig) String() string {
	return fmt.Sprintf("%s@%s:%s/%s", c.User, c.Host, c.Port, c.DB)
}

func (c SQLConfig) ConnString() string {
	usrPwd := c.User
	if c.Password != "" {
		usrPwd += ":" + c.Password
	}

	switch c.Driver {
	case "mysql":
		return fmt.Sprintf("%s@tcp(%s:%s)/%s", usrPwd, c.Host, c.Port, c.DB)
	case "postgres":
		return fmt.Sprintf("postgres://%s@%s:%s/%s", usrPwd, c.Host, c.Port, c.DB)
	}

	return fmt.Sprintf("%s@%s:%s/%s", usrPwd, c.Host, c.Port, c.DB)
}
