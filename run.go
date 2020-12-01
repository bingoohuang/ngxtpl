package ngxtpl

import (
	"time"

	"github.com/sirupsen/logrus"
)

// Run runs the config.
func (c Cfg) Run() {
	for {
		m, err := c.dataSource.Read()
		if err != nil {
			logrus.Warnf("failed to read template data: %v", err)
		}

		if err := c.Tpl.Execute(m); err != nil {
			logrus.Warnf("failed to execute tpl: %v", err)
		}

		if c.Tpl.interval == 0 {
			return
		}

		time.Sleep(c.Tpl.interval)
	}
}
