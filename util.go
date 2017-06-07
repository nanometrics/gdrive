package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

func GetDefaultConfigDir() string {
	return filepath.Join(Homedir(), ".godrive")
}

func ConfigFilePath(basePath, name string) string {
	return filepath.Join(basePath, name)
}

func Homedir() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("APPDATA")
	}
	return os.Getenv("HOME")
}

func equal(a, b []string) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func ExitF(format string, a ...interface{}) {
	log.Fatalf(format, a...)
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func writeJson(path string, data interface{}) error {
	tmpFile := path + ".tmp"
	f, err := os.Create(tmpFile)
	if err != nil {
		return err
	}

	err = json.NewEncoder(f).Encode(data)
	f.Close()
	if err != nil {
		os.Remove(tmpFile)
		return err
	}

	return os.Rename(tmpFile, path)
}
