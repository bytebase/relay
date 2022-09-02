package hook

import (
	"sync"

	"github.com/flamego/flamego"
	"github.com/hashicorp/go-multierror"
	flag "github.com/spf13/pflag"
)

// Hooker is the interface for the webhook originator.
type Hooker interface {
	// register registers itself to be used, common tasks include:
	// 1. Declare flags
	// 2. Register router handler for the webhook event.
	register(fs *flag.FlagSet, f *flamego.Flame, path string)
	// prepare performs validation checks before to be used, common tasks include:
	// 1. Check flag values
	prepare() error
}

var (
	hookersMu sync.RWMutex
	hookers   = make(map[string]Hooker)
)

// Register registers the hook under path.
// e.g. If you register the foo hook at /foo, then you go to service foo's webhook
// setting page and configure the webhook to post events to <<Relay Host>>/foo.
func Register(fs *flag.FlagSet, f *flamego.Flame, h Hooker, path string) {
	if h == nil {
		panic("hook: Register hooker is nil")
	}

	hookersMu.Lock()
	defer hookersMu.Unlock()
	if _, dup := hookers[path]; dup {
		panic("hook: Register called twice for hooker " + path)
	}
	h.register(fs, f, path)
	hookers[path] = h
}

// Prepare prepares all registered hooks before running (e.g. validate flags).
func Prepare() error {
	hookersMu.Lock()
	defer hookersMu.Unlock()

	var result error
	for _, hook := range hookers {
		if err := hook.prepare(); err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result
}
