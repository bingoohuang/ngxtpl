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
	KVSql          string            `hcl:"kvSql"`
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

// Get gets the value of key from mysql.
func (t Mysql) Get(key string) (string, error) {
	if t.KVSql == "" {
		return "", errors.Wrapf(ErrCfg, "kvSql is not set")
	}

	db, err := sql.Open("mysql", t.DataSourceName)
	if err != nil {
		return "", err
	}

	defer db.Close()

	query := strings.ReplaceAll(t.KVSql, "{{key}}", key)
	results, cols, err := QueryRows(db, query, 1)
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "", errors.Wrapf(ErrCfg, "no value found")
	}

	return results[0][cols[0]].(string), nil
}

func (t Mysql) Read() (interface{}, error) {
	db, err := sql.Open("mysql", t.DataSourceName)
	if err != nil {
		return nil, err
	}

	defer db.Close()

	results, _, err := QueryRows(db, t.DataSQL, 0)
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

func (t Mysql) fulfil(db *sql.DB, m map[string]interface{}) {
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

		sub, _, err := QueryRows(db, query, 0)
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
