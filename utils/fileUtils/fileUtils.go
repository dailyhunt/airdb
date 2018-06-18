package fileUtils

import (
	"os"
	"fmt"
	"github.com/pkg/errors"
)

// DirExists checks if a path exists and is a directory.
//func DirExists(path string) (bool, error) {
//	fi, err := os.Stat(path)
//	if err == nil && fi.IsDir() {
//		return true, nil
//	}
//	if os.IsNotExist(err) {
//		return false, nil
//	}
//	return false, err
//}

// IsDir checks if a given path is a directory.
func IsDir(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fi.IsDir(), nil
}

func AssertDir(path string) (error) {
	value, err := IsDir(path)
	if err != nil {
		return err
	}

	if !value {
		return errors.New(path + " is not a directory")
	}

	return nil
}

// IsEmpty checks if a given file or directory is empty.
func IsEmpty(path string) (bool, error) {
	if b, _ := Exists(path); !b {
		return false, fmt.Errorf("%q path does not exist", path)
	}
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if fi.IsDir() {
		f, err := os.Open(path)
		if err != nil {
			return false, err
		}
		defer f.Close()
		list, err := f.Readdir(-1)
		return len(list) == 0, nil
	}
	return fi.Size() == 0, nil
}

func AssertEmpty(path string) (error) {
	value, err := IsEmpty(path)
	if err != nil {
		return err
	}

	if !value {
		return errors.New(path + " is not empty")
	}

	return nil
}

// Check if a file or directory exists.
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func AssertExists(path string) (error) {
	value, err := Exists(path)
	if err != nil {
		return err
	}

	if !value {
		return errors.New(path + " does not exist")
	}

	return nil
}