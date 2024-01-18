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
		fmt.Println("all paths", spider.AllPaths())
		fmt.Println("all files", spider.AllFiles())
		fmt.Println("all dirs", spider.AllDirs())
	}
	<-make(chan struct{})
}
