package fspider

import (
	"errors"
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

type Spider interface {
	// use to watch a path
	Spide(path string) error
	// use to unwatch a path
	UnSpide(path string) error
	// return a channel that will receive all the files that have been changed(edited,or created), this method is thread safe
	FilesChanged() <-chan string
	// return all paths that are being watched, this method is thread safe
	AllPaths() []string
	// return all files that are being watched, this method is thread safe
	AllFiles() []string
	// return all dirs that are being watched, this method is thread safe
	AllDirs() []string
	// stop watching files,releasing all resources, this method is thread safe,
	// this function must be called at the end of the life cycle
	Stop()
}

type spiderImpl struct {
	// 对于文件,每个文件准备一个带过滤通道的watcher,每个watcher只监控一个文件,
	// 如果监控到某文件已经被删除,则删除该watcher
	fileWatcher    *fsnotify.Watcher
	fileFilters    map[string]PathSet
	fileFilterLock *sync.RWMutex
	// 对于目录,用一个watcher监控所有
	dirWatcher *fsnotify.Watcher
	// 所有的文件关于internalChan进行
	internalChan chan string
	outputs      []chan string
	outputMutex  *sync.RWMutex
}

func NewSpider() Spider {
	Log("NewSpider")
	dirWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	fileWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	internal := make(chan string)
	s := &spiderImpl{
		fileWatcher:    fileWatcher,
		fileFilters:    make(map[string]PathSet),
		fileFilterLock: &sync.RWMutex{},
		dirWatcher:     dirWatcher,
		internalChan:   internal,
		outputs:        make([]chan string, 0),
		outputMutex:    &sync.RWMutex{},
	}
	process_event := func(event fsnotify.Event, ok bool, sender chan<- string, filter func(string) bool) {
		if !ok {
			return
		}
		var path string
		if event.Op&fsnotify.Chmod == fsnotify.Chmod {
			return
		} else {
			// create, remove, write, rename
			path = event.Name
		}
		path = filepath.Clean(path)
		if filter != nil && !filter(path) {
			Log("process_event", "filter", path)
			return
		}
		sender <- path
		if event.Op&fsnotify.Create == fsnotify.Create {
			s.Spide(path)
		} else if event.Op&fsnotify.Remove == fsnotify.Remove {
			s.UnSpide(path)
		} else if event.Op&fsnotify.Rename == fsnotify.Rename {
			s.UnSpide(path)
		} else {
			// do nothing
			Log("process_event", "do nothing", event.Op, path)
		}
	}

	go func(sender chan<- string) {
		// process event
		for {
			select {
			case event, ok := <-dirWatcher.Events:
				process_event(event, ok, sender, func(path string) bool {
					s.fileFilterLock.RLock()
					defer s.fileFilterLock.RUnlock()
					dir := filepath.Clean(filepath.Dir(path))
					pathSet, ok := s.fileFilters[dir]
					return !ok || !pathSet.Contains(path)
				})
			case event, ok := <-fileWatcher.Events:
				process_event(event, ok, sender, func(path string) bool {
					s.fileFilterLock.RLock()
					defer s.fileFilterLock.RUnlock()
					dir := filepath.Clean(filepath.Dir(path))
					pathSet, ok := s.fileFilters[dir]
					return ok && pathSet.Contains(path)
				})
			case err, ok := <-dirWatcher.Errors:
				_, _ = err, ok
			case err, ok := <-fileWatcher.Errors:
				_, _ = err, ok
			}
		}
	}(internal)
	go func(reciver <-chan string) {
		for path := range reciver {
			s.outputMutex.RLock()
			for _, ch := range s.outputs {
				ch <- path
			}
			s.outputMutex.RUnlock()
		}
	}(internal)
	return s
}

// use to watch a dir recursively,path can only be a dir
func (s *spiderImpl) Spide(path string) error {
	Log("Spide", path)
	if !s.isSpiderable(path) {
		return errors.New("path is not spiderable")
	}
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return s.spideDir(path)
	} else {
		return s.spiderFile(path)
	}
}
func (s *spiderImpl) spideDir(path string) error {
	return filepath.WalkDir(path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			path = filepath.Clean(path)
			Log("spideDir", path)
			s.dirWatcher.Add(path)
			return nil
		} else {
			return s.spiderFile(path)
		}
	})
}
func (s *spiderImpl) spiderFile(path string) error {
	Log("spiderFile", path)
	dir := filepath.Dir(path)
	dir = filepath.Clean(dir)
	Log("spiderFile", "dir:", dir)
	s.fileWatcher.Add(dir)
	s.fileFilterLock.RLock()
	pathSet, ok := s.fileFilters[dir]
	s.fileFilterLock.RUnlock()
	if !ok {
		Log("spiderFile", "dir:", dir, "is not in fileFilters")
		s.fileFilterLock.Lock()
		newPathSet := NewPathSet()
		s.fileFilters[dir] = newPathSet
		s.fileFilterLock.Unlock()
		pathSet = newPathSet
	}
	Log("spiderFile", "add", path, "to", dir)
	pathSet.Add(path)
	return nil
}

// use to unwatch a dir unrecursively,path can only be a dir
func (s *spiderImpl) UnSpide(path string) error {
	path = filepath.Clean(path)
	Log("UnSpide", path)
	s.unspideDir(path)
	s.unspiderFile(path)
	return nil
}

func (s *spiderImpl) unspideDir(path string) error {
	s.dirWatcher.Remove(path)
	return nil
}
func (s *spiderImpl) unspiderFile(path string) error {
	dir := filepath.Dir(path)
	dir = filepath.Clean(dir)
	s.fileFilterLock.RLock()
	if v, ok := s.fileFilters[dir]; ok {
		v.Remove(path)
	}
	s.fileFilterLock.RUnlock()
	return nil
}

// 理论上认为所有路径可以分为: 已经监控了的路径, 尚未监控但是可以监控的路径, 未监控也无法监控的特殊路径
// 如果是 尚未监控但是可以监控的路径,则返回true,否则返回false
// this functions is implemted time-costly, so it should be used carefully
func (s *spiderImpl) isSpiderable(path string) bool {
	stat, err := os.Stat(path)
	Log("isSpiderable", path, "err:", err)
	if err != nil {
		return false
	}
	if stat.IsDir() {
		dirs := s.AllDirs()
		for _, v := range dirs {
			if v == path {
				return false
			}
		}
		Log("isSpiderable", path, "is a dir")
		return true
	}
	files := s.AllFiles()
	for _, v := range files {
		if v == path {
			return false
		}
	}
	return true
}

// return a channel that will receive all the files that have been changed(edited,or created), this method is thread safe
func (s *spiderImpl) FilesChanged() <-chan string {
	s.outputMutex.Lock()
	fileChanged := make(chan string)
	s.outputs = append(s.outputs, fileChanged)
	s.outputMutex.Unlock()
	return fileChanged
}

// return all paths that are being watched, this method is thread safe
func (s *spiderImpl) AllPaths() []string {
	paths := s.AllDirs()
	paths = append(paths, s.AllFiles()...)
	return paths
}
func (s *spiderImpl) AllFiles() []string {
	// 清理掉已经被删除的文件
	fDirs := s.fileWatcher.WatchList()
	files := make([]string, 0)
	if len(s.fileFilters) != len(fDirs) {
		newFDirs := make([]string, len(fDirs))
		for _, v := range fDirs {
			if _, ok := s.fileFilters[v]; !ok {
				s.fileWatcher.Remove(v)
			} else {
				newFDirs = append(newFDirs, v)
			}
		}
		fDirs = newFDirs
	}
	for _, v := range fDirs {
		Log("AllFiles", "fDir:", v)
		if rules, ok := s.fileFilters[v]; ok {
			files = append(files, rules.Paths()...)
		}
	}
	return files
}
func (s *spiderImpl) AllDirs() []string {
	return s.dirWatcher.WatchList()
}

// stop watching files, this method is thread safe
func (s *spiderImpl) Stop() {
	s.fileFilterLock.Lock()
	s.dirWatcher.Close()
	s.fileWatcher.Close()
	close(s.internalChan)
	for _, output := range s.outputs {
		close(output)
	}
	s.fileFilterLock.Unlock()
}
