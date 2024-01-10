package fspider

import (
	"log"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// event type
const (
	FS_CREATE = 1 << iota
	FS_DELETE
	FS_MODIFY
	FS_RENAME
)

type Spider struct {
	*sync.Mutex
	watcher *fsnotify.Watcher
	output  chan string
}

func NewSpider() *Spider {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	output := make(chan string)
	s := &Spider{
		Mutex:   &sync.Mutex{},
		watcher: watcher,
		output:  output,
	}
	go func(sender chan<- string) {
		// process event
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				var path string
				if event.Op&fsnotify.Create == fsnotify.Create {
					path = event.Name
				}
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					path = event.Name
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					path = event.Name
				}
				if event.Op&fsnotify.Rename == fsnotify.Rename {
					path = event.Name
				}
				if event.Op&fsnotify.Chmod == fsnotify.Chmod {
					continue
				}
				// log.Println(path)
				s.output <- path
				if event.Op&fsnotify.Create == fsnotify.Create {
					s.Spide(path)
				}
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					s.UnSpide(path)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				if err != nil {
					log.Println("error:", err)
				}
			}
		}
	}(output)
	return s
}

// use to watch a dir recursively
func (s *Spider) Spide(path string) error {
	s.Lock()
	defer s.Unlock()
	if s.IsSpiderable(path) {
		s.watcher.Add(path)
		return nil
	}
	return nil
}
func (s *Spider) UnSpide(path string) error {
	s.Lock()
	defer s.Unlock()
	if s.IsSpiderable(path) {
		s.watcher.Remove(path)
		return nil
	}
	return nil
}

// / to check if a dir is spiderable
// if it has been spidered, return false
func (s *Spider) IsSpiderable(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	if !stat.IsDir() {
		return false
	}
	// TODO: if path has been spidered, return false
	return true
}

func (s *Spider) FilesChanged() <-chan string {
	return s.output
}
func (s *Spider) Stop() {
	s.Lock()
	s.watcher.Close()
	close(s.output)
	s.Unlock()
}
