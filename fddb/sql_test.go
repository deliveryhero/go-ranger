package fddb_test

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/foodora/go-ranger/fddb"
	"github.com/stretchr/testify/assert"
)

func init() {
	fddb.SetLogger(log.New(ioutil.Discard, "", 0))
}

func TestOpenSQL_WithoutRegisterDriver(t *testing.T) {
	_, err := fddb.OpenSQL(fddb.DBConfig{
		Driver: "mysql",
	})
	assert.Error(t, err)
}
