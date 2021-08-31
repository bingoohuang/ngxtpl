package main

import (
	"github.com/bingoohuang/gg/pkg/ctl"
	"os"

	"github.com/bingoohuang/golog"

	"github.com/bingoohuang/ngxtpl"
	"github.com/spf13/pflag"
)

func main() {
	f := pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	initing := f.Bool("init", false, "init sample conf.yaml/ctl and then exit")
	version := f.Bool("version", false, "show version info and exit")
	configFiles := f.StringSliceP("conf", "c", []string{"ngxtpl.hcl"}, "config files")
	ngxtpl.PflagParse(f, os.Args[1:])

	ctl.Config{
		Initing:      *initing,
		PrintVersion: *version,
		InitFiles:    ngxtpl.InitAssets,
	}.ProcessInit()

	if len(*configFiles) == 0 {
		pflag.PrintDefaults()
		return
	}

	golog.SetupLogrus()
	tpls := ngxtpl.DecodeCfgFiles(*configFiles)
	tpls.Run()
}
