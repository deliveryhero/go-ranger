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
		HTTPSrv  *http.Server

		// Logger will be setted with DefaultLogger when NewServer is called
		// but you can overwrite later only in this instance.
		Logger Logger

		// Timeouts will be setted with default values according
		// https://blog.cloudflare.com/exposing-go-on-the-internet/#timeouts
		// when NewServer is called but you can overwrite them later only in this instance
		ReadTimeout  time.Duration
		WriteTimeout time.Duration
		IdleTimeout  time.Duration
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
	// Default timeouts to  prevent unclosed requests leaking memory.
	// https://blog.cloudflare.com/exposing-go-on-the-internet/#timeouts
	return &Server{
		addr:         addr,
		Logger:       defaultLogger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
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

	s.HTTPSrv = &http.Server{
		Addr:         s.addr,
		ReadTimeout:  s.ReadTimeout,
		WriteTimeout: s.WriteTimeout,
		IdleTimeout:  s.IdleTimeout,
	}

	if r != nil {
		if r.parent != nil {
			panic("Unable to start server with a sub router")
		}

		s.router = r
		s.router.Init()
		s.HTTPSrv.Handler = s.router
	}

	errChan := make(chan error)

	go func() {
		errChan <- s.HTTPSrv.ListenAndServe()
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
func (s *Server) Stop(ctx context.Context) error {
	if atomic.LoadUint32(&s.running) == 0 {
		return ErrServerNotRunning
	}

	defer Un(Lock(&s.stopLock))

	if s.running == 0 {
		return ErrServerNotRunning
	}

	defer atomic.StoreUint32(&s.running, 0)

	s.Logger.Printf("Stopping http server...")

	return s.HTTPSrv.Shutdown(ctx)
}
