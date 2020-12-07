package main

import (
	"fmt"
	"os"

	"github.com/bingoohuang/golog"

	"github.com/bingoohuang/ngxtpl"
	"github.com/spf13/pflag"
)

func main() {
	demoCfg := pflag.BoolP("demo", "", false, "create demo.hcl file")
	version := pflag.BoolP("version", "v", false, "create demo.hcl file")
	configFiles := pflag.StringSliceP("cfgs", "c", nil, "config files")

	ngxtpl.PflagParse()

	if *demoCfg {
		ngxtpl.CreateDemoCfg("./demo.hcl")
		os.Exit(0)
	}

	if *version {
		fmt.Println("v1.0.1 released at 2020-12-07 16:03:12")
		os.Exit(0)
	}

	if len(*configFiles) == 0 {
		pflag.PrintDefaults()
		return
	}

	_, _ = golog.SetupLogrus(nil, "", "")
	tpls := ngxtpl.DecodeCfgFiles(*configFiles)
	tpls.Run()
}
