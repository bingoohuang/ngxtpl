package main

import (
	"fmt"
	"github.com/bingoohuang/gg/pkg/ctl"
	"github.com/takama/daemon"
	"os"

	"github.com/bingoohuang/golog"

	"github.com/bingoohuang/ngxtpl"
	"github.com/spf13/pflag"
)

func main() {
	f := pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	initing := f.Bool("init", false, "init sample conf.yaml/ctl and then exit")
	version := f.Bool("version", false, "show version info and exit")
	configFiles := f.StringSliceP("conf", "c", nil, "config files")
	ngxtpl.PflagParse(f, serviceProcess("ngxtpl service"))

	ctl.Config{
		Initing:      *initing,
		PrintVersion: *version,
		VersionInfo:  "ngxtpl v1.0.2",
		ConfTemplate: ngxtpl.ConfBytes,
		ConfFileName: "demo.hcl",
	}.ProcessInit()

	if len(*configFiles) == 0 {
		pflag.PrintDefaults()
		return
	}

	golog.SetupLogrus()
	tpls := ngxtpl.DecodeCfgFiles(*configFiles)
	tpls.Run()
}

func serviceProcess(desc string) []string {
	if len(os.Args) == 1 {
		return nil
	}

	var err error
	var srv daemon.Daemon
	switch c := os.Args[1]; c {
	case "install", "remove", "start", "stop", "status":
		srv, err = daemon.New(os.Args[0], desc, daemon.SystemDaemon)
		if err != nil {
			fmt.Print("Error: ", err)
			os.Exit(1)
		}
	}

	switch c := os.Args[1]; c {
	case "install":
		exitWithMsg(srv.Install(os.Args[2:]...))
	case "remove":
		exitWithMsg(srv.Remove())
	case "start":
		exitWithMsg(srv.Start())
	case "stop": // No need to explicitly stop cron since job will be killed
		exitWithMsg(srv.Stop())
	case "status":
		exitWithMsg(srv.Status())
	case "-h":
		return os.Args
	}

	return os.Args[1:]
}

func exitWithMsg(msg string, err error) {
	if msg != "" {
		fmt.Println(msg)
	}
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
