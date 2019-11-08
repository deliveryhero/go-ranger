package fdapm

import (
	"context"
	"errors"
	"net/http"

	"github.com/foodora/go-ranger/fdhttp/fdmiddleware"
	newrelic "github.com/newrelic/go-agent"
)

// NewRelicMiddleware create a newrelic middleware
func NewRelicMiddleware(app newrelic.Application) fdmiddleware.Middleware {
	return fdmiddleware.MiddlewareFunc(func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, req *http.Request) {
			txn := newrelic.FromContext(req.Context())
			defer txn.End()

			ctx := SetNewRelicTransaction(req.Context(), txn)
			req = req.WithContext(ctx)

			next.ServeHTTP(txn, req)
		}

		return http.HandlerFunc(fn)
	})
}

// SetNewRelicTransaction set newrelic transaction into context.
func SetNewRelicTransaction(ctx context.Context, txn newrelic.Transaction) context.Context {
	return newrelic.NewContext(ctx, txn)
}

// NewRelicTransaction get newrelic transaction from context.
//
// To send an error to newrelic use:
//  txn := fdapm.NewRelicTransaction(ctx)
//  txn.NoticeError(errors.New("my error message"))
func NewRelicTransaction(ctx context.Context) newrelic.Transaction {
	txn := newrelic.FromContext(ctx)
	if txn == nil {
		// your app is trying to access newrelic transaction but you're not
		// using newrelic middleware
		panic(errors.New("fdapm: newrelic middleware was not called"))
	}

	return txn
}

// NewRelicStartSegment start a segment inside of transaction.
//
// To monitor the time spend inside a function use:
//  defer fdapm.NewRelicStartSegment(ctx, "my-function").End()
// You also can use nested segments:
//  s1 := fdapm.NewRelicStartSegment(ctx, "outerSegment")
//  s2 := fdapm.NewRelicStartSegment(ctx, "innerSegment")
//  // s2 must be ended before s1
//  s2.End()
//  s1.End()
func NewRelicStartSegment(ctx context.Context, name string) *newrelic.Segment {
	txn := NewRelicTransaction(ctx)
	return newrelic.StartSegment(txn, name)
}
