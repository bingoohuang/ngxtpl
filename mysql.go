package ngxtpl

import (
	// import mysql driver
	_ "github.com/go-sql-driver/mysql"
)

// Mysql represents the structure of mysql config.
type Mysql struct {
	DataSourceName string `hcl:"dataSourceName"`
	UpstreamsTable string `hcl:"upstreamsTable"`
	ServersTable   string `hcl:"serversTable"`
}

// Parse parses the mysql config.
func (t *Mysql) Parse() (DataSource, error) {
	return nil, nil
}
