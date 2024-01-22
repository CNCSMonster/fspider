//go:build log

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

func (s *Spider) String() string {
	s.RLock()
	defer s.RUnlock()
	sb := strings.Builder{}
	sb.WriteString("files:\n")
	for k, v := range s.files {
		sb.WriteString(fmt.Sprintf("%s: %v\n", k, v))
	}
	sb.WriteString("dirs:\n")
	for k, v := range s.dirs {
		sb.WriteString(fmt.Sprintf("%s: %v\n", k, v))
	}
	return sb.String()
}
