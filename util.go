package ngxtpl

import (
	"context"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

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

// SetupSingals setup the signal.
func SetupSingals(sig ...os.Signal) context.Context {
	return SetupSingalsWithContext(context.Background(), sig...)
}

// SetupSingalsWithContext setup the signal.
func SetupSingalsWithContext(parent context.Context, sig ...os.Signal) context.Context {
	ch := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(parent)

	if len(sig) == 0 {
		sig = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}

	signal.Notify(ch, sig...)

	go func() {
		<-ch
		cancel()
	}()

	return ctx
}
