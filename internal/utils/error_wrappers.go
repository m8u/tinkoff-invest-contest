package utils

import (
	"fmt"
	"log"
	"runtime"
	"strconv"
)

// MaybeCrash выводит подробности об ошибке и завершает программу с кодом 1
// если ошибка != nil
func MaybeCrash(err error) {
	_, filename, line, _ := runtime.Caller(1)
	if err != nil {
		log.Fatalln(PrettifyError(err, filename, strconv.Itoa(line)))
	}
}

func PrettifyError(err error, callerData ...string) string {
	if callerData[0] == "" {
		_, filename, line, _ := runtime.Caller(1)
		return fmt.Sprintf("[error] %s:%d %v", filename, line, err)
	} else {
		return fmt.Sprintf("[error] %s:%s %v", callerData[0], callerData[1], err)
	}
}
