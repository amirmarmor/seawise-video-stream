package persistance

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type Persist struct {
	redis 	*redis.Client
}

func Create() (*Persist, error) {
	addr := RedisConfig.Host + ":6379"
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	persist := &Persist{
		redis: rdb,
	}

	err := persist.ping()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis - %s", err)
	}
	return persist, nil
}

func(p *Persist) ping() error {
	pong, err := p.redis.Ping(ctx).Result()
	if err != nil {
		return err
	}
	fmt.Println(pong, err)
	return nil
}

func(p *Persist) Set(key string, value string) error {
	err := p.redis.Set(ctx, key, value, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to set %s %s: %s", key, value, err)
	}
	return nil
}

func(p *Persist) Get(key string) (string, error) {
	result, err := p.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get %s: %s", key, err)
	}
	return result, nil
}

func(p *Persist) HGetAll(key string) (interface{}, error){
	result, err := p.redis.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	fmt.Println(fmt.Sprintf("result %v", result))
	return result, nil
}