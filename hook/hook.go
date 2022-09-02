package hook

import (
	"sync"

	"github.com/flamego/flamego"
	"github.com/hashicorp/go-multierror"
	flag "github.com/spf13/pflag"
)

// Hooker is the interface for the webhook originator.
type Hooker interface {
	prepare() error
	register(fs *flag.FlagSet, f *flamego.Flame, path string)
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

	hooksMu.Lock()
	defer hooksMu.Unlock()
	if _, dup := hooks[path]; dup {
		panic("hook: Register called twice for hooker " + path)
	}
	h.register(fs, f, path)
	hooks[path] = h
}

// Prepare prepares all registered hooks before running (e.g. validate flags).
func Prepare() error {
	hooksMu.Lock()
	defer hooksMu.Unlock()

	var result error
	for _, hook := range hooks {
		if err := hook.prepare(); err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result
}
