package ranger_metrics

import (
	"testing"
)

func TestInitNewRelic(t *testing.T) {
	//t.Error("@todo TestInitNewRelic")
}

func ExampleStartTransactionManually() {
	newRelic := NewNewRelic("appName", "license", nil)

	// Start and end manually
	closeTxn := newRelic.StartTransaction(nil, nil)
	// your code here
	closeTxn()

	// Start and end with defer
	closeTxn = newRelic.StartTransaction(nil, nil)
	defer closeTxn()
	// your code here

	// Start and end with defer at same line
	defer newRelic.StartTransaction(nil, nil)()
	// your code here
}
