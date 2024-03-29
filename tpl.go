package ngxtpl

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/bingoohuang/gg/pkg/codec"
	"github.com/bingoohuang/gg/pkg/iox"
	"github.com/gobars/cmd"
	"github.com/pkg/errors"
)

// InitAssets gives the initial assets for the program initialization.
//
//go:embed initassets
var InitAssets embed.FS

// Tpl represents a tpl config section.
type Tpl struct {
	ticker      *time.Ticker
	DataSource  string `hcl:"dataSource"`
	Interval    string `hcl:"interval"`
	TplSource   string `hcl:"tplSource"`
	Destination string `hcl:"destination"`
	// 执行命令（包括测试命令）失败时，写入导致失败的内容
	FailedDestination string `hcl:"failDestination"`
	// 测试命令
	TestCommand string `hcl:"testCommand"`
	// 测试命令执行结果检查，例如有OK字眼，不配置，则只检测测试命令的执行状态
	TestCommandCheck string `hcl:"testCommandCheck"`

	Command string `hcl:"command"`
	// 结果检查，例如有OK字眼，不配置，则只检测测试命令的执行状态ß
	CommandCheck string `hcl:"commandCheck"`

	Perms int `hcl:"perms"`

	interval time.Duration
}

// Execute executes the template.
func (t *Tpl) Execute(data interface{}, ds DataSource, cfgName string, result *Result) error {
	var out bytes.Buffer

	sourceContent, err := t.parseSource(ds)
	if err != nil {
		return err
	}

	source, err := template.New("TplSource").Parse(sourceContent)
	if err != nil {
		if t := iox.WriteTempFile(iox.WithTempContent([]byte(sourceContent))); t.Err == nil {
			log.Printf("parse template %s failed: %v", t.Name, err)
		}
		return errors.Wrapf(ErrCfg, "TplSource is invalid. "+
			"it should be a template file or direct template content string")
	}

	if err := source.Execute(&out, data); err != nil {
		t1 := iox.WriteTempFile(iox.WithTempContent([]byte(sourceContent)))
		t2 := iox.WriteTempFile(iox.WithTempContent(codec.Json(data)))
		if t1.Err == nil && t2.Err == nil {
			log.Printf("evaluting template %s with data %s failed: %v", t1.Name, t2.Name, err)
		}
		return err
	}

	newContent := out.Bytes()
	oldContent, err := t.readDestination()
	if err != nil {
		return err
	}

	result.Old = string(oldContent)
	if bytes.Equal(newContent, oldContent) {
		result.StatusCode = 304
		log.Printf("nothing changed for config file: %s", cfgName)
		return nil
	}

	result.Old = string(oldContent)
	result.New = string(newContent)
	result.StatusCode = 200

	log.Printf("{PRE}new content:\n%s", result.New)
	log.Printf("{PRE}old content:\n%s", result.Old)

	if err := t.writeDestination(t.Destination, newContent); err != nil {
		log.Printf("E! failed to write destination %s err: %v", t.Destination, err)
		return err
	}

	if err := t.executeCommand(); err != nil {
		_ = t.writeDestination(t.Destination, oldContent)       // rollback destination
		_ = t.writeDestination(t.FailedDestination, newContent) // try save failed context
		return err
	}

	return nil
}

// Parse parses and validates the template.
func (t *Tpl) Parse() error {
	if err := t.parseInterval(); err != nil {
		return err
	}

	if t.TplSource == "" {
		return errors.Wrapf(ErrCfg, "source is empty")
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
		k := source[len(dataSourcePrefix):]
		if kr, ok := ds.(KeyReader); ok {
			return kr.Get(k)
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
	t.Perms = ZeroTo(t.Perms, 0o644)

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

func (t *Tpl) writeDestination(destination string, content []byte) error {
	if destination == "" {
		return nil
	}

	if IsHTTPAddress(destination) {
		resp, err := HTTPPost(destination, content)
		if err != nil {
			return err
		}

		log.Printf("POST %s response %s", destination, string(resp))
		return nil
	}

	return os.WriteFile(destination, content, os.FileMode(t.Perms))
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
	ExecError error  `json:"execError"`
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr"`
	ExitCode  int    `json:"exitCode"`
}

// Sh executes a bash scripts.
func Sh(bash string) (*cmd.Cmd, cmd.Status) {
	p := cmd.NewCmd("sh", "-c", bash)
	return p, <-p.Start()
}

func executeCommand(command, check string) (*CommandResult, bool) {
	_, status := Sh(command)
	if status.Exit == 0 {
		log.Printf("exec command %s successfully", command)
	} else {
		log.Printf("exec command %s failed with exit code %d", command, status.Exit)
	}

	if len(status.Stdout) > 0 {
		log.Printf("%s", strings.Join(status.Stdout, "\n"))
	}

	if len(status.Stderr) > 0 {
		log.Printf("E! %s", strings.Join(status.Stderr, "\n"))
	}

	if check == "" && status.Exit == 0 ||
		check != "" && SliceContains(append(status.Stdout, status.Stderr...), check) {
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

// MapInt returns the int value associated with given key in the map.
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

// MapStr returns the string value associated with given key in the map.
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
