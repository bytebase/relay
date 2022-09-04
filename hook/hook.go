package hook

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/bytebase/relay/sink"
	"github.com/flamego/flamego"
	"github.com/hashicorp/go-multierror"
)

// Response defines the handler's return value
// - Sets http.StatusOK and payload if you want the coresponding sink list to process the payload.
// - Sets other 2xx code if you want to short-circuit the processing, but you don't want to indicate
//   an error (e.g. skip the processing). You can optionally set a detail to explain the reason.
//   This will show up on the webhook sender's page.
// - Sets other HTTP code and err if you do want to indicate an error.
type Response struct {
	httpCode int
	err      error
	detail   string
	payload  interface{}
}

// Hooker is the interface for the webhook originator.
type Hooker interface {
	// handler returns the hook handler, returns error if precondition fails such as invalid flag values.
	handler() (func(r *http.Request) Response, error)
}

var (
	hookersMu sync.RWMutex
	hookers   = make(map[string]Hooker)
)

// Mount mounts the hook and corresponding sink list under the given path.
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

	for _, s := range ss {
		if err := s.Mount(); err != nil {
			panic("hook: Failed to mount sink " + err.Error())
		}
	}

	f.Post(path, func(r *http.Request) (int, string) {
		resp := handler(r)

		if resp.httpCode == http.StatusOK {
			var result error
			for _, s := range ss {
				if err := s.Process(r.Context(), path, resp.payload); err != nil {
					result = multierror.Append(result, err)
				}
			}
			if result != nil {
				return http.StatusInternalServerError, fmt.Sprintf("Encountered error send to sink %q: %v", path, err)
			}
			return http.StatusOK, "OK"
		}

		if resp.err != nil {
			return resp.httpCode, resp.err.Error()
		}
		return resp.httpCode, resp.detail
	})

	hookers[path] = h
}
