package main

import (
	"os"

	"github.com/bingoohuang/golog"
	"github.com/bingoohuang/ngxtpl"
	"github.com/spf13/pflag"
)

func main() {
	_, _ = golog.SetupLogrus(nil, "", "")

	demoCfg := pflag.BoolP("demo", "", false, "create demo.hcl file")
	configFiles := pflag.StringSliceP("cfgs", "c", nil, "config files")

	ngxtpl.PflagParse()

	if *demoCfg {
		ngxtpl.CreateDemoCfg("./demo.hcl")
		os.Exit(0)
	}

	if len(*configFiles) == 0 {
		pflag.PrintDefaults()
		return
	}

	tpls := ngxtpl.DecodeCfgFiles(*configFiles)
	tpls.Run()
}
