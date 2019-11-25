package client

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	cecontext "github.com/cloudevents/sdk-go/pkg/cloudevents/context"
	"github.com/fsnotify/fsnotify"
)

// NewDefaultOverrides reads from the given root file system path, reads all
// files in that directory, and then returns a defaulting function that applies
// the files as override extensions to the outbound events. The file names are
// used as the attribute key, and the file contents is used as the attribute
// value.
func NewDefaultOverrides(ctx context.Context, root string) (EventDefaulter, error) {
	logger := cecontext.LoggerFrom(ctx)
	o := NewOverridesObserver(root)
	go func() {
		if err := o.Watch(ctx); err != nil {
			logger.Error(err)
		}
	}()
	return o.Apply, nil
}

// NewOverridesObserver returns an overrides observer that is able to watch a
// filesystem path for updates.
// root/{file} filename as the key of the attribute extension.
// root/{file} contents will be used as the attribute extension value.
func NewOverridesObserver(root string) *overridesObserver {
	return &overridesObserver{
		root:      root,
		overrides: make(map[string]string, 0),
	}
}

type overridesObserver struct {
	root      string
	overrides map[string]string
	mux       sync.Mutex
	once      sync.Once
}

// Apply implements EventDefaulter, it will apply the current overrides to the
// incoming event and return the mutated event.
func (o *overridesObserver) Apply(ctx context.Context, event cloudevents.Event) cloudevents.Event {
	o.once.Do(func() {
		_ = o.Walk()
	})

	o.mux.Lock()
	for k, v := range o.overrides {
		event.SetExtension(k, v)
	}
	o.mux.Unlock()
	return event
}

func (o *overridesObserver) read(file string) error {
	key := filepath.Base(file)
	if !cloudevents.IsAlphaNumericLowercaseLetters(key) {
		logger := cecontext.LoggerFrom(context.Background())
		logger.Warnw("bad file name as attribute key, CloudEvents attribute names MUST consist of lower-case letters ('a' to 'z') or digits ('0' to '9') from the ASCII character set",
			zap.String("file", file),
			zap.String("key", key))
		return nil
	}
	value, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file %s, %v", file, err)
	}
	o.mux.Lock()
	o.overrides[key] = string(value)
	o.mux.Unlock()
	return nil
}

func (o *overridesObserver) delete(file string) {
	o.mux.Lock()
	delete(o.overrides, filepath.Base(file))
	o.mux.Unlock()
}

// String prints out the current override map.
func (o *overridesObserver) String() string {
	b := strings.Builder{}
	for k, v := range o.overrides {
		b.WriteString(fmt.Sprintf("%s: %q\n", k, v))
	}
	return b.String()
}

// Walk will look at filesystem at root and load each file and cache the
// key/value.
func (o *overridesObserver) Walk() error {
	return filepath.Walk(o.root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			if err := o.read(path); err != nil {
				return err
			}
		}
		return nil
	})
}

// Watch will maintain the overrides map to match what is on the filesystem.
// Note: this is a blocking call.
func (o *overridesObserver) Watch(ctx context.Context) error {
	logger := cecontext.LoggerFrom(ctx)
	var err error
	o.once.Do(func() {
		err = o.Walk()
	})
	if err != nil {
		return err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer func() {
		if err := watcher.Close(); err != nil {
			logger.Error(err)
		}
	}()

	if err := watcher.Add(o.root); err != nil {
		return err
	}

	logger.Debugw("Starting watch", zap.Any("path", o.root))
	for {
		select {
		case event := <-watcher.Events:
			logger.Debugw("Got event", zap.Any("event", event))
			switch event.Op {
			case fsnotify.Create, fsnotify.Write:
				if err := o.read(event.Name); err != nil {
					return err
				}

			case fsnotify.Remove, fsnotify.Rename:
				o.delete(event.Name)

			}

		case err := <-watcher.Errors:
			return err

		case <-ctx.Done():
			if err := watcher.Close(); err != nil {
				return fmt.Errorf("unable to close watcher, %s", err.Error())
			}
			return nil
		}
	}
}
