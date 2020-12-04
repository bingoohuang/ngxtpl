package ngxtpl

import (
	"database/sql"
	"strings"

	"github.com/sirupsen/logrus"

	// import mysql driver
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

// Mysql represents the structure of mysql config.
type Mysql struct {
	DataSourceName string            `hcl:"dataSourceName"`
	DataKey        string            `hcl:"dataKey"`
	DataSQL        string            `hcl:"dataSql"`
	Sqls           map[string]string `hcl:"sqls"`
}

// Parse parses the mysql config.
func (t *Mysql) Parse() (DataSource, error) {
	if t.DataSourceName == "" {
		return nil, errors.Wrapf(ErrCfg, "dataSourceName is required")
	}

	if t.DataSQL == "" {
		return nil, errors.Wrapf(ErrCfg, "dataSql is required")
	}

	if t.DataKey == "" {
		t.DataKey = "data"
	}

	return t, nil
}

func (t *Mysql) Read() (interface{}, error) {
	db, err := sql.Open("mysql", t.DataSourceName)
	if err != nil {
		return nil, err
	}

	defer db.Close()

	results, err := QueryRows(db, t.DataSQL)
	if err != nil {
		return nil, err
	}

	for _, m := range results {
		t.fulfil(db, m)
	}

	out := make(map[string]interface{})
	out[t.DataKey] = results

	return out, nil
}

func (t *Mysql) fulfil(db *sql.DB, m map[string]interface{}) {
	for k, v := range m {
		queryTemplate, ok := t.Sqls[parsePlaceholder(v)]
		if !ok {
			continue
		}

		query, err := TemplateEval(queryTemplate, m)
		if err != nil {
			logrus.Warnf("failed to execute template %s, error: %v", queryTemplate, err)
			continue
		}

		sub, err := QueryRows(db, query)
		if err != nil {
			logrus.Warnf("failed to execute sql %s, error: %v", query, err)
			continue
		}

		m[k] = sub
	}
}

func parsePlaceholder(v interface{}) string {
	s, ok := v.(string)
	if !ok {
		return ""
	}

	if HasBrace(s, "{{", "}}") {
		return strings.TrimSpace(s[2 : len(s)-2])
	}

	return ""
}
