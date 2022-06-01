package utils

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func DoOrDie(err error) {
	if err != nil {
		logrus.Fatalf("Fatal error: %+v\n", err)
	}
}

func SetUpLogger(logLevelStr string) error {
	logLevel, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		return errors.Wrapf(err, "unable to parse the specified log level: '%s'", logLevel)
	}
	logrus.SetLevel(logLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.Infof("log level set to '%s'", logrus.GetLevel())
	return nil
}

func DumpJSON(obj interface{}) string {
	bytes, err := json.MarshalIndent(obj, "", "  ")
	DoOrDie(err)
	return string(bytes)
}

func ReadJsonFromFile(obj interface{}, path string) error {
	bytes, err := ReadFileBytes(path)
	if err != nil {
		return err
	}
	return errors.Wrapf(json.Unmarshal(bytes, obj), "unable to unmarshal json at %s", path)
}

func WriteJsonToFile(obj interface{}, path string) error {
	content := DumpJSON(obj)
	return WriteFile(path, content, 0644)
}

func PrintJSON(obj interface{}) {
	fmt.Printf("%s\n", DumpJSON(obj))
}

func DoesFileExist(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		panic(errors.Wrapf(err, "unable to determine if file %s exists", path))
	}
}

// WriteFile wraps calls to ioutil.WriteFile, ensuring that errors are wrapped in a stack trace
func WriteFile(filename string, contents string, perm fs.FileMode) error {
	return errors.Wrapf(ioutil.WriteFile(filename, []byte(contents), perm), "unable to write file %s", filename)
}

// ReadFile wraps calls to ioutil.ReadFile, ensuring that errors are wrapped in a stack trace
func ReadFile(filename string) (string, error) {
	bytes, err := ioutil.ReadFile(filename)
	return string(bytes), errors.Wrapf(err, "unable to read file %s", filename)
}

// ReadFileBytes wraps calls to ioutil.ReadFile, ensuring that errors are wrapped in a stack trace
func ReadFileBytes(filename string) ([]byte, error) {
	bytes, err := ioutil.ReadFile(filename)
	return bytes, errors.Wrapf(err, "unable to read file %s", filename)
}
