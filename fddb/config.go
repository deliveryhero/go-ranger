package fddb

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type SQLConfig struct {
	Driver   string
	Host     string
	Port     string
	User     string
	Password string
	DB       string
}

func (c SQLConfig) DefaultConfig() SQLConfig {
	if c.Driver == "" {
		panic("fddb: sql driver was not specified")
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
