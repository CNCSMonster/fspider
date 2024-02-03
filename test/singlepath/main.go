package main

import (
	"github.com/cncsmonster/fspider"
)

func main() {
	spider := fspider.NewSpider()
	// add current dir to watch list
	err := spider.Spide("./")
	if err != nil {
		panic(err)
	}
	for v := range spider.FilesChanged() {
		println("file changed: ", v)
	}
	<-make(chan struct{})
}
