package cache

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type Cacheable interface {
	redis.Cmdable
}

func NewRedis(c *redis.Options) (rdb Cacheable, cleanup func(), err error) {
	rdb = redis.NewClient(c)

	cleanup = func() {
		if err0 := rdb.(*redis.Client).Close(); err0 != nil {
			log.Errorf("redis close failed. %+v", err0)
		} else {
			log.Infof("redis close success")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.DialTimeout)
	defer cancel()

	if err = rdb.Ping(ctx).Err(); err != nil {
		err = errors.Wrapf(err, "redis ping failed")
		return
	}
	return
}

func NewRedisCluster(c *redis.ClusterOptions) (rdb Cacheable, cleanup func(), err error) {
	rdb = redis.NewClusterClient(c)

	cleanup = func() {
		if err0 := rdb.(*redis.ClusterClient).Close(); err0 != nil {
			log.Errorf("redis cluster close failed. %+v", err0)
		} else {
			log.Infof("redis cluster close success")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.DialTimeout)
	defer cancel()

	if err = rdb.Ping(ctx).Err(); err != nil {
		err = errors.Wrapf(err, "redis cluster ping failed")
		return
	}
	return
}
