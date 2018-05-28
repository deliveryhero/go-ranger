package fddb_test

import (
	"testing"

	"github.com/foodora/go-ranger/fddb"
	"github.com/stretchr/testify/assert"
)

func TestSQLConfigString(t *testing.T) {
	c := fddb.SQLConfig{
		Driver:   "mysql",
		Host:     "127.0.0.1",
		Port:     "3306",
		User:     "root",
		Password: "r007",
		DB:       "test",
	}

	assert.Equal(t, c.String(), "root@127.0.0.1:3306/test")
}

func TestSQLConfigConnString_MySQL(t *testing.T) {
	c := fddb.SQLConfig{
		Driver: "mysql",
		Host:   "127.0.0.1",
		Port:   "3306",
		User:   "root",
		DB:     "test",
	}
	assert.Equal(t, c.ConnString(), "root@tcp(127.0.0.1:3306)/test")

	c.Password = "r007"
	assert.Equal(t, c.ConnString(), "root:r007@tcp(127.0.0.1:3306)/test")
}

func TestSQLConfigConnString_Postgres(t *testing.T) {
	c := fddb.SQLConfig{
		Driver: "postgres",
		Host:   "127.0.0.1",
		Port:   "3306",
		User:   "root",
		DB:     "test",
	}
	assert.Equal(t, c.ConnString(), "postgres://root@127.0.0.1:3306/test")

	c.Password = "r007"
	assert.Equal(t, c.ConnString(), "postgres://root:r007@127.0.0.1:3306/test")
}
