package ngxtpl

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"text/template"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

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
	d, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return d, nil
}

// ReadFileStrE reads the file content of file with name filename.
func ReadFileStrE(filename string) (string, error) {
	d, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(d), nil
}

// SetupSingals setups the signal.
func SetupSingals(sig ...os.Signal) context.Context {
	return SetupSingalsWithContext(context.Background(), sig...)
}

// SetupSingalsWithContext setups the signal.
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
func PflagParse(f *pflag.FlagSet, args []string) {
	f.Parse(args)

	if f.NArg() > 0 {
		log.Printf("E! Unknown args %s\n", strings.Join(f.Args(), " "))
		f.PrintDefaults()
		os.Exit(-1)
	}
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

	if r != nil {
		req.Header.Set("Content-Type", If(IsJSONBytes(body),
			"application/json; charset=utf-8", "text/plain; charset=utf-8"))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return io.ReadAll(resp.Body)
	}

	return nil, errors.Wrapf(err, "status:%d", resp.StatusCode)
}

// IsJSONBytes tests bytes b is in JSON format.
func IsJSONBytes(b []byte) bool {
	if len(b) == 0 {
		return false
	}

	var m interface{}
	return json.Unmarshal(b, &m) == nil
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

// FormatFloat format float with specified precision.
func FormatFloat(num float64, precision int) string {
	if precision == 0 {
		return fmt.Sprintf("%d", int(num))
	}

	zero, dot := "0", "."
	str := fmt.Sprintf("%."+strconv.Itoa(precision)+"f", num)

	return strings.TrimRight(strings.TrimRight(str, zero), dot)
}

// ZeroTo return a == 0 ? b : a.
func ZeroTo(a, b int) int {
	if a == 0 {
		return b
	}

	return a
}

// If tests condition to return a or b.
func If(condition bool, a, b string) string {
	if condition {
		return a
	}

	return b
}
