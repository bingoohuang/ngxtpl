package ngxtpl

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/hashicorp/hcl"
	"github.com/pkg/errors"
)

// KeyReader read by key.
type KeyReader interface {
	Get(key string) (string, error)
}

// DataSource defines the data source interface for reading the data.
type DataSource interface {
	Read() (interface{}, error)
}

// HTTPSource defines the data source interface for reading the data from http.
type HTTPSource struct {
	Address string
}

func (h HTTPSource) Read() (interface{}, error) {
	content, err := HTTPGetStr(h.Address)
	if err != nil {
		return nil, err
	}

	return JSONDecode(content)
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

	return c.Tpl.Parse(c.dataSource)
}

// ErrCfg defines the error type of bad config.
var ErrCfg = errors.New("Unknown dataSource(should be redis, mysql or http address)")

func (c *Cfg) parseDataSource() (err error) {
	switch v := c.Tpl.DataSource; v {
	case "redis":
		c.dataSource, err = c.Redis.Parse()
		return err
	case "mysql":
		c.dataSource, err = c.Mysql.Parse()
		return err
	default:
		if IsHTTPAddress(v) {
			c.dataSource, err = &HTTPSource{Address: v}, nil
			return
		}
	}

	return errors.Wrapf(ErrCfg,
		"Unknown dataSource %s, it should be redis or mysql", c.Tpl.DataSource)
}

func (c *Cfg) tikerC() <-chan time.Time {
	if c.Tpl.tiker != nil {
		return c.Tpl.tiker.C
	}

	return nil
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

// Run runs the config.
func (c Cfg) Run() {
	defer c.Tpl.resetTicker()

	m, err := c.dataSource.Read()
	if err != nil {
		logrus.Warnf("failed to read template data: %v", err)
		return
	}

	if err := c.Tpl.Execute(m, c.dataSource); err != nil {
		logrus.Warnf("failed to execute tpl: %v", err)
	}
}

// Cfgs is alias for slice of Cfg.
type Cfgs []Cfg

// Run runs the configs.
func (c Cfgs) Run() {
	agg := make(chan int)
	aggSize := 0

	singalCtx := SetupSingals()

	for i, cfg := range c {
		if cfg.Tpl.interval == 0 {
			continue
		}

		go func(c chan int, i int) {
			for {
				select {
				case <-singalCtx.Done():
					return
				case <-cfg.tikerC():
					agg <- i
				}
			}
		}(agg, i)

		aggSize++
	}

	for _, cfg := range c {
		cfg.Run()
	}

	if aggSize == 0 {
		logrus.Infof("finished")
		return
	}

	for {
		select {
		case <-singalCtx.Done():
			logrus.Info("exit by signal")
			return
		case i := <-agg:
			cfg := c[i]
			cfg.Run()
		}
	}
}
