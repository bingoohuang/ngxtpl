package ngxtpl

import (
	"io/ioutil"

	"github.com/markbates/pkger"
)

// ReadPkger reads the content of pkger file.
func ReadPkger(file string) []byte {
	f, err := pkger.Open(file)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	d, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	return d
}

// ReadFile reads the file content of file with name filename.
// or panic if error happens.
func ReadFile(filename string) []byte {
	d, err := ReadFileE(filename)
	if err != nil {
		panic(err)
	}

	return d
}

// ReadFileE reads the file content of file with name filename.
func ReadFileE(filename string) ([]byte, error) {
	d, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return d, nil
}
