package filex

import (
	"path/filepath"
	"runtime"
)

func CurrentFilePath() string {
	_, filePath, _, ok := runtime.Caller(1)
	if !ok {
		return ""
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return ""
	}

	return absPath
}
