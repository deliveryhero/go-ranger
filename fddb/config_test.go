package fddb

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDBConfig(t *testing.T) {
	testCases := map[string]struct {
		cfg DBConfig
		res string
	}{
		"mysql basic": {
			cfg: DBConfig{Driver: "mysql"},
			res: "root@tcp(127.0.0.1:3306)/",
		},
		"mysql complete": {
			cfg: DBConfig{
				Driver:   "mysql",
				Host:     "mysql.local",
				Port:     "4321",
				User:     "foodora",
				Password: "pd123",
				DB:       "test",
			},
			res: "foodora:pd123@tcp(mysql.local:4321)/test",
		},
		"postgres basic": {
			cfg: DBConfig{Driver: "postgres"},
			res: "postgres://root@127.0.0.1:5432/",
		},
		"postgres complete": {
			cfg: DBConfig{
				Driver:   "postgres",
				Host:     "postgres.local",
				Port:     "4321",
				User:     "foodora",
				Password: "pd123",
				DB:       "test",
			},
			res: "postgres://foodora:pd123@postgres.local:4321/test",
		},
		"mongodb basic": {
			cfg: DBConfig{Driver: "mongodb"},
			res: "mongodb://root@127.0.0.1:27017/",
		},
		"mongodb complete": {
			cfg: DBConfig{
				Driver:   "mongodb",
				Host:     "mongodb.local",
				Port:     "4321",
				User:     "foodora",
				Password: "pd123",
				DB:       "test",
			},
			res: "mongodb://foodora:pd123@mongodb.local:4321/test",
		},
		"mongodb multiple addrs": {
			cfg: DBConfig{
				Driver: "mongodb",
				Addrs: []string{
					"master.mongodb.local:27017",
					"slave1.mongodb.local:27017",
					"slave2.mongodb.local:27017",
				},
				User: "foodora",
				DB:   "test",
			},
			res: "mongodb://foodora@master.mongodb.local:27017,slave1.mongodb.local:27017,slave2.mongodb.local:27017/test",
		},
		"mongodb with host and address": {
			cfg: DBConfig{
				Driver: "mongodb",
				Host:   "master.mongodb.local",
				Addrs: []string{
					"slave1.mongodb.local:27017",
					"slave2.mongodb.local:27017",
				},
				User:     "foodora",
				Password: "pd123",
				DB:       "test",
			},
			res: "mongodb://foodora:pd123@master.mongodb.local:27017,slave1.mongodb.local:27017,slave2.mongodb.local:27017/test",
		},
		"redis basic": {
			cfg: DBConfig{Driver: "redis"},
			res: "127.0.0.1:6379",
		},
		"redis complete": {
			cfg: DBConfig{
				Driver:   "redis",
				Host:     "redis.remove.host",
				Port:     "4321",
				User:     "foodora",
				Password: "pd123",
				DB:       "test",
			},
			res: "redis.remove.host:4321",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.res, tc.cfg.init().ConnString())
		})
	}

}

func TestDBConfig_NoDriver(t *testing.T) {
	assert.Panics(t, func() {
		DBConfig{}.init()
	})
}

func TestDBConfig_InvalidDriver(t *testing.T) {
	assert.Panics(t, func() {
		DBConfig{Driver: "invalid"}.init()
	})
}

func TestDBConfigString(t *testing.T) {
	c := DBConfig{
		Driver:   "mysql",
		Host:     "127.0.0.1",
		Port:     "3306",
		User:     "root",
		Password: "r007",
		DB:       "test",
	}

	assert.Equal(t, c.String(), "root@127.0.0.1:3306/test")
}

func TestDBConfigConnString_MySQL(t *testing.T) {
	c := DBConfig{
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

func TestDBConfigConnFullString_MySQL(t *testing.T) {
	c := DBConfig{
		Driver: "mysql",
		Host:   "127.0.0.1",
		Port:   "3306",
		User:   "root",
		DB:     "test",
		MysqlOptions: MysqlOptions{
			Timeout:        10000000,
			ReadTimeout:    20000000,
			WriteTimeout:   30000000,
			RejectReadOnly: true,
		},
	}
	assert.Equal(t, c.ConnString(), "root@tcp(127.0.0.1:3306)/test?timeout=10ms&readTimeout=20ms&writeTimeout=30ms&rejectReadOnly=true")

	c.Password = "r007"
	assert.Equal(t, c.ConnString(), "root:r007@tcp(127.0.0.1:3306)/test?timeout=10ms&readTimeout=20ms&writeTimeout=30ms&rejectReadOnly=true")
}

func TestDBConfigConnString_Postgres(t *testing.T) {
	c := DBConfig{
		Driver: "postgres",
		Host:   "127.0.0.1",
		Port:   "5432",
		User:   "root",
		DB:     "test",
	}
	assert.Equal(t, c.ConnString(), "postgres://root@127.0.0.1:5432/test")

	c.Password = "r007"
	assert.Equal(t, c.ConnString(), "postgres://root:r007@127.0.0.1:5432/test")
}

func TestDBConfigConnString_MongoDB(t *testing.T) {
	c := DBConfig{
		Driver: "mongodb",
		Host:   "127.0.0.1",
		Port:   "27017",
		User:   "root",
		DB:     "test",
	}
	assert.Equal(t, c.ConnString(), "mongodb://root@127.0.0.1:27017/test")

	c.Password = "r007"
	assert.Equal(t, c.ConnString(), "mongodb://root:r007@127.0.0.1:27017/test")
}

func TestDBConfigConnString_Redis(t *testing.T) {
	c := DBConfig{
		Driver: "redis",
		Host:   "127.0.0.1",
		Port:   "6379",
		User:   "root",
		DB:     "test",
	}
	assert.Equal(t, c.ConnString(), "127.0.0.1:6379")

	c.Password = "r007"
	assert.Equal(t, c.ConnString(), "127.0.0.1:6379")
}
