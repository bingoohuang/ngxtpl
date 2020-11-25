package ngxtpl

import (
	"sort"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

func RefreshServers(gdb *gorm.DB) string {
	var serverBeans []Server
	db := gdb.Find(&serverBeans)
	if db.Error != nil {
		logrus.Errorf("execute sql error %v\n", db.Error)
		return ""
	}

	if len(serverBeans) == 0 {
		logrus.Errorf("no servers available\n")
		return ""
	}

	servers := make([]string, len(serverBeans))
	for i, b := range serverBeans {
		servers[i] = b.ServerLine()
	}
	sort.Strings(servers)

	contents := strings.Join(servers, "\n")

	logrus.Infof("servers changes detected\n")

	return contents
}
