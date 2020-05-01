package util

import (
	"os"
)

func CreateDir(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(path, 0755)
		if errDir != nil {
			return errDir
		}
	}
	return nil
}
