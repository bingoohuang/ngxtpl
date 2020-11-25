package main

import (
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/bingoohuang/ngxtpl"

	"github.com/gobars/cmd"

	"github.com/jinzhu/gorm"

	"github.com/sirupsen/logrus"

	"github.com/bingoohuang/gou/lo"
	"github.com/bingoohuang/sqlmore"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type App struct {
	Servers  string
	Db       *gorm.DB
	Interval time.Duration
	File     string
	Reload   string
}

func main() {
	app := &App{}
	app.init()

	logrus.Infof("init app %+v\n", app)

	t := time.NewTicker(app.Interval)
	defer t.Stop()

	for {
		servers := ngxtpl.RefreshServers(app.Db)
		if app.Servers == servers {
			continue
		}

		app.Servers = servers
		if err := ioutil.WriteFile(app.File, []byte(servers), 0644); err != nil {
			logrus.Errorf("write file %s error %v\n", app.File, err)
			return
		}

		if app.Reload != "" {
			_, status := cmd.Bash(app.Reload, cmd.Timeout(10*time.Second))
			logrus.Infof("reload status %v\n", status)
		}

		<-t.C
	}
}

func (a *App) init() {
	pflag.StringP("file", "f", "servers.conf", "upstream servers list file")
	pflag.StringP("mysql", "m", "user:pass@tcp(127.0.0.1:3306)/mydb", "mysql data source name")
	pflag.StringP("interval", "i", "15s", "refresh interval")
	pflag.StringP("reload", "r", "nginx -s reload", "nginx reload command")
	pflag.Parse()

	args := pflag.Args()
	if len(args) > 0 {
		logrus.Errorf("Unknown args %s\n", strings.Join(args, " "))
		pflag.PrintDefaults()
		os.Exit(-1)
	}

	// 从当前位置读取config.toml配置文件
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	lo.Err(viper.ReadInConfig())

	lo.Err(viper.BindPFlags(pflag.CommandLine))

	more := sqlmore.NewSQLMore("mysql", viper.GetString("mysql"))
	gdb, err := more.GormOpen()
	if err != nil {
		logrus.Panicf("fail to open db error %v\n", err)
	}

	db := gdb.DB()
	db.SetConnMaxLifetime(1 * time.Second)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(0)

	a.Db = gdb

	interval := viper.GetString("interval")
	if du, err := time.ParseDuration(interval); err != nil {
		logrus.Panicf("fail to parse interval %s, error %v\n", interval, err)
	} else {
		a.Interval = du
	}

	if a.Interval < 0 {
		a.Interval = 15 * time.Second
	}

	a.File = viper.GetString("file")
	a.Reload = viper.GetString("reload")
}
