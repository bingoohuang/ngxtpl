package ngxtpl

import (
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
	Nacos Nacos `hcl:"nacos"`

	dataSource DataSource
	name       string
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
	case "nacos":
		c.dataSource, err = c.Nacos.Parse()
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

// DecodeCfgFiles decodes the config files to Cfg structs.
func DecodeCfgFiles(cfgFiles []string) (cfgs Cfgs) {
	cfgs = make([]Cfg, len(cfgFiles))

	for i, cfgFile := range cfgFiles {
		if err := hcl.Unmarshal(ReadFile(cfgFile), &cfgs[i]); err != nil {
			panic(err)
		}

		cfgs[i].name = cfgFile

		if err := cfgs[i].Parse(); err != nil {
			panic(err)
		}
	}

	return
}

// Run runs the config.
func (c Cfg) Run() {
	defer c.Tpl.resetTicker()

	logrus.Infof("Start to run config file %s", c.name)

	m, err := c.dataSource.Read()
	if err != nil {
		logrus.Warnf("failed to read template data: %v", err)
		return
	}

	if err := c.Tpl.Execute(m, c.dataSource, c.name); err != nil {
		logrus.Warnf("failed to execute tpl: %v", err)
	}
}

// Cfgs is alias for slice of Cfg.
type Cfgs []Cfg

// Run runs the configs.
func (c Cfgs) Run() {
	loopCfgs := make(Cfgs, 0, len(c))
	for _, cfg := range c {
		cfg.Run()

		if cfg.Tpl.interval > 0 {
			loopCfgs = append(loopCfgs, cfg)
		}
	}

	if len(loopCfgs) == 0 {
		logrus.Infof("finished")
		return
	}

	loopCfgs.loopRun()
}

func (c Cfgs) loopRun() {
	agg := make(chan int)
	singalCtx := SetupSingals()

	for i, cfg := range c {
		go func(c chan int, i int, cfg Cfg) {
			for {
				select {
				case <-singalCtx.Done():
					return
				case <-cfg.Tpl.tiker.C:
					c <- i
				}
			}
		}(agg, i, cfg)
	}

	for {
		select {
		case <-singalCtx.Done():
			logrus.Info("exit by signal")
			return
		case i := <-agg:
			c[i].Run()
		}
	}
}
