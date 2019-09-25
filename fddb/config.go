package fddb

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type DBConfig struct {
	initialized int32

	// Driver for now can be mysql, postgres, mongodb or dynamodb
	Driver string

	// Host should be use together with Port to build the db address
	Host string

	// Port will be append to Host to build the db address
	Port string

	// Addrs should be use when you want to inform more than one
	// host and port, but also will work with just one.
	// If you specify Host, Port and Addrs we'll try to connect
	// to all (host:port and each address informed in Addrs)
	Addrs []string

	User     string
	Password string
	DB       string
	Timeout  time.Duration
}

var availableDrivers = map[string]DBConfig{
	"mysql":    {Port: "3306"},
	"postgres": {Port: "5432"},
	"mongodb":  {Port: "27017"},
	"redis":    {Port: "6379"},
}

// ErrNoDriverSpecified will be panic when you call fddb.Open without
// specificy the driver name.
var ErrNoDriverSpecified = errors.New("fddb: driver was not specified")

// ErrUnknownDriver will be panic when you call fddb.Open specificing
// a unknown driver name.
var ErrUnknownDriver = errors.New("fddb: driver unknown")

// ErrInvalidPort will be panic when you call fddb.Open without specify
// a port and we don't have a default one.
var ErrInvalidPort = errors.New("fddb: invalid port")

// DefaultConfig return a new config filling with default values
// field that was not provided.
func (c DBConfig) init() DBConfig {
	if c.Driver == "" {
		panic(ErrNoDriverSpecified)
	}

	var (
		defaultCfg DBConfig
		ok         bool
	)

	if defaultCfg, ok = availableDrivers[c.Driver]; !ok {
		panic(ErrUnknownDriver)
	}

	if len(c.Addrs) == 0 && c.Host == "" {
		if defaultCfg.Host != "" {
			c.Host = defaultCfg.Host
		} else {
			c.Host = "127.0.0.1"
		}
	}

	if c.Host != "" && c.Port == "" {
		if defaultCfg.Port != "" {
			c.Port = defaultCfg.Port
		} else {
			panic(ErrInvalidPort)
		}
	}

	if c.User == "" {
		if defaultCfg.User != "" {
			c.User = defaultCfg.User
		} else {
			c.User = "root"
		}
	}

	if c.Timeout.String() == "0s" {
		if defaultCfg.Timeout.String() != "0s" {
			c.Timeout = defaultCfg.Timeout
		}
	}

	return c
}

func (c DBConfig) String() string {
	return fmt.Sprintf("%s@%s:%s/%s", c.User, c.Host, c.Port, c.DB)
}

func (c DBConfig) ConnString() string {
	usrPwd := c.User
	if c.Password != "" {
		usrPwd += ":" + c.Password
	}

	timeOut := ""
	if c.Timeout.String() != "0s" {
		timeOut = "?" + c.Timeout.String()
	}

	var host string
	if c.Host != "" {
		host = fmt.Sprintf("%s:%s", c.Host, c.Port)
	}

	switch c.Driver {
	case "mysql":
		return fmt.Sprintf("%s@tcp(%s)/%s%s", usrPwd, host, c.DB, timeOut)
	case "redis":
		return host
	case "mongodb":
		if len(c.Addrs) > 0 {
			addrs := c.Addrs[:]
			if host != "" {
				addrs = append([]string{host}, addrs...)
			}

			host = strings.Join(addrs, ",")
		}
		fallthrough
	default:
		return fmt.Sprintf("%s://%s@%s/%s", c.Driver, usrPwd, host, c.DB)
	}

	// should never reach it
	return ""
}
