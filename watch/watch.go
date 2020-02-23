package watch

import (
	"gos/config"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/radovskyb/watcher"
)

type Watcher struct {
	port      int
	gosConfig *config.GosConfig
	watcher   *watcher.Watcher
	// when a file gets changed a message is sent to the update channel
	update chan string
}

func (w *Watcher) Watch() {
	// SetMaxEvents to 1 to allow at most 1 event's to be received
	// on the Event channel per watching cycle.
	// If SetMaxEvents is not set, the default is to send all events.
	w.watcher.SetMaxEvents(10)

	runner := NewRunner()
	go func() {
		for {
			select {
			case event := <-w.watcher.Event:
				if !event.IsDir() {
					currentPath, err := os.Getwd()
					if err != nil {
						log.Fatalln(err)
					}
					pth, err := filepath.Rel(currentPath, event.Path)
					if err != nil {
						log.Fatalln(err)
					}
					for _, svc := range w.gosConfig.Services {
						if strings.HasPrefix(pth, svc) {
							w.update <- svc
						}
					}
					if pth == config.FileName {
						gosConfig, err := config.Read()
						if err != nil {
							continue
						}
						w.gosConfig = gosConfig
						for _, svc := range w.gosConfig.Services {
							w.update <- svc
						}
					}
				}
			case err := <-w.watcher.Error:
				log.Fatalln(err)
			case <-w.watcher.Closed:
				return
			}
		}
	}()
	if err := w.watcher.Add(config.FileName); err != nil {
		log.Fatalln(err)
	}

	if err := w.watcher.Ignore(".git"); err != nil {
		log.Fatalln(err)
	}
	for _, service := range w.gosConfig.Services {
		if err := w.watcher.AddRecursive(service); err != nil {
			log.Fatalln(err)
		}
		if err := w.watcher.Ignore(path.Join(service, "gen")); err != nil {
			log.Fatalln(err)
		}
	}
	go func() {
		time.Sleep(1 * time.Second)
		runner.Run()
	}()
	if err := w.watcher.Start(time.Millisecond * 50); err != nil {
		log.Fatalln(err)
	}
}

// Wait waits for the latest messages
func (w *Watcher) Wait() <-chan string {
	return w.update
}

// Close closes the fsnotify watcher channel
func (w *Watcher) Close() {
	close(w.update)
}

func NewWatcher(port int, gosConfig *config.GosConfig) *Watcher {
	return &Watcher{
		port:      port,
		update:    make(chan string),
		gosConfig: gosConfig,
		watcher:   watcher.New(),
	}
}
func Run(port int) {
	gosConfig, err := config.Read()
	if err != nil {
		panic(err)
	}
	r := NewRunner()
	w := NewWatcher(port, gosConfig)
	// wait for build and run the binary with given params
	go r.Run()

	b := NewBuilder(w, r)

	// build given package
	go b.Build()

	// listen for further changes
	go w.Watch()

	r.Wait()
}
