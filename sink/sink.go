package sink

import (
	"context"
)

// Sinker is the interface for receiving the webhook payload from the Hooker
type Sinker interface {
	// Mount is called upon being mount to a hooker, common tasks performed inside Mount:
	// - Check flag values.
	Mount() error
	// Process processes the payload extracted by the Hooker.
	Process(c context.Context, path string, payload interface{}) error
}
