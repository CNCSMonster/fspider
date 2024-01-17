package main

import (
	"fmt"

	"github.com/cncsmonster/fspider"
)

func main() {
	spider := fspider.NewSpider()
	// add current dir to watch list
	err := spider.Spide("./")
	if err != nil {
		panic(err)
	}
	fmt.Println(spider.AllPaths())
	for v := range spider.FilesChanged() {
		println("file changed: ", v)
		fmt.Println(spider.AllPaths())
	}
	<-make(chan struct{})
}
