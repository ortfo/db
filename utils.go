package main

import (
	"os"
	"io/ioutil"
)

func ReadFile(filepath string) string {
    file, err := os.Open(filepath)
    if err != nil {
        panic(err)
    }
    defer file.Close()
    b, err := ioutil.ReadAll(file)
    return string(b)
}
