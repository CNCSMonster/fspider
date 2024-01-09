package fspider

import (
	"context"
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
	cancels map[string]context.CancelFunc
	output  chan string
}

func NewSpider() *Spider {
	return &Spider{
		Mutex:   &sync.Mutex{},
		cancels: make(map[string]context.CancelFunc),
		output:  make(chan string),
	}
}

// use to watch a dir recursively
func (s *Spider) Watch(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}
	s.Lock()
	ctx, cancel := context.WithCancel(context.Background())
	if stat.IsDir() {
		go func(sender chan<- string, ctx context.Context) {
			// use fsnotify to notify a directory
			// Create new watcher.
			watcher, err := fsnotify.NewWatcher()
			if err != nil {
				log.Fatal(err)
			}
			watcher.Add(path)
			defer watcher.Close()
			// process event
			for {
				select {
				case <-ctx.Done():
					return
				case event := <-watcher.Events:
					var path string
					//log.Println("event:", event)
					if event.Op&fsnotify.Create == fsnotify.Create {
						path = event.Name
						log.Println("create:", event.Name)
					}
					if event.Op&fsnotify.Remove == fsnotify.Remove {
						path = event.Name
						log.Println("remove:", event.Name)
					}
					if event.Op&fsnotify.Write == fsnotify.Write {
						path = event.Name
						log.Println("write:", event.Name)
					}
					if event.Op&fsnotify.Rename == fsnotify.Rename {
						path = event.Name
						log.Println("rename:", event.Name)
					}
					// if is dir, add to watcher if create, remove from watcher if remove
					stat, err := os.Stat(path)
					if err != nil {
						log.Println(err)
						continue
					}
					if stat.IsDir() {
						if event.Op&fsnotify.Create == fsnotify.Create {
							watcher.Add(path)
						}
						if event.Op&fsnotify.Remove == fsnotify.Remove {
							s.Mutex.Lock()
							watcher.Remove(path)
							delete(s.cancels, path)
							s.Mutex.Unlock()
						}
					}
					s.output <- path
				case err := <-watcher.Errors:
					log.Println("error:", err)

				}
			}
		}(s.output, ctx)
	} else {
		panic("not support file")
	}
	s.cancels[path] = cancel
	s.Unlock()
	return nil
}

func (s *Spider) FilesChanged() <-chan string {
	return s.output
}
func (s *Spider) Stop() {
	s.Lock()
	// 关闭所有输入协程
	for _, cancel := range s.cancels {
		cancel()
	}
	s.cancels = make(map[string]context.CancelFunc)
	s.Unlock()
}
