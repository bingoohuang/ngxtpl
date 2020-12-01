package ngxtpl

import (
	"github.com/hashicorp/hcl"
	"github.com/pkg/errors"
)

// Upstream defines the data structure for nginx upstream.
type Upstream struct {
	Name    string   `json:"name"`
	Servers []Server `json:"servers"`
}

// Server struct by http://nginx.org/en/docs/http/ngx_http_upstream_module.html
type Server struct {
	Address     string `gorm:"address" json:"address"`
	Port        int    `gorm:"port" json:"port"`
	Weight      int    `gorm:"weight" json:"weight"`            // eg. weight=5， sets the weight of the server.
	MaxConns    int    `gorm:"max_conns" json:"maxConns"`       // Default value is zero, meaning there is no limit.
	MaxFails    int    `gorm:"max_fails" json:"maxFails"`       // By default, the number of unsuccessful attempts is set to 1
	FailTimeout string `gorm:"fail_timeout" json:"failTimeout"` // By default, the parameter is set to 10 seconds.
	Backup      bool   `gorm:"backup" json:"backup"`
	SlowStart   string `gorm:"slow_start" json:"slowStart"` // Default value is zero, i.e. slow start is disabled.
}

// DataSource defines the data source interface for reading the template data.
type DataSource interface {
	Read() (interface{}, error)
}

// Cfg represents the root structure of the config.
type Cfg struct {
	Tpl   Tpl   `hcl:"tpl"`
	Redis Redis `hcl:"redis"`
	Mysql Mysql `hcl:"mysql"`

	dataSource DataSource
}

// Parse parses the config.
func (c *Cfg) Parse() error {
	if err := c.parseDataSource(); err != nil {
		return err
	}

	return c.Tpl.Parse()
}

// ErrCfg defines the error type of bad config.
var ErrCfg = errors.New("Unknown dataSource(should be redis or mysql)")

func (c *Cfg) parseDataSource() (err error) {
	switch c.Tpl.DataSource {
	case "redis":
		c.dataSource, err = c.Redis.Parse()
		return err
	case "mysql":
		c.dataSource, err = c.Mysql.Parse()
		return err
	}

	return errors.Wrapf(ErrCfg, "Unknown dataSource %s, it should be redis or mysql", c.Tpl.DataSource)
}

// DecodeCfgFile decodes the config file to Cfg struct.
func DecodeCfgFile(cfgFile string) (cfg Cfg) {
	if err := hcl.Decode(&cfg, string(ReadFile(cfgFile))); err != nil {
		panic(err)
	}

	if err := cfg.Parse(); err != nil {
		panic(err)
	}

	return
}
