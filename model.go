package nginxtemplate

import (
	"fmt"
)

// Server struct by http://nginx.org/en/docs/http/ngx_http_upstream_module.html
type Server struct {
	Address     string `gorm:"address"`
	Port        int    `gorm:"port"`
	Weight      int    `gorm:"weight"`       // eg. weight=5， sets the weight of the server, by default, 1.
	MaxConns    int    `gorm:"max_conns"`    // Default value is zero, meaning there is no limit.
	MaxFails    int    `gorm:"max_fails"`    // By default, the number of unsuccessful attempts is set to 1
	FailTimeout string `gorm:"fail_timeout"` // By default, the parameter is set to 10 seconds.
	Backup      bool   `gorm:"backup"`
	SlowStart   string `gorm:"slow_start"` // Default value is zero, i.e. slow start is disabled.
}

func (Server) TableName() string { return "t_server" }

func (s Server) ServerLine() string {
	line := fmt.Sprintf("%s:%d", s.Address, s.Port)
	if s.Weight != 1 {
		line += fmt.Sprintf(" weight=%d", s.Weight)
	}

	if s.MaxConns != 0 {
		line += fmt.Sprintf(" max_conns=%d", s.MaxConns)
	}

	if s.MaxFails != 1 {
		line += fmt.Sprintf(" max_fails=%d", s.MaxFails)
	}

	if s.FailTimeout != "10s" {
		line += fmt.Sprintf(" fail_timeout=%s", s.FailTimeout)
	}

	if s.Backup {
		line += " backup"
	}

	if s.SlowStart != "0" {
		line += fmt.Sprintf(" slow_start=%s", s.SlowStart)
	}

	return line + ";"
}
