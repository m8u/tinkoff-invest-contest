package utils

import (
	"fmt"
	"log"
	"runtime"
)

// MaybeCrash выводит подробности об ошибке и завершает программу с кодом 1
// если ошибка != nil
func MaybeCrash(err error) {
	if err != nil {
		log.Fatalln(PrettifyError(err))
	}
}

func PrettifyError(err error) string {
	_, filename, line, _ := runtime.Caller(1)
	return fmt.Sprintf("[error] %s:%d %v", filename, line, err)
}
