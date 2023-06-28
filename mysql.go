package ngxtpl

import (
	"database/sql"
	"log"
	"strings"

	"github.com/bingoohuang/gg/pkg/sqx"
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
	val, err := sqx.SQL{Q: query}.QueryAsString(db)
	if err != nil {
		return "", err
	}

	return val, nil
}

func (t Mysql) Read() (interface{}, error) {
	db, err := sql.Open("mysql", t.DataSourceName)
	if err != nil {
		return nil, err
	}

	defer db.Close()

	result, err := sqx.SQL{Q: t.DataSQL}.QueryAsMaps(db)
	if err != nil {
		return nil, err
	}

	mm := ConvertToMapInterfaceSlice(result)
	for _, m := range mm {
		t.fulfil(db, m)
	}

	out := make(map[string]interface{})
	out[t.DataKey] = mm

	return out, nil
}

// ConvertToMapInterfaceSlice convert slice of map[string]interface{} to slice of map[string]interface{}.
func ConvertToMapInterfaceSlice(m []map[string]string) []map[string]interface{} {
	vv := make([]map[string]interface{}, len(m))

	for i, v := range m {
		vv[i] = ConvertToMapInterface(v)
	}

	return vv
}

// ConvertToMapInterface converts map[string]string to map[string]interface{}.
func ConvertToMapInterface(v map[string]string) map[string]interface{} {
	m := make(map[string]interface{})
	for k, v := range v {
		m[k] = v
	}
	return m
}

func (t Mysql) fulfil(db *sql.DB, m map[string]interface{}) {
	for k, v := range m {
		queryTemplate, ok := t.Sqls[parsePlaceholder(v)]
		if !ok {
			continue
		}

		query, err := TemplateEval(queryTemplate, m)
		if err != nil {
			log.Printf("W! failed to execute template %s, error: %v", queryTemplate, err)
			continue
		}

		sub, err := sqx.SQL{Q: query}.QueryAsMaps(db)
		if err != nil {
			log.Printf("W! failed to execute sql %s, error: %v", query, err)
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
