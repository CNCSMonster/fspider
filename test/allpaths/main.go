package main

import (
	"fmt"
	"log"

	"github.com/cncsmonster/fspider"
)

func main() {
	spider := fspider.NewSpider()
	// add current dir to watch list
	err := spider.Spide("./")
	if err != nil {
		panic(err)
	}
	stop := make(chan struct{})
	go func() {
		ChangedPath := spider.FilesChanged()
		for {
			select {
			case path := <-ChangedPath:
				log.Println("file changed:" + path)
			case <-stop:
				return
			}
		}
	}()
	var input int
	fmt.Scan(&input)
	fmt.Println("all paths", spider.AllFiles())
	fmt.Println("all dirs", spider.AllDirs())
	fmt.Println("all files", spider.AllFiles())
	stop <- struct{}{}
	spider.Stop()
}
