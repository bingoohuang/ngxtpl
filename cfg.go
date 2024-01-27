package ngxtpl

import (
	"log"

	"github.com/hashicorp/hcl"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// KeyWriter writes key and value.
type KeyWriter interface {
	Write(key, value string) error
}

// Result ...
type Result struct {
	Time       string `json:"time"`
	Old        string `json:"old,omitempty"`
	New        string `json:"new,omitempty"`
	Error      string `json:"error,omitempty"`
	StatusCode int    `json:"statusCode"`
}

// ResultWriter writes error.
type ResultWriter interface {
	WriteResult(Result) error
}

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
	dataSource DataSource
	Redis      *Redis `hcl:"redis"`
	Mysql      *Mysql `hcl:"mysql"`
	Nacos      *Nacos `hcl:"nacos"`

	name string

	Tpl Tpl `hcl:"tpl"`
}

// Parse parses the config.
func (c *Cfg) Parse() error {
	if err := c.parseDataSource(); err != nil {
		return err
	}

	return c.Tpl.Parse()
}

// ErrCfg defines the error type of bad config.
var ErrCfg = errors.New("Errors in the config file")

func (c *Cfg) parseDataSource() (err error) {
	if c.Nacos != nil {
		// nacos 只要配置，就初始化，可能需要注册自己
		c.Nacos.Parse()
	}

	switch v := c.Tpl.DataSource; v {
	case "redis":
		c.dataSource, err = c.Redis.Parse()
	case "mysql":
		c.dataSource, err = c.Mysql.Parse()
	case "nacos":
		c.dataSource, err = c.Nacos.Parse()
	default:
		switch {
		case IsHTTPAddress(v):
			c.dataSource, err = &HTTPSource{Address: v}, nil
		case c.Redis != nil:
			c.dataSource, err = c.Redis.Parse()
		case c.Mysql != nil:
			c.dataSource, err = c.Mysql.Parse()
		case c.Nacos != nil:
			c.dataSource, err = c.Nacos.Parse()
		default:
			err = errors.Wrapf(ErrCfg, "Unknown dataSource %s", c.Tpl.DataSource)
		}
	}

	return err
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

	log.Printf("Start to run config file %s", c.name)
	m, err := c.dataSource.Read()
	if err != nil {
		log.Printf("W! failed to read template data: %v", err)
		return
	}

	result := Result{}
	if err := c.Tpl.Execute(m, c.dataSource, c.name, &result); err != nil {
		log.Printf("W! failed to execute tpl: %v", err)
		result.StatusCode = 400
		result.Error = err.Error()
	}

	if ew, ok := c.dataSource.(ResultWriter); ok {
		_ = ew.WriteResult(result)
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
		log.Printf("finished")
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
				case <-cfg.Tpl.ticker.C:
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
