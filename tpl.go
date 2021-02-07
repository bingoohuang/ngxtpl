package ngxtpl

import (
	"bytes"
	"encoding/json"
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
	// 测试命令
	TestCommand string `hcl:"testCommand"`
	// 测试命令执行结果检查，例如有OK字眼，不配置，则只检测测试命令的执行状态
	TestCommandCheck string `hcl:"testCommandCheck"`

	Command string `hcl:"command"`
	// 结果检查，例如有OK字眼，不配置，则只检测测试命令的执行状态ß
	CommandCheck string `hcl:"commandCheck"`

	interval time.Duration
	ticker   *time.Ticker
}

// Execute executes the template.
func (t *Tpl) Execute(data interface{}, ds DataSource, cfgName string) error {
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
		logrus.Infof("nothing changed for config file: %s", cfgName)
		return nil
	}

	logrus.Infof("new content %s", string(newContent))
	logrus.Infof("old content %s", string(oldContent))

	if err := t.writeDestination(newContent); err != nil {
		logrus.Errorf("failed to write destination %s err: %v", t.Destination, err)
		return err
	}

	if err := t.executeCommand(); err != nil {
		_ = t.writeDestination(oldContent) // rollback destination
		return err
	}

	return nil
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
	v, err := time.ParseDuration(DefaultTo(t.Interval, "0"))
	if err != nil {
		return err
	}

	if t.interval = v; t.interval > 0 {
		t.ticker = time.NewTicker(t.interval)
	}

	return nil
}

// DefaultTo test a, return b if a is empty.
func DefaultTo(a, b string) string {
	if a == "" {
		return b
	}

	return a
}

func (t *Tpl) resetTicker() {
	if t.ticker != nil {
		t.ticker.Reset(t.interval)
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

func (t *Tpl) parseDestination() (err error) {
	if t.Destination == "" {
		return nil
	}
	t.Perms = ZeroTo(t.Perms, 0644)

	dir := filepath.Dir(t.Destination)
	if _, err = os.Stat(dir); err == nil {
		return nil
	}

	if IsHTTPAddress(t.Destination) {
		return nil
	}

	return errors.Wrapf(ErrCfg, "Destination dir %s is invalid. "+
		"it should be valid file or http addr. error: %v", dir, err)
}

func (t *Tpl) readDestination() ([]byte, error) {
	if t.Destination == "stdout" {
		return nil, nil
	}

	f, err := ReadFileE(t.Destination)
	if err != nil && os.IsNotExist(err) {
		return nil, nil
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

	if t.TestCommand != "" {
		if ret, ok := executeCommand(t.TestCommand, t.TestCommandCheck); !ok {
			retJSON, _ := json.Marshal(ret)
			return fmt.Errorf("executeTestCommand %s failed: %s", t.TestCommand, string(retJSON))
		}
	}

	if ret, ok := executeCommand(t.Command, t.CommandCheck); !ok {
		retJSON, _ := json.Marshal(ret)
		return fmt.Errorf("executeCommand %s failed: %s", t.Command, string(retJSON))
	}

	return nil
}

// CommandResult is the result of command execution.
type CommandResult struct {
	ExitCode  int    `json:"exitCode"`
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr"`
	ExecError error  `json:"execError"`
}

// Sh executes a bash scripts.
func Sh(bash string) (*cmd.Cmd, cmd.Status) {
	p := cmd.NewCmd("sh", "-c", bash)
	return p, <-p.Start()
}

func executeCommand(command, commandCheck string) (*CommandResult, bool) {
	_, status := Sh(command)
	if status.Exit == 0 {
		logrus.Infof("exec command %s successfully", command)
	} else {
		logrus.Infof("exec command %s failed with exit code %d", command, status.Exit)
	}

	if len(status.Stdout) > 0 {
		logrus.Infof("%s", strings.Join(status.Stdout, "\n"))
	}

	if len(status.Stderr) > 0 {
		logrus.Errorf("%s", strings.Join(status.Stderr, "\n"))
	}

	if commandCheck == "" && status.Exit == 0 ||
		commandCheck != "" && SliceContains(append(status.Stdout, status.Stderr...), commandCheck) {
		// successfully
		return nil, true
	}

	return &CommandResult{
		ExitCode:  status.Exit,
		Stdout:    strings.Join(status.Stdout, "\n"),
		Stderr:    strings.Join(status.Stderr, "\n"),
		ExecError: status.Error,
	}, false
}

// SliceContains test if any element in slice contains sub.
func SliceContains(ss []string, sub string) bool {
	for _, s := range ss {
		if strings.Contains(s, sub) {
			return true
		}
	}

	return false
}

//  MapInt returns the int value associated with given key in the map.
func MapInt(m map[string]interface{}, key string, defaultValue int) int {
	if len(m) == 0 {
		return defaultValue
	}

	f, err := strconv.ParseFloat(fmt.Sprintf("%v", m[key]), 64)
	if err != nil {
		return defaultValue
	}

	return int(f)
}

//  MapStr returns the string value associated with given key in the map.
func MapStr(m map[string]interface{}, key, defaultValue string) string {
	if len(m) == 0 {
		return defaultValue
	}

	v, ok := m[key]
	if ok {
		return fmt.Sprintf("%v", v)
	}

	return defaultValue
}
