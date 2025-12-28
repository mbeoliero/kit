package filex

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCurrentFilePath(t *testing.T) {
	fp := CurrentFilePath()
	filename := filepath.Base(fp)

	assert.Equal(t, "file_test.go", filename)
}
