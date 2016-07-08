package hutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/tinylib/msgp/msgp"
)

var (
	rcPool redis.Pool
)

// CacheDial simply creates connection to redis
func CacheDial(network, address, password, dbnum string) (redis.Conn, error) {

	c, err := redis.Dial(network, address)

	if err != nil {
		return nil, err
	}

	if password != "" {
		if _, err := c.Do("AUTH", password); err != nil {
			c.Close()
			return nil, err
		}
	}

	if _, err := c.Do("SELECT", dbnum); err != nil {
		c.Close()
		return nil, err
	}

	return c, err
}

// CacheAlive returns nil if anything is ok or error
func CacheAlive() error {
	c := rcPool.Get()
	defer c.Close()
	_, err := c.Do("PING")
	if nil != err {
		return err
	}
	return nil
}

// CacheInit creates redis connections pool and tests paramaters
func CacheInit(network, address, password, dbnum string) error {
	rcPool = redis.Pool{
		MaxIdle:     64,                // max number of idle connections... TODO: move to config
		IdleTimeout: 300 * time.Second, // 5-minutes idle
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
		Dial: func() (redis.Conn, error) {
			return CacheDial(network, address, password, dbnum)
		},
	}
	return CacheAlive()
}

// CacheSet sets redis key/value for timeout expire
func CacheSet(key string, value interface{}, timeout int64) error {
	c := rcPool.Get()
	defer c.Close()
	res, err := c.Do("SET", key, value)
	if nil != err {
		return err
	}
	if res != "OK" {
		return errors.New(res.(string))
	}
	res, err = c.Do("EXPIRE", key, timeout)
	if nil != err {
		return err
	}

	if rI, ok := res.(int64); ok {
		if 1 != rI {
			return errors.New("Unable to set EXPIRE ri / " + fmt.Sprintf("%#v %T", rI, rI))
		}
	} else if rS, ok := res.(string); ok && "1" != rS {
		return errors.New("Unable to set EXPIRE O / " + fmt.Sprintf("%#v %T", res, res))
	}

	return nil
}

// CacheDelete removes key from redis
func CacheDelete(key string) error {
	c := rcPool.Get()
	defer c.Close()
	res, err := c.Do("DEL", key)
	if nil != err {
		return err
	}
	rI, ok := res.(int64)
	if ok && 1 != rI {
		return errors.New("Key not found in redis database")
	} else if "1" != res {
		return errors.New("Key not found in redis database")
	}
	return nil
}

// CacheProlong set expire of key
func CacheProlong(key string, timeout int64) error {
	c := rcPool.Get()
	defer c.Close()
	res, err := c.Do("EXPIRE", key, timeout)
	if nil != err {
		return err
	}

	rI, ok := res.(int64)
	if ok && 1 != rI {
		return errors.New("Unable to prolong cache")
	} else if "1" != res {
		return errors.New("Unable to prolong cache: " + res.(string))
	}

	return nil
}

// CacheGet returns value for key in redis
func CacheGet(key string) ([]byte, error) {
	c := rcPool.Get()
	defer c.Close()
	res, err := c.Do("GET", key)
	if nil != err {
		return nil, err
	}
	if nil == res {
		return nil, nil
	}
	return res.([]byte), nil
}

// CacheSetEncoded sets key/value using gob enoder
func CacheSetEncoded(key string, value interface{}, timeout int64) error {
	out, err := value.(msgp.Marshaler).MarshalMsg(nil)
	if nil != err {
		return err
	}
	return CacheSet(key, out, timeout)
}

// CacheGetEncoded returns decoded gob value for key
func CacheGetEncoded(key string, value interface{}) error {
	temp, err := CacheGet(key)

	if nil != err {
		return err
	}

	if nil == temp {
		return errors.New("No cache value")
	}

	value.(msgp.Unmarshaler).UnmarshalMsg(temp)

	if nil != err {
		return err
	}
	return nil
}

// CachePublish publish value to redis channel
func CachePublish(channel string, value interface{}) error {
	content, err := json.Marshal(value)
	if nil != err {
		return err
	}

	c := rcPool.Get()
	defer c.Close()
	res, err := c.Do("PUBLISH", channel, content)

	if nil != err {
		return err
	}

	if res != "OK" {
		if str, ok := res.(string); ok {
			return errors.New(str)
		}
	}

	return nil
}

// CacheSubscribe subscribes to redis channel
func CacheSubscribe(channel string) redis.PubSubConn {
	c := rcPool.Get()
	s := redis.PubSubConn{Conn: c}
	s.Subscribe(channel)

	return s
}
