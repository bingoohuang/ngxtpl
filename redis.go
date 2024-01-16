package ngxtpl

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/bingoohuang/gg/pkg/iox"
	"github.com/bingoohuang/gg/pkg/ss"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

// Redis represents the structure of redis config.
type Redis struct {
	Addr        string `hcl:"addr"`
	Password    string `hcl:"password"`
	ServicesKey string `hcl:"servicesKey"`
	ResultKey   string `hcl:"resultKey"`
}

// Get gets the value of key from redis.
func (r Redis) Get(key string) (string, error) {
	addrs := ss.Split(r.Addr)
	rdb := redis.NewClusterClient(&redis.ClusterOptions{Addrs: addrs, Password: r.Password})
	defer iox.Close(rdb)

	ctx := context.Background()

	const sep = " "

	if strings.Contains(key, sep) {
		// treat as a hash
		hashKey, field := Split2(key, sep)
		return rdb.HGet(ctx, hashKey, field).Result()
	}

	return rdb.Get(ctx, key).Result()
}

// Write writes key and it's value to redis.
func (r Redis) Write(key, value string) (err error) {
	addrs := ss.Split(r.Addr)
	rdb := redis.NewClusterClient(&redis.ClusterOptions{Addrs: addrs, Password: r.Password})
	defer iox.Close(rdb)

	ctx := context.Background()

	const sep = " "

	if strings.Contains(key, sep) {
		// treat as a hash
		hashKey, field := Split2(key, sep)
		_, err = rdb.HSet(ctx, hashKey, field, value).Result()
	}

	_, err = rdb.Set(ctx, key, value, 0).Result()
	return err
}

// WriteResult writes error.
func (r Redis) WriteResult(result Result) error {
	if r.ResultKey == "" {
		return nil
	}

	result.Time = time.Now().Format("2006-01-02 15:04:05.000 ")
	bytes, _ := json.Marshal(result)
	return r.Write(r.ResultKey, string(bytes))
}

// Read reads the value.
func (r Redis) Read() (interface{}, error) {
	v, err := r.Get(r.ServicesKey)
	if err != nil {
		return nil, err
	}

	return JSONDecode(v)
}

// Parse parses the redis config.
func (r *Redis) Parse() (DataSource, error) {
	if r.ServicesKey == "" {
		return nil, errors.Wrapf(ErrCfg, "ServicesKey is required")
	}

	return r, nil
}
