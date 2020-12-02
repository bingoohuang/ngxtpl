package ngxtpl

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
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
	Source      string `hcl:"source"`
	Destination string `hcl:"destination"`
	Perms       int    `hcl:"perms"`
	Command     string `hcl:"command"`

	interval time.Duration
	source   *template.Template
	tiker    *time.Ticker
}

// Execute executes the template.
func (t *Tpl) Execute(data interface{}) error {
	var out bytes.Buffer

	if err := t.source.Execute(&out, data); err != nil {
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
func (t *Tpl) Parse() error {
	if err := t.parseInterval(); err != nil {
		return err
	}

	if err := t.parseSource(); err != nil {
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

func (t *Tpl) parseSource() error {
	if t.Source == "" {
		return errors.Wrapf(ErrCfg, "source is empty")
	}

	var v []byte

	if stat, err := os.Stat(t.Source); err == nil && !stat.IsDir() {
		v, err = ReadFileE(t.Source)
		if err != nil {
			return err
		}
	} else {
		v = []byte(t.Source)
	}

	var err error

	if t.source, err = template.New("Source").Parse(string(v)); err != nil {
		return errors.Wrapf(ErrCfg, "source is invalid. it should be a template file or direct template string")
	}

	return nil
}

func (t *Tpl) parseDestination() error {
	if t.Destination == "" {
		return nil
	}

	dir := filepath.Dir(t.Destination)
	if v, err := os.Stat(dir); err != nil {
		return errors.Wrapf(ErrCfg, "Destination is invalid. it should be stdout of valid file. error: %v", err)
	} else if !v.IsDir() {
		return errors.Wrapf(ErrCfg, "Destination's dir %s does not exist", dir)
	}

	if t.Perms == 0 {
		t.Perms = 0644
	}

	return nil
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
