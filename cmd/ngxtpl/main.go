package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/bingoohuang/ngxtpl"
	"github.com/markbates/pkger"
	"github.com/sirupsen/logrus"

	"github.com/spf13/pflag"
)

func main() {
	demoCfg := pflag.BoolP("demo", "", false, "create demo.hcl file")
	cfg := pflag.StringP("cfg", "c", "", "config file")
	pflag.Parse()

	if pflag.NArg() > 0 {
		logrus.Errorf("Unknown args %s\n", strings.Join(pflag.Args(), " "))
		pflag.PrintDefaults()
		os.Exit(-1)
	}

	if *demoCfg {
		createDemoCfg()
		os.Exit(0)
	}

	if *cfg == "" {
		pflag.PrintDefaults()
		return
	}

	ngxtpl.DecodeCfgFile(*cfg).Run()
}

func createDemoCfg() {
	const demoHcl = "./demo.hcl"
	if v, err := os.Stat(demoHcl); err == nil && !v.IsDir() {
		fmt.Printf("%s exists already\n", demoHcl)
		os.Exit(0)
	}

	v := ngxtpl.ReadPkger(pkger.Include("/assets/cfg.hcl"))
	if err := ioutil.WriteFile(demoHcl, v, 0644); err != nil {
		panic(err)
	}

	fmt.Printf("%s created\n", demoHcl)
}
