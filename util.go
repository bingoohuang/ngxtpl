package ngxtpl

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

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

// PflagParse parses the plag and check unknown args.
func PflagParse() {
	pflag.Parse()

	if pflag.NArg() > 0 {
		logrus.Errorf("Unknown args %s\n", strings.Join(pflag.Args(), " "))
		pflag.PrintDefaults()
		os.Exit(-1)
	}
}

// CreateDemoCfg  creates a demo config file with name demoHclFile.
func CreateDemoCfg(demoHclFile string) {
	if v, err := os.Stat(demoHclFile); err == nil && !v.IsDir() {
		fmt.Printf("%s exists already\n", demoHclFile)
		os.Exit(0)
	}

	v := ReadPkger(pkger.Include("/assets/cfg.hcl"))
	if err := ioutil.WriteFile(demoHclFile, v, 0644); err != nil {
		panic(err)
	}

	fmt.Printf("%s created\n", demoHclFile)
}

// QueryRows querys the database and returns a slice of map.
func QueryRows(db *sql.DB, query string) ([]map[string]interface{}, error) {
	rows, _ := db.Query(query) // Note: Ignoring errors for brevity
	defer rows.Close()

	cols, _ := rows.Columns()

	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		columns := make([]string, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		m := make(map[string]interface{})
		for i, colName := range cols {
			m[colName] = columns[i]
		}

		results = append(results, m)
	}

	return results, nil
}
