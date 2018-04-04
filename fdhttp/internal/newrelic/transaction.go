package newrelic

import (
	"context"
	"errors"

	newrelic "github.com/newrelic/go-agent"
)

// NewRelicTransactionKey s a key used inside of context.Context to save the newrelic transaction
var NewRelicTransactionKey = "newrelic_transaction"

// SetTransaction set newrelic transaction into context.
func SetTransaction(ctx context.Context, value newrelic.Transaction) context.Context {
	return context.WithValue(ctx, NewRelicTransactionKey, value)
}

// Transaction get newrelic transaction from context.
func Transaction(ctx context.Context) newrelic.Transaction {
	v, ok := ctx.Value(NewRelicTransactionKey).(newrelic.Transaction)
	if !ok {
		// your app is trying to access newrelic transaction but you're not
		// using newrelic middleware
		panic(errors.New("fdhttp: newrelic middleware was not called"))
	}

	return v
}
