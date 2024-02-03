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
	fmt.Println(spider.AllPaths())
	go func() {
		for v := range spider.FilesChanged() {
			log.Println("file changed: ", v)
			// fmt.Println("all paths", spider.AllPaths())
			fmt.Println("all files", spider.AllFiles())
			fmt.Println("all dirs", spider.AllDirs())
		}
	}()
	var input int
	fmt.Scan(&input)
	fmt.Println("all paths", spider.AllFiles())
	fmt.Println("all dirs", spider.AllDirs())
	fmt.Println("all files", spider.AllFiles())

	spider.Stop()
}
