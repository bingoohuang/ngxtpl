package ngxtpl

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

// Redis represents the structure of redis config.
type Redis struct {
	Addr        string `hcl:"addr"`
	Password    string `hcl:"password"`
	Db          int    `hcl:"db"`
	ServicesKey string `hcl:"servicesKey"`
}

func (r Redis) Read() (interface{}, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     r.Addr,
		Password: r.Password,
		DB:       r.Db,
	})
	defer rdb.Close()

	ctx := context.Background()

	var (
		v    string
		data map[string]interface{}
		err  error
	)

	const sep = " "

	if strings.Contains(r.ServicesKey, sep) {
		// treat as a hash
		p := strings.LastIndex(r.ServicesKey, sep)
		hashKey := strings.TrimSpace(r.ServicesKey[:p])
		field := strings.TrimSpace(r.ServicesKey[p+1:])

		v, err = rdb.HGet(ctx, hashKey, field).Result()
	} else {
		v, err = rdb.Get(ctx, r.ServicesKey).Result()
	}

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(v), &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Parse parse the redis config.
func (r *Redis) Parse() (DataSource, error) {
	if r.ServicesKey == "" {
		return nil, errors.Wrapf(ErrCfg, "ServicesKey is required")
	}

	return r, nil
}
