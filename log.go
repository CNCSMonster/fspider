//go:build !release

package fspider

import (
	"fmt"
	"strings"
)

func init() {
	fmt.Println("use log mod fspider")
}

func Log(arg ...any) {
	fmt.Println(arg...)
}

func (s *spiderImpl) String() string {
	sb := strings.Builder{}
	sb.WriteString("files:\n")
	for _, v := range s.AllFiles() {
		sb.WriteString(fmt.Sprintf("%v\n", v))
	}
	sb.WriteString("dirs:\n")
	for _, v := range s.AllDirs() {
		sb.WriteString(fmt.Sprintf("%v\n", v))
	}
	return sb.String()
}
