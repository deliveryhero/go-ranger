package fdhttp_test

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/foodora/go-ranger/fdhttp"
	"github.com/stretchr/testify/assert"
)

func init() {
	fdhttp.DefaultLogger = log.New(os.Stdout, "", 0)
}

func startServer(addr string) (srv *fdhttp.Server, stopChan chan struct{}) {
	srv = fdhttp.NewServer(addr)
	stopChan = make(chan struct{})

	return srv, stopChan
}

func stopServer(t *testing.T, srv *fdhttp.Server, stopChan chan struct{}) error {
	// give some time to server runs
	time.Sleep(1 * time.Millisecond)

	err := srv.Stop()
	fmt.Println(err)

	select {
	case <-stopChan:
		return err
	case <-time.After(2 * time.Second):
		t.Error("Server took so long to stop")
	}

	return nil
}

func TestServe_InvalidPort(t *testing.T) {
	srv := fdhttp.NewServer("invalid")
	err := srv.Start(nil)
	assert.EqualError(t, err, "address invalid: missing port in address")
}

func TestServe_DifferentAddresses(t *testing.T) {
	addrs := []string{
		"8123",           // only port
		"127.0.0.1:8124", // host and port
	}

	for _, addr := range addrs {
		srv, stopChan := startServer(addr)

		go func() {
			err := srv.Start(nil)
			assert.Error(t, fdhttp.ErrServerStopped, err)
			stopChan <- struct{}{}
		}()

		err := stopServer(t, srv, stopChan)
		assert.NoError(t, err)
	}
}

func TestServe_RunningTwice(t *testing.T) {
	srv, stopChan := startServer("8125")
	go func() {
		err := srv.Start(nil)
		assert.Error(t, fdhttp.ErrServerStopped, err)
		stopChan <- struct{}{}
	}()

	// give some time to server runs
	time.Sleep(1 * time.Millisecond)

	err := srv.Start(nil)
	assert.Error(t, fdhttp.ErrServerAlreadyRunning, err)

	stopServer(t, srv, stopChan)
}

func TestServe_StoppingTwice(t *testing.T) {
	srv, stopChan := startServer("8126")

	err := srv.Stop()
	assert.Equal(t, fdhttp.ErrServerNotRunning, err)

	go func() {
		err := srv.Start(nil)
		assert.Error(t, fdhttp.ErrServerStopped, err)
		stopChan <- struct{}{}
	}()

	err = stopServer(t, srv, stopChan)
	assert.NoError(t, err)

	err = srv.Stop()
	assert.Equal(t, fdhttp.ErrServerNotRunning, err)
}
