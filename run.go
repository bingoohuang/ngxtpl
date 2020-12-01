package ngxtpl

import (
	"time"

	"github.com/sirupsen/logrus"
)

// Run runs the config.
func (c Cfg) Run() {
	m, err := c.dataSource.Read()
	if err != nil {
		logrus.Warnf("failed to read template data: %v", err)
	}

	if err := c.Tpl.Execute(m); err != nil {
		logrus.Warnf("failed to execute tpl: %v", err)
	}
}

// Run runs the configs.
func (c Cfgs) Run() {
	tikers := make([]*time.Ticker, len(c))

	agg := make(chan int)
	aggSize := 0

	for i, cfg := range c {
		if cfg.Tpl.interval == 0 {
			continue
		}

		tikers[i] = time.NewTicker(cfg.Tpl.interval)
		go func(c chan int, i int) {
			for range tikers[i].C {
				agg <- i
			}
		}(agg, i)

		aggSize++
	}

	for _, cfg := range c {
		cfg.Run()
	}

	if aggSize > 0 {
		for i := range agg {
			c[i].Run()
		}
	}
}
