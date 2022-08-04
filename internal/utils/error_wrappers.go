package utils

import (
	"fmt"
	"log"
	"runtime"
	"strconv"
)

// MaybeCrash prints error details (filename, line number) and crashes the program if err != nil
func MaybeCrash(err error) {
	_, filename, line, _ := runtime.Caller(1)
	if err != nil {
		log.Fatalln(PrettifyError(err, filename, strconv.Itoa(line)))
	}
}

// PrettifyError prints error details (filename, line number).
// Caller data can be overridden with optional string arguments ([1]:filename, [2]:line)
func PrettifyError(err error, callerData ...string) string {
	if len(callerData) > 0 {
		return fmt.Sprintf("[error] %s:%s %v", callerData[0], callerData[1], err)
	} else {
		_, filename, line, _ := runtime.Caller(1)
		return fmt.Sprintf("[error] %s:%d %v", filename, line, err)
	}
}
