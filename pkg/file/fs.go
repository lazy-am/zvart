package file

import (
	"errors"
	"os"
)

func FileExists(fn string) bool {
	_, err := os.Stat(fn)
	return !errors.Is(err, os.ErrNotExist)
}
