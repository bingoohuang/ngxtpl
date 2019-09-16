package main

import (
	"database/sql"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/bingoohuang/gou/lo"
	"github.com/bingoohuang/gou/str"
	"github.com/bingoohuang/sqlmore"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gobars/cmd"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type App struct {
	Servers  []string
	Sql      string
	Db       *sql.DB
	Interval time.Duration
	File     string
	Reload   string
}

func main() {
	app := &App{Servers: make([]string, 0)}
	app.init()

	logrus.Infof("init app %+v\n", app)

	t := time.NewTicker(app.Interval)
	defer t.Stop()

	for {
		app.refreshServers()
		<-t.C
	}
}

func (a *App) init() {
	pflag.StringP("file", "f", "servers.conf", "upstream servers list file")
	pflag.StringP("sql", "s", "select server, port from t_server where state=1", "query sql to get server's list")
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
	db := more.MustOpen()
	db.SetConnMaxLifetime(1 * time.Second)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(0)

	a.Db = db

	interval := viper.GetString("interval")
	if du, err := time.ParseDuration(interval); err != nil {
		panic(err)
	} else {
		a.Interval = du
	}

	if a.Interval < 0 {
		a.Interval = 15 * time.Second
	}

	sqlString := viper.GetString("sql")
	if typ, yes := sqlmore.IsQuerySQL(sqlString); !(yes && typ == "SELECT") {
		logrus.Panicf("%s is not a SELECT SQL!\n", sqlString)
	}

	result := sqlmore.ExecSQL(db, sqlString, 1000, "")
	if result.Error != nil {
		logrus.Panicf("execute sql %s error %v!\n", result.Error)
	}

	a.Sql = sqlString
	a.File = viper.GetString("file")
	a.Reload = viper.GetString("reload")
}
func (a *App) refreshServers() {
	result := sqlmore.ExecSQL(a.Db, a.Sql, 1000, "")
	if result.Error != nil {
		logrus.Errorf("execute sql error %v\n", result.Error)
		return
	}

	if len(result.Rows) == 0 {
		logrus.Errorf("no servers available\n")
		return
	}

	servers := make([]string, len(result.Rows))
	for i, row := range result.Rows {
		if len(row) != 2 {
			logrus.Errorf("result does not contains only two columns\n")
			continue
		}

		server := row[0]
		port := row[1]

		servers[i] = server + str.If(port != "", ":"+port, "") + ";"
	}
	sort.Strings(servers)

	if reflect.DeepEqual(a.Servers, servers) {
		return
	}

	a.Servers = servers

	logrus.Infof("servers changes detected\n")

	contents := []byte(strings.Join(servers, "\n"))
	if err := ioutil.WriteFile(a.File, contents, 0644); err != nil {
		logrus.Errorf("write file %s error %v\n", a.File, err)
		return
	}

	if a.Reload != "" {
		_, status := cmd.Bash(a.Reload, cmd.Timeout(10*time.Second))
		logrus.Infof("reload status %v\n", status)
	}
}
