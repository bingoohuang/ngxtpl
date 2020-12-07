package ngxtpl

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

// https://github.com/nacos-group/nacos-sdk-go

// Nacos represents the structure of Nacos config.
type Nacos struct {
	ClientConfig  ClientConfig  `hcl:"clientConfig"`
	ServerConfigs ServerConfigs `hcl:"serverConfigs"`
	ServiceParam  ServiceParam  `hcl:"serviceParam"`

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
	// Create config client for dynamic configuration
	p := map[string]interface{}{
		"serverConfigs": n.ServerConfigs.toNacosConfig(),
		"clientConfig":  n.ClientConfig.toNacosConfig(),
	}

	var err error

	if n.configClient, err = clients.CreateConfigClient(p); err != nil {
		return nil, err
	}

	// Create naming client for service discovery
	if n.namingClient, err = clients.CreateNamingClient(p); err != nil {
		return nil, err
	}

	return n, err
}

// ServiceParam map config for service discovery.
type ServiceParam struct {
	Clusters    []string `hcl:"Clusters"`    // optional,default:DEFAULT
	ServiceName string   `hcl:"ServiceName"` // required
	GroupName   string   `hcl:"GroupName"`   // optional,default:DEFAULT_GROUP
}

// ServerConfig defines the config to nacos server.
type ServerConfig struct {
	Scheme      string `hcl:"Scheme"`      // the nacos server scheme
	ContextPath string `hcl:"ContextPath"` // the nacos server contextpath
	IPAddr      string `hcl:"IpAddr"`      // the nacos server address
	Port        int    `hcl:"Port"`        // the nacos server port
}

// ServerConfigs is alias of service config slice.
type ServerConfigs []ServerConfig

// ClientConfig defines the structure of config of nacos client.
type ClientConfig struct {
	TimeoutMs            int    `hcl:"TimeoutMs"`            // timeout for requesting Nacos server, default value is 10000ms
	BeatInterval         int    `hcl:"BeatInterval"`         // the time interval for sending beat to server,default value is 5000ms
	NamespaceID          string `hcl:"NamespaceId"`          // the namespaceId of Nacos
	Endpoint             string `hcl:"Endpoint"`             // the endpoint for get Nacos server addresses
	RegionID             string `hcl:"RegionId"`             // the regionId for kms
	AccessKey            string `hcl:"AccessKey"`            // the AccessKey for kms
	SecretKey            string `hcl:"SecretKey"`            // the SecretKey for kms
	OpenKMS              bool   `hcl:"OpenKMS"`              // it's to open kms,default is false. https://help.aliyun.com/product/28933.html
	CacheDir             string `hcl:"CacheDir"`             // the directory for persist nacos service info,default value is current path
	UpdateThreadNum      int    `hcl:"UpdateThreadNum"`      // the number of goroutine for update nacos service info,default value is 20
	NotLoadCacheAtStart  bool   `hcl:"NotLoadCacheAtStart"`  // not to load persistent nacos service info in CacheDir at start time
	UpdateCacheWhenEmpty bool   `hcl:"UpdateCacheWhenEmpty"` // update cache when get empty service instance from server
	Username             string `hcl:"Username"`             // the username for nacos auth
	Password             string `hcl:"Password"`             // the password for nacos auth
	LogDir               string `hcl:"LogDir"`               // the directory for log, default is current path
	RotateTime           string `hcl:"RotateTime"`           // the rotate time for log, eg: 30m, 1h, 24h, default is 24h
	MaxAge               int    `hcl:"MaxAge"`               // the max age of a log file, default value is 3
	LogLevel             string `hcl:"LogLevel"`             // the level of log, it's must be debug,info,warn,error, default value is info
}

func (s ServiceParam) toNacosConfig() vo.GetServiceParam {
	return vo.GetServiceParam{
		ServiceName: s.ServiceName,
		Clusters:    s.Clusters,
		GroupName:   s.GroupName,
	}
}

func (s ServerConfig) toNacosConfig() constant.ServerConfig {
	return constant.ServerConfig{
		Scheme:      s.Scheme,
		ContextPath: s.ContextPath,
		IpAddr:      s.IPAddr,
		Port:        uint64(s.Port),
	}
}

func (s ServerConfigs) toNacosConfig() []constant.ServerConfig {
	c := make([]constant.ServerConfig, len(s))

	for i, cnf := range s {
		c[i] = cnf.toNacosConfig()
	}

	return c
}

func (s ClientConfig) toNacosConfig() constant.ClientConfig {
	return constant.ClientConfig{
		TimeoutMs:            uint64(s.TimeoutMs),
		BeatInterval:         int64(s.BeatInterval),
		NamespaceId:          s.NamespaceID,
		Endpoint:             s.Endpoint,
		RegionId:             s.RegionID,
		AccessKey:            s.AccessKey,
		SecretKey:            s.SecretKey,
		OpenKMS:              s.OpenKMS,
		CacheDir:             s.CacheDir,
		UpdateThreadNum:      s.UpdateThreadNum,
		NotLoadCacheAtStart:  s.NotLoadCacheAtStart,
		UpdateCacheWhenEmpty: s.UpdateCacheWhenEmpty,
		Username:             s.Username,
		Password:             s.Password,
		LogDir:               s.LogDir,
		RotateTime:           s.RotateTime,
		MaxAge:               int64(s.MaxAge),
		LogLevel:             s.LogLevel,
	}
}
