package sink

import (
	"context"
	"sync"

	"github.com/flamego/flamego"
	"github.com/hashicorp/go-multierror"
	flag "github.com/spf13/pflag"
)

// Sinker is the interface for receiving the webhook payload from the Hooker
type Sinker interface {
	register(fs *flag.FlagSet)
	prepare() error
	send(c context.Context, path string, payload interface{}) error
}

var (
	sinksMu sync.RWMutex
	sinks   = make(map[string][]Sinker)
)

// Register registers the sinker for path.
// e.g. If service foo posts the webhook event to path /foo,
// You can register a sinker at path /foo to receive that webhook event payload.
func Register(fs *flag.FlagSet, f *flamego.Flame, s Sinker, path string) {
	sinksMu.Lock()
	defer sinksMu.Unlock()
	if s == nil {
		panic("sink: Register sinker is nil")
	}

	list, has := sinks[path]
	if has {
		for _, item := range list {
			if item == s {
				panic("sink: Register called twice for sinker " + path)
			}
		}
		s.register(fs)
		sinks[path] = append(list, s)
	} else {
		s.register(fs)
		sinks[path] = []Sinker{s}
	}
}

// Prepare prepares all registered sinks before running (e.g. validate flags)
func Prepare() error {
	sinksMu.Lock()
	defer sinksMu.Unlock()

	var result error
	for _, sinkList := range sinks {
		for _, sink := range sinkList {
			if err := sink.prepare(); err != nil {
				result = multierror.Append(result, err)
			}
		}
	}
	return result
}

func Send(c context.Context, path string, payload interface{}) error {
	var result error
	if list := sinks[path]; list != nil {
		for _, s := range list {
			if err := s.send(c, path, payload); err != nil {
				result = multierror.Append(result, err)
			}
		}
	}
	return result
}
