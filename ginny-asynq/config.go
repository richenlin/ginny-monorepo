package asyncq

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/hibiken/asynq"
)

// Config
type Config struct {
	Dsn                    string `json:"dsn" yaml:"dsn"`
	redisClientOpt         *asynq.RedisClientOpt
	redisFailoverClientOpt *asynq.RedisFailoverClientOpt
	redisClusterClientOpt  *asynq.RedisClusterClientOpt
	asynq.Config
}

// NewConfig
func NewConfig(ctx context.Context, dsn string) (*Config, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	db := q.Get("db")
	cluster := q.Get("cluster")
	userName := q.Get("username")
	password := q.Get("password")
	masterName := q.Get("masterName")
	poolSizeNum := 10
	poolSize := q.Get("poolSize")
	if poolSize != "" {
		if i, err := strconv.Atoi(poolSize); err == nil {
			poolSizeNum = i
		}
	}
	c := &Config{Dsn: dsn}

	if masterName != "" {
		sentinelPassword := q.Get("sentinelPassword")
		var sentinelAddrs []string
		sentinelAddr := q.Get("sentinelAddrs")
		if strings.Contains(sentinelAddr, ",") {
			sentinelAddrs = strings.Split(sentinelAddr, ",")
		} else {
			sentinelAddrs = []string{sentinelAddr}
		}
		redisFailoverClientOpt := &asynq.RedisFailoverClientOpt{
			MasterName:       masterName,
			SentinelAddrs:    sentinelAddrs,
			Username:         userName,
			Password:         password,
			SentinelPassword: sentinelPassword,
			PoolSize:         poolSizeNum,
		}
		if i, err := strconv.Atoi(db); err == nil {
			redisFailoverClientOpt.DB = i
		}
		c.redisFailoverClientOpt = redisFailoverClientOpt
	} else if cluster != "" {
		var clusterAddrs []string
		if strings.Contains(cluster, ",") {
			clusterAddrs = strings.Split(cluster, ",")
		} else {
			clusterAddrs = []string{u.Host}
		}
		redisClusterClientOpt := &asynq.RedisClusterClientOpt{
			MaxRedirects: 3,
			Username:     userName,
			Password:     password,
			Addrs:        clusterAddrs,
		}
		c.redisClusterClientOpt = redisClusterClientOpt
	} else {
		redisClientOpt := &asynq.RedisClientOpt{
			Addr:     u.Host,
			Username: userName,
			Password: password,
			PoolSize: poolSizeNum,
		}
		if i, err := strconv.Atoi(db); err == nil {
			redisClientOpt.DB = i
		}
		c.redisClientOpt = redisClientOpt
	}

	return c, nil
}
