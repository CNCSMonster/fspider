package main

import (
	"github.com/cncsmonster/fspider"
)

func main() {
	spider := fspider.NewSpider()
	// add current dir to watch list
	err := spider.Watch("./")
	if err != nil {
		panic(err)
	}
	for v := range spider.FilesChanged() {
		println(v)
	}
	<-make(chan struct{})
}
