package spider

import (
	"fmt"
	"regexp"
)

const (
	threadnum = 10
	isDebug   = false
)

var regHan = regexp.MustCompile(`[\p{Han}]+`)
var regNumber = regexp.MustCompile(`[0-9]+`)

func PrintInfo(msg interface{}) {
	if isDebug {
		fmt.Println(msg)
	}
}
