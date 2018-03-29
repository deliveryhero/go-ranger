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
	Server struct {
		logger   Logger
		addr     string
		running  uint32
		runLock  sync.Mutex
		stopLock sync.Mutex
		httpSrv  *http.Server
	}

	Router struct {
	}
)

var (
	ErrServerStopped        = errors.New("fdhttp: server stopped")
	ErrServerNotRunning     = errors.New("fdhttp: server not running")
	ErrServerAlreadyRunning = errors.New("fdhttp: server already running")
)

func NewServer(addr string) *Server {
	return &Server{
		addr:   addr,
		logger: DefaultLogger,
	}
}

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

	s.logger.Printf("Running http server on %s...", s.addr)

	s.httpSrv = &http.Server{
		Addr: s.addr,
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

func (s *Server) Stop() error {
	if atomic.LoadUint32(&s.running) == 0 {
		return ErrServerNotRunning
	}

	defer Un(Lock(&s.stopLock))

	if s.running == 0 {
		return ErrServerNotRunning
	}

	defer atomic.StoreUint32(&s.running, 0)

	s.logger.Printf("Stopping http server...")

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	return s.httpSrv.Shutdown(ctx)
}
