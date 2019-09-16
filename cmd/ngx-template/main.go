package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/bingoohuang/gou/str"
	"github.com/bingoohuang/sqlmore"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type App struct {
	Servers  []string
	Sql      string
	Db       *sql.DB
	Interval time.Duration
	File     string
}

func main() {
	app := &App{Servers: make([]string, 0)}
	app.init()

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
	pflag.Parse()
	_ = viper.BindPFlags(pflag.CommandLine)

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
		panic(sqlString + " is not a SELECT SQL!\n")
	}

	result := sqlmore.ExecSQL(db, sqlString, 1000, "")
	if result.Error != nil {
		panic(result.Error)
	}

	a.Sql = sqlString
	a.File = viper.GetString("file")
}
func (a *App) refreshServers() {
	result := sqlmore.ExecSQL(a.Db, a.Sql, 1000, "")
	if result.Error != nil {
		fmt.Printf("execute sql error %v\n", result.Error)
		return
	}

	if len(result.Rows) == 0 {
		fmt.Printf("no servers available\n")
		return
	}

	servers := make([]string, len(result.Rows))
	for i, row := range result.Rows {
		if len(row) != 2 {
			fmt.Printf("result does not contains two columns\n")
			return
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

	fmt.Printf("servers changes detected\n")

	contents := []byte(strings.Join(servers, "\n"))
	if err := ioutil.WriteFile(a.File, contents, 0644); err != nil {
		fmt.Printf("write file %s error %v\n", a.File, err)
		return
	}
}

