package main

import (
	"fmt"
	"github.com/takama/daemon"
	"log"
	"os"

	"github.com/bingoohuang/golog"

	"github.com/bingoohuang/ngxtpl"
	"github.com/spf13/pflag"
)

func main() {
	f := pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	demoCfg := f.BoolP("demo", "", false, "create demo.hcl file")
	version := f.BoolP("version", "v", false, "create demo.hcl file")
	configFiles := f.StringSliceP("cfgs", "c", nil, "config files")
	ngxtpl.PflagParse(f, serviceProcess("ngxtpl service"))

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

	golog.SetupLogrus()
	tpls := ngxtpl.DecodeCfgFiles(*configFiles)
	tpls.Run()
}

func serviceProcess(desc string) []string {
	if len(os.Args) == 1 {
		return nil
	}

	srv, err := daemon.New(os.Args[0], desc, daemon.SystemDaemon)
	if err != nil {
		log.Println("Error: ", err)
		os.Exit(1)
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
		log.Printf("Usage: %s install args... | remove | start | stop | status | run args...", os.Args[0])
		os.Exit(0)
		return nil
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
