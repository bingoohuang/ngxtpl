package ngxtpl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bingoohuang/gg/pkg/goip"
	"github.com/goccy/go-yaml"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/common/logger"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"log"
	"os"
	"strings"
	"text/template"
	"time"
)

// https://github.com/nacos-group/nacos-sdk-go

// Nacos represents the structure of Nacos config.
type Nacos struct {
	ConfigFile   string       `hcl:"configFile"`
	ServiceParam ServiceParam `hcl:"serviceParam"`

	namingClient naming_client.INamingClient
	configClient config_client.IConfigClient
}

// Get gets the value of key from mysql.
func (n Nacos) Get(key string) (string, error) {
	dataID, group := Split2(key, " ")
	return n.configClient.GetConfig(vo.ConfigParam{
		DataId: dataID,
		Group:  group,
	})
}

func (n Nacos) Read() (interface{}, error) {
	c := n.ServiceParam.toNacosConfig()
	service, err := n.namingClient.GetService(c)
	if err != nil {
		return nil, err
	}

	upstreamConfig, err := n.upstreamConfig()
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"upstreams": []map[string]interface{}{n.createUpstream(upstreamConfig, service)},
	}

	return data, nil
}

func (n Nacos) upstreamConfig() (map[string]interface{}, error) {
	config, err := n.configClient.GetConfig(vo.ConfigParam{
		DataId: n.ServiceParam.ServiceName,
		Group:  "upstreams",
	})

	upstreamConfig := map[string]interface{}{}

	if err != nil && !is404Error(err) {
		return nil, err
	}

	if config != "" {
		if err := json.Unmarshal([]byte(config), &upstreamConfig); err != nil {
			return nil, err
		}
	}

	return upstreamConfig, nil
}

func is404Error(err error) bool {
	return err != nil && strings.Contains(err.Error(), "config not found")
}

func (n Nacos) createUpstream(config map[string]interface{}, service model.Service) map[string]interface{} {
	servers := make([]map[string]interface{}, 0, len(service.Hosts))

	for _, s := range service.Hosts {
		if s.Enable {
			servers = append(servers, createUpstreamHost(s))
		}
	}

	return map[string]interface{}{
		"name":      n.ServiceParam.ServiceName,
		"state":     MapStr(config, "state", "1"),
		"keepalive": MapInt(config, "keepalive", 0),
		"servers":   servers,
	}
}

func createUpstreamHost(s model.Instance) map[string]interface{} {
	return map[string]interface{}{
		"address": s.Ip,
		"port":    fmt.Sprintf("%d", s.Port),
		"state":   "1",
		"weight":  FormatFloat(s.Weight, 0),
	}
}

// Parse parses the nacos config.
func (n *Nacos) Parse() (DataSource, error) {
	if err := n.loadNaocsClient(n.ConfigFile); err != nil {
		return nil, err
	}

	return n, nil
}

// ServiceParam map config for service discovery.
type ServiceParam struct {
	ServiceName string   `hcl:"ServiceName"` // required
	GroupName   string   `hcl:"GroupName"`   // optional,default:DEFAULT_GROUP
	Clusters    []string `hcl:"Clusters"`    // optional,default:DEFAULT
}

func (s ServiceParam) toNacosConfig() vo.GetServiceParam {
	return vo.GetServiceParam{
		ServiceName: s.ServiceName,
		Clusters:    s.Clusters,
		GroupName:   s.GroupName,
	}
}

func (n *Nacos) loadNaocsClient(configYamlFileName string) error {
	var config Config
	if err := ParseConfig(configYamlFileName, &config); err != nil {
		return fmt.Errorf("parse config error: %w", err)
	}

	// Another way of create naming client for service discovery (recommend)
	clientParam := vo.NacosClientParam{
		ClientConfig:  &config.ClientConfig,
		ServerConfigs: config.ServerConfigs,
	}
	var err error

	// Another way of create config client for dynamic configuration (recommend)
	n.configClient, err = clients.NewConfigClient(clientParam)
	if err != nil {
		return fmt.Errorf("clients.NewConfigClient error: %w", err)
	}
	if config.ClientConfig.LogLevel == "" {
		logger.SetLogger(&noLogger{})
	}

	var registerInstanceParam *vo.RegisterInstanceParam

	p, err := GetRegisterParam(config, n.configClient)
	if err != nil {
		log.Printf("GetRegisterParam error: %v", err)
	}
	if p != "" {
		registerInstanceParam, err = createRegisterInstanceParam(p)
		if err != nil {
			log.Printf("createRegisterInstanceParam error: %v", err)
		}
	} else if config.RegisterInstanceParam != nil {
		registerInstanceParam = config.RegisterInstanceParam
	}

	n.namingClient, err = clients.NewNamingClient(clientParam)
	if err != nil {
		return fmt.Errorf("clients.NewNamingClient error: %w", err)
	}
	if config.ClientConfig.LogLevel == "" {
		logger.SetLogger(&noLogger{})
	}

	if registerInstanceParam != nil {
		success, err := n.namingClient.RegisterInstance(*registerInstanceParam)
		if err != nil {
			log.Fatalf("namingClient.RegisterInstance error: %v", err)
		}
		log.Printf("namingClient.RegisterInstance result: %t", success)
	}

	return nil
}

func createRegisterInstanceParam(registerParam string) (*vo.RegisterInstanceParam, error) {
	tmpl, err := template.New("").Parse(registerParam)
	if err != nil {
		log.Fatalf("template.New error: %v", err)
	}
	ip, ips := goip.MainIP()
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]any{
		"Ip":           ip,
		"Ips":          strings.Join(ips, ","),
		"Hostname":     Pick1(os.Hostname()),
		"Pid":          os.Getpid(),
		"RegisterTime": time.Now().Format(time.RFC3339),
	}); err != nil {
		return nil, fmt.Errorf("tmpl.Execute error: %w", err)
	}

	var param vo.RegisterInstanceParam
	if err := yaml.UnmarshalWithOptions(buf.Bytes(), &param, yaml.WithKeyMatchMode(yaml.KeyMatchStrict)); err != nil {
		return nil, fmt.Errorf("yaml.Unmarshal error: %w", err)
	}
	return &param, nil
}

func GetRegisterParam(config Config, c config_client.IConfigClient) (registerParam string, err error) {
	if config.RegisterParam == nil {
		return "", nil
	}

	_, err = c.GetConfig(vo.ConfigParam{Group: "xxx", DataId: "yyy"})
	// 不存在，会返回错误，告知: config data not exist
	if err != nil && strings.Contains(err.Error(), "config data not exist") {
		log.Printf("W! clients.NewConfigClient get xxx.yyy error: %v", err)
	}

	if registerParam, err = c.GetConfig(*config.RegisterParam); err != nil {
		return "", fmt.Errorf("clients.NewConfigClient error: %w", err)
	}

	return registerParam, nil
}

func Pick1[T any](a T, _ ...any) T {
	return a
}

func ParseConfig(configYamlFileName string, v any) error {
	config, err := os.ReadFile(configYamlFileName)
	if err != nil {
		return fmt.Errorf("os.ReadFile %s error: %w", configYamlFileName, err)
	}

	if err := yaml.UnmarshalWithOptions(config, v, yaml.WithKeyMatchMode(yaml.KeyMatchStrict)); err != nil {
		return fmt.Errorf("yaml.Unmarshal error: %w", err)
	}

	return nil
}

type Config struct {
	ServerConfigs         []constant.ServerConfig
	ClientConfig          constant.ClientConfig
	RegisterParam         *vo.ConfigParam
	RegisterInstanceParam *vo.RegisterInstanceParam
}

type noLogger struct{}

func (l noLogger) Info(args ...interface{})               {}
func (l noLogger) Warn(args ...interface{})               {}
func (l noLogger) Error(args ...interface{})              {}
func (l noLogger) Debug(args ...interface{})              {}
func (l noLogger) Infof(fmt string, args ...interface{})  {}
func (l noLogger) Warnf(fmt string, args ...interface{})  {}
func (l noLogger) Errorf(fmt string, args ...interface{}) {}
func (l noLogger) Debugf(fmt string, args ...interface{}) {}
