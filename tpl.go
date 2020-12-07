package ngxtpl

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/gobars/cmd"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Tpl represents a tpl config section.
type Tpl struct {
	DataSource  string `hcl:"dataSource"`
	Interval    string `hcl:"interval"`
	TplSource   string `hcl:"tplSource"`
	Destination string `hcl:"destination"`
	Perms       int    `hcl:"perms"`
	Command     string `hcl:"command"`

	interval time.Duration
	tiker    *time.Ticker
}

// Execute executes the template.
func (t *Tpl) Execute(data interface{}, ds DataSource) error {
	var out bytes.Buffer

	sourceContent, err := t.parseSource(ds)
	if err != nil {
		return err
	}

	source, err := template.New("TplSource").Parse(sourceContent)
	if err != nil {
		return errors.Wrapf(ErrCfg, "TplSource is invalid. "+
			"it should be a template file or direct template content string")
	}

	if err := source.Execute(&out, data); err != nil {
		return err
	}

	newContent := out.Bytes()
	oldContent, err := t.readDestination()
	if err != nil {
		return err
	}

	if bytes.Equal(newContent, oldContent) {
		return nil
	}

	logrus.Infof("new content %s", string(newContent))
	logrus.Infof("old content %s", string(oldContent))

	if err := t.writeDestination(newContent); err != nil {
		logrus.Errorf("failed to write destination %s err: %v", t.Destination, err)
		return nil
	}

	return t.executeCommand()
}

// Parse parses and validates the template.
func (t *Tpl) Parse(ds DataSource) error {
	if err := t.parseInterval(); err != nil {
		return err
	}

	if _, err := t.parseSource(ds); err != nil {
		return err
	}

	return t.parseDestination()
}

func (t *Tpl) parseInterval() error {
	if t.Interval == "" {
		t.Interval = "0"
	}

	v, err := time.ParseDuration(t.Interval)
	if err != nil {
		return err
	}

	t.interval = v

	if t.interval > 0 {
		t.tiker = time.NewTicker(t.interval)
	}

	return nil
}

func (t *Tpl) resetTicker() {
	if t.tiker != nil {
		t.tiker.Reset(t.interval)
	}
}

func (t *Tpl) parseSource(ds DataSource) (string, error) {
	if t.TplSource == "" {
		return "", errors.Wrapf(ErrCfg, "source is empty")
	}

	return loadSourceContent(t.TplSource, ds)
}

func loadSourceContent(source string, ds DataSource) (string, error) {
	// 1. 尝试从文件读
	if stat, err := os.Stat(source); err == nil && !stat.IsDir() {
		return ReadFileStrE(source)
	}

	// 2. 尝试从datasource读
	const dataSourcePrefix = "dataSource:"
	if strings.HasPrefix(source, dataSourcePrefix) {
		dataSourceKey := source[len(dataSourcePrefix):]
		if kr, ok := ds.(KeyReader); ok {
			return kr.Get(dataSourceKey)
		}
	}

	// 3. 尝试从http读
	if IsHTTPAddress(source) {
		return HTTPGetStr(source)
	}

	// 4. 直接当做模板内容使用
	return source, nil
}

func (t *Tpl) parseDestination() error {
	if t.Destination == "" {
		return nil
	}

	if t.Perms == 0 {
		t.Perms = 0644
	}

	dir := filepath.Dir(t.Destination)
	_, err := os.Stat(dir)
	if err == nil {
		return nil
	}

	if IsHTTPAddress(t.Destination) {
		return nil
	}

	return errors.Wrapf(ErrCfg, "Destination is invalid. "+
		"it should be valid file or http addr. error: %v", err)
}

func (t *Tpl) readDestination() ([]byte, error) {
	if t.Destination == "stdout" {
		return nil, nil
	}

	f, err := ReadFileE(t.Destination)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
	}

	return f, err
}

func (t *Tpl) writeDestination(content []byte) error {
	if t.Destination == "" {
		return nil
	}

	if IsHTTPAddress(t.Destination) {
		resp, err := HTTPPost(t.Destination, content)
		if err != nil {
			return err
		}

		logrus.Infof("POST %s response %s", t.Destination, string(resp))
		return nil
	}

	return ioutil.WriteFile(t.Destination, content, os.FileMode(t.Perms))
}

func (t *Tpl) executeCommand() error {
	if t.Command == "" {
		return nil
	}

	_, status := cmd.Bash(t.Command)
	if status.Exit == 0 {
		logrus.Infof("exec command %s successfully", t.Command)
	} else {
		logrus.Infof("exec command %s failed with exit code %d", t.Command, status.Exit)
	}

	if len(status.Stdout) > 0 {
		logrus.Infof("%s", strings.Join(status.Stdout, "\n"))
	}

	if len(status.Stderr) > 0 {
		logrus.Errorf("%s", strings.Join(status.Stderr, "\n"))
	}

	return nil
}

//  MapInt returns the int value associated with given key in the map.
func MapInt(m map[string]interface{}, key string, defaultValue int) int {
	if len(m) == 0 {
		return defaultValue
	}

	v, ok := m[key]
	if ok && v != "" {
		f, err := strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
		if err != nil {
			return defaultValue
		}

		return int(f)
	}

	return defaultValue
}

//  MapStr returns the string value associated with given key in the map.
func MapStr(m map[string]interface{}, key, defaultValue string) string {
	if len(m) == 0 {
		return defaultValue
	}

	v, ok := m[key]
	if ok && v != "" {
		return fmt.Sprintf("%v", v)
	}

	return defaultValue
}
