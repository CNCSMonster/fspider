package fspider

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	*sync.RWMutex
	watcher *fsnotify.Watcher
	// 记录被监控到的当前仍然存在的文件
	files  map[string]bool
	dirs   map[string]bool
	output chan string
}

func NewSpider() *Spider {
	Log("NewSpider")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	output := make(chan string)
	s := &Spider{
		RWMutex: &sync.RWMutex{},
		files:   make(map[string]bool),
		dirs:    make(map[string]bool),
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
				if event.Op&fsnotify.Chmod == fsnotify.Chmod {
					continue
				} else {
					// create, remove, write, rename
					path = event.Name
				}
				// log.Println(path)
				path = filepath.Clean(path)
				s.output <- path
				if event.Op&fsnotify.Create == fsnotify.Create {
					s.Spide(path)
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					s.UnSpide(path)
				} else if event.Op&fsnotify.Rename == fsnotify.Rename {
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

// use to watch a dir recursively,path can only be a dir
func (s *Spider) Spide(path string) error {
	s.Lock()
	defer s.Unlock()
	path = filepath.Clean(path)
	toSpide := []string{path}
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		s.files[path] = true
		return nil
	}
	for len(toSpide) > 0 {
		nextToSpide := []string{}
		for _, p := range toSpide {
			if s.isSpiderable(p) {
				s.watcher.Add(p)
				s.dirs[p] = true
				items, err := os.ReadDir(p)
				if err != nil {
					return err
				}
				for _, item := range items {
					if item.IsDir() {
						subDir := filepath.Clean(p + "/" + item.Name())
						nextToSpide = append(nextToSpide, subDir)
					} else {
						file := filepath.Clean(p + "/" + item.Name())
						s.files[file] = true
					}
				}
			}
		}
		toSpide = nextToSpide
	}
	return nil
}

// use to unwatch a dir unrecursively,path can only be a dir
func (s *Spider) UnSpide(path string) error {
	s.Lock()
	defer s.Unlock()
	delete(s.dirs, path)
	delete(s.files, path)
	err := s.watcher.Remove(path)
	return err
}

// to check if a path is spiderable, and this method is not thread safe
func (s *Spider) isSpiderable(path string) bool {
	stat, err := os.Stat(path)
	if err != nil || !stat.IsDir() {
		fmt.Println("path is not a dir so can not be watched")
		return false
	}
	_, exists := s.dirs[path]
	return !exists
}

// to check if a path is spiderable, and this method is thread safe
func (s *Spider) IsSpiderable(path string) bool {
	s.RLock()
	defer s.RUnlock()
	return s.isSpiderable(path)
}

func (s *Spider) FilesChanged() <-chan string {
	return s.output
}

// return all paths that are being watched, this method is thread safe
func (s *Spider) AllPaths() []string {
	s.RLock()
	defer s.RUnlock()
	paths := make([]string, 0, len(s.dirs))
	for dir := range s.dirs {
		paths = append(paths, dir)
	}
	for file := range s.files {
		paths = append(paths, file)
	}
	return paths
}
func (s *Spider) AllFiles() []string {
	s.RLock()
	defer s.RUnlock()
	files := make([]string, 0, len(s.files))
	for file := range s.files {
		files = append(files, file)
	}
	return files
}
func (s *Spider) AllDirs() []string {
	s.RLock()
	defer s.RUnlock()
	dirs := make([]string, 0, len(s.dirs))
	for dir := range s.dirs {
		dirs = append(dirs, dir)
	}
	return dirs
}

// stop watching files, this method is thread safe
func (s *Spider) Stop() {
	s.Lock()
	s.watcher.Close()
	s.watcher = nil
	close(s.output)
	s.Unlock()
}
