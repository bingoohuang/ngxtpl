package ngxtpl

import (
	"github.com/hashicorp/hcl"
	"github.com/pkg/errors"
)

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

// Cfgs is alias for slice of Cfg.
type Cfgs []Cfg

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

// DecodeCfgFiles decodes the config files to Cfg structs.
func DecodeCfgFiles(cfgFiles []string) (cfgs Cfgs) {
	cfgs = make([]Cfg, len(cfgFiles))

	for i, cfgFile := range cfgFiles {
		var cfg Cfg

		if err := hcl.Decode(&cfg, string(ReadFile(cfgFile))); err != nil {
			panic(err)
		}

		if err := cfg.Parse(); err != nil {
			panic(err)
		}

		cfgs[i] = cfg
	}

	return
}
