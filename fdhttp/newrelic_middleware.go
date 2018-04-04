package fdhttp

import (
	"context"
	"fmt"
	"net/http"

	fdnewrelic "github.com/foodora/go-ranger/fdhttp/internal/newrelic"
	newrelic "github.com/newrelic/go-agent"
)

// NewRelicMiddleware create a newrelic middleware
func NewRelicMiddleware(appName, license string) Middleware {
	config := newrelic.NewConfig(appName, license)
	app, err := newrelic.NewApplication(config)
	if err != nil {
		panic(fmt.Errorf("Cannot create newrelic application: %s", err))
	}

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, req *http.Request) {
			txn := app.StartTransaction(req.URL.Path, w, req)
			defer txn.End()

			ctx := fdnewrelic.SetTransaction(req.Context(), txn)
			req = req.WithContext(ctx)

			next.ServeHTTP(txn, req)
		}

		return http.HandlerFunc(fn)
	}
}

// NewRelicTransaction get newrelic transaction from context.
//
// To send an error to newrelic use:
//  txn := fdhttp.NewRelicTransaction(ctx)
//  txn.NoticeError(errors.New("my error message"))
func NewRelicTransaction(ctx context.Context) newrelic.Transaction {
	return fdnewrelic.Transaction(ctx)
}

// NewRelicStartSegment start a segment inside of transaction.
//
// To monitor the time spend inside a function use:
//  defer fdhttp.NewRelicStartSegment(ctx, "my-function").End()
// You also can use nested segments:
//  s1 := fdhttp.NewRelicStartSegment(ctx, "outerSegment")
//  s2 := fdhttp.NewRelicStartSegment(ctx, "innerSegment")
//  // s2 must be ended before s1
//  s2.End()
//  s1.End()
func NewRelicStartSegment(ctx context.Context, name string) newrelic.Segment {
	txn := fdnewrelic.Transaction(ctx)
	return newrelic.StartSegment(txn, name)
}
