package fdhttp

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type (
	// Server represent a server that can be Started or Stopped.
	Server struct {
		addr     string
		running  uint32
		runLock  sync.Mutex
		stopLock sync.Mutex
		router   *Router
		httpSrv  *http.Server

		// Logger will be setted with DefaultLogger when NewServer is called
		// but you can overwrite later only in this instance.
		Logger Logger
	}
)

// Errors that can be returned when Start() or Stop() are called.
var (
	ErrServerStopped        = errors.New("fdhttp: server stopped")
	ErrServerNotRunning     = errors.New("fdhttp: server not running")
	ErrServerAlreadyRunning = errors.New("fdhttp: server already running")
)

// NewServer return a new server instance and will be run in the address informed.
// Address can be "0.0.0.0:8080", ":8080" or just the port "8080".
func NewServer(addr string) *Server {
	return &Server{
		addr:   addr,
		Logger: defaultLogger,
	}
}

// checkAddr verify if address informed is valid
func checkAddr(addr string) (string, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		// verifiy if just port was informed
		if _, portErr := strconv.Atoi(addr); portErr != nil {
			return "", err
		}

		host, port = "0.0.0.0", addr
	}

	return net.JoinHostPort(host, port), nil
}

// Start the server and block until another go rotine call Stop()
// or return imediatily in case is not possible to start the server.
func (s *Server) Start(r *Router) error {
	if atomic.LoadUint32(&s.running) == 1 {
		return ErrServerAlreadyRunning
	}

	// As this function block I cannot use defer to Unlock
	s.runLock.Lock()

	if s.running == 1 {
		s.runLock.Unlock()
		return ErrServerAlreadyRunning
	}

	var err error

	s.addr, err = checkAddr(s.addr)
	if err != nil {
		s.runLock.Unlock()
		return err
	}

	s.Logger.Printf("Running http server on %s...", s.addr)

	s.httpSrv = &http.Server{
		Addr: s.addr,
	}

	if r != nil {
		s.router = r
		s.router.Init()
		s.httpSrv.Handler = s.router
	}

	errChan := make(chan error)

	go func() {
		errChan <- s.httpSrv.ListenAndServe()
	}()

	atomic.StoreUint32(&s.running, 1)
	s.runLock.Unlock()

	err = <-errChan
	if err == http.ErrServerClosed {
		err = ErrServerStopped
	}

	return err
}

// Stop the server. Return nil in case of success, but can fail due
// take so long to shutdown. In case of success Start() will return
// fdhttp.ErrServerStopped
func (s *Server) Stop() error {
	if atomic.LoadUint32(&s.running) == 0 {
		return ErrServerNotRunning
	}

	defer Un(Lock(&s.stopLock))

	if s.running == 0 {
		return ErrServerNotRunning
	}

	defer atomic.StoreUint32(&s.running, 0)

	s.Logger.Printf("Stopping http server...")

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	return s.httpSrv.Shutdown(ctx)
}
