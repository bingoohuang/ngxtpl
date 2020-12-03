package ngxtpl

import (
	"bytes"
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/template"
	"time"

	"github.com/pkg/errors"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	"github.com/markbates/pkger"
)

// ReadPkger reads the content of pkger file.
func ReadPkger(file string) []byte {
	f, err := pkger.Open(file)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	d, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	return d
}

// ReadFile reads the file content of file with name filename.
// or panic if error happens.
func ReadFile(filename string) []byte {
	d, err := ReadFileE(filename)
	if err != nil {
		panic(err)
	}

	return d
}

// ReadFileE reads the file content of file with name filename.
func ReadFileE(filename string) ([]byte, error) {
	d, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return d, nil
}

// ReadFileStrE reads the file content of file with name filename.
func ReadFileStrE(filename string) (string, error) {
	d, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(d), nil
}

// SetupSingals setup the signal.
func SetupSingals(sig ...os.Signal) context.Context {
	return SetupSingalsWithContext(context.Background(), sig...)
}

// SetupSingalsWithContext setup the signal.
func SetupSingalsWithContext(parent context.Context, sig ...os.Signal) context.Context {
	ch := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(parent)

	if len(sig) == 0 {
		sig = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}

	signal.Notify(ch, sig...)

	go func() {
		<-ch
		cancel()
	}()

	return ctx
}

// PflagParse parses the plag and check unknown args.
func PflagParse() {
	pflag.Parse()

	if pflag.NArg() > 0 {
		logrus.Errorf("Unknown args %s\n", strings.Join(pflag.Args(), " "))
		pflag.PrintDefaults()
		os.Exit(-1)
	}
}

// CreateDemoCfg  creates a demo config file with name demoHclFile.
func CreateDemoCfg(demoHclFile string) {
	if v, err := os.Stat(demoHclFile); err == nil && !v.IsDir() {
		fmt.Printf("%s exists already\n", demoHclFile)
		os.Exit(0)
	}

	v := ReadPkger(pkger.Include("/assets/cfg.hcl"))
	if err := ioutil.WriteFile(demoHclFile, v, 0644); err != nil {
		panic(err)
	}

	fmt.Printf("%s created\n", demoHclFile)
}

// QueryRows querys the database and returns a slice of map.
func QueryRows(db *sql.DB, query string) ([]map[string]interface{}, error) {
	rows, _ := db.Query(query) // Note: Ignoring errors for brevity
	defer rows.Close()

	cols, _ := rows.Columns()

	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		columns := make([]string, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		m := make(map[string]interface{})
		for i, colName := range cols {
			m[colName] = columns[i]
		}

		results = append(results, m)
	}

	return results, nil
}

// IsHTTPAddress tests whether the string s starts with http:// or https://.
func IsHTTPAddress(s string) bool {
	return HasPrefix(s, "http://", "https://")
}

// HTTPGetStr execute HTTP GET request.
func HTTPGetStr(addr string) (string, error) {
	v, err := HTTPGet(addr)

	return string(v), err
}

// HTTPPost execute HTTP POST request.
func HTTPPost(addr string, body []byte) ([]byte, error) {
	return HTTPInvoke(http.MethodPost, addr, body)
}

// HTTPGet execute HTTP GET request.
func HTTPGet(addr string) ([]byte, error) {
	return HTTPInvoke(http.MethodGet, addr, nil)
}

// HTTPInvoke execute HTTP method request.
func HTTPInvoke(method, addr string, body []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives: true,
		},
	}

	var r io.Reader

	if len(body) > 0 {
		r = bytes.NewReader(body)
	}

	req, _ := http.NewRequestWithContext(ctx, method, addr, r)
	req.Close = true

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return ioutil.ReadAll(resp.Body)
	}

	return nil, errors.Wrapf(err, "status:%d", resp.StatusCode)
}

// HasPrefix tests whether the string s begins with any of prefix.
func HasPrefix(s string, prefix ...string) bool {
	for _, p := range prefix {
		if strings.HasPrefix(s, p) {
			return true
		}
	}

	return false
}

// HasBrace tests whether the string s begins with head and ends with tail.
func HasBrace(s string, head, tail string) bool {
	return strings.HasPrefix(s, head) && strings.HasSuffix(s, tail)
}

// JSONDecode decodes JSON string v to map[string]interface.
func JSONDecode(v string) (interface{}, error) {
	var data map[string]interface{}

	if err := json.Unmarshal([]byte(v), &data); err != nil {
		return nil, err
	}

	return data, nil
}

// TemplateEval execute a template s with data.
func TemplateEval(s string, data interface{}) (string, error) {
	t, err := template.New("").Parse(s)
	if err != nil {
		return "", err
	}

	var b bytes.Buffer

	if err := t.Execute(&b, data); err != nil {
		return "", err
	}

	return b.String(), nil
}

// Split2 splits the s with sep and return the trimmed string pair.
func Split2(s, sep string) (string, string) {
	p := strings.LastIndex(s, sep)
	return strings.TrimSpace(s[:p]), strings.TrimSpace(s[p+len(sep):])
}
