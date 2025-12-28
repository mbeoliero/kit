package log

import (
	"os"
	"runtime"
	"strconv"
	"strings"
)

func GetPID() string {
	return strconv.FormatInt(int64(os.Getpid()), 10)
}

func GetGID() string {
	defer func() {
		_ = recover()
	}()
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		return "-"
	}
	return strconv.FormatInt(int64(id), 10)
}
