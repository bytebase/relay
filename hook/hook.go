package hook

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/bytebase/relay/sink"
	"github.com/flamego/flamego"
	"github.com/hashicorp/go-multierror"
)

// Hooker is the interface for the webhook originator.
type Hooker interface {
	// handler returns the hook handler, returns error if precondition
	// fails such as invalid flag values.
	// For the returned hook handler:
	// - Returns http.StatusOK and payload if you want the coresponding sink list to process the payload.
	// - Returns other 2xx code if you want to short-circuit the processing, but still indicate success.
	// - Returns other http code if you want to indicate error.
	handler() (func(r *http.Request) (int, interface{}), error)
}

var (
	hookersMu sync.RWMutex
	hookers   = make(map[string]Hooker)
)

// Mount mounts the hook and corresponding sink list under path.
//
// - If you mount the foo hook handler at /foo, then you go to service foo's webhook
// setting page and configure the webhook to post events to <<Relay Host>>/foo.
// - If you want the hook handler at /foo to pass the payload to sink [bar, baz], then
// you pass the [bar, baz] sink list.
//
// e.g  hook.Mount(fs, f, "/foo", fooHook, [barSink, bazSink])
func Mount(f *flamego.Flame, path string, h Hooker, ss []sink.Sinker) {
	if h == nil {
		panic("hook: Mount hooker is nil")
	}

	hookersMu.Lock()
	defer hookersMu.Unlock()
	if _, dup := hookers[path]; dup {
		panic("hook: Mount called twice for hooker " + path)
	}
	handler, err := h.handler()
	if err != nil {
		panic("hook: Failed to init handler " + err.Error())
	}

	f.Post(path, func(r *http.Request) (int, string) {
		code, payload := handler(r)

		if code == http.StatusOK {
			var result error
			for _, s := range ss {
				if err := s.Process(r.Context(), path, payload); err != nil {
					result = multierror.Append(result, err)
				}
			}
			if result != nil {
				return http.StatusInternalServerError, fmt.Sprintf("Encountered error send to sink %q: %v", path, err)
			}
			return http.StatusOK, ""
		}

		// TODO(tianzhou): remove type assert
		return code, payload.(string)
	})

	hookers[path] = h
}
