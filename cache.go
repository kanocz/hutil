package hutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"sync/atomic"

	"github.com/garyburd/redigo/redis"
	"github.com/tinylib/msgp/msgp"
)

var (
	rcPool redis.Pool
)

// internal function to detect redis master
func getMaster(servers []string, password string, dbnum int, last string, data []interface{}) (string, redis.Conn) {
	tested := make(map[string]bool, len(servers))

	if data == nil && last != "" {
		c, _ := CacheDial(last, password, dbnum)
		var m bool
		if nil != c {
			m, data, _ = IsRedisMaster(c)
			if m {
				return last, c
			}
			c.Do("QUIT")
			c.Close()
		}

		tested[last] = true
	}

	if nil != data && len(data) > 2 {
		// if we just have data in some case...
		mode, _ := redis.String(data[0], nil)
		if mode == "slave" {
			s, _ := redis.String(data[1], nil)
			p, _ := redis.Int(data[2], nil)
			if s != "" && p > 0 {
				// try to connect to detected master
				test := fmt.Sprintf("%s:%d", s, p)
				c, _ := CacheDial(test, password, dbnum)
				if nil != c {
					if m, _, _ := IsRedisMaster(c); m {
						return test, c
					}
					c.Do("QUIT")
					c.Close()
				}
				tested[test] = true
			}
		}
	}

	for _, server := range servers {
		if _, ok := tested[server]; ok {
			continue
		}

		c, _ := CacheDial(server, password, dbnum)
		if nil != c {
			if m, _, _ := IsRedisMaster(c); m {
				return server, c
			}
			c.Do("QUIT")
			c.Close()
		}
		tested[server] = true
	}

	return "", nil
}

// IsRedisMaster returns true, nil, []interface{} if connected to redis master
func IsRedisMaster(c redis.Conn) (bool, []interface{}, error) {
	if nil == c {
		return false, nil, nil
	}

	res, err := c.Do("ROLE")
	if nil != err {
		return false, nil, err
	}
	values, err := redis.Values(res, nil)
	if nil != err {
		return false, nil, err
	}

	if len(values) < 3 {
		return false, values, errors.New("invalid ROLE responce")
	}

	role, err := redis.String(values[0], nil)
	if nil != err {
		return false, values, err
	}

	if role != "master" {
		return false, values, errors.New("not a master")
	}

	return true, values, nil
}

// CacheDial simply creates connection to redis
func CacheDial(address, password string, dbnum int) (redis.Conn, error) {
	// todo: move timeouts to config
	c, err := redis.Dial("tcp", address,
		redis.DialConnectTimeout(time.Second),
		redis.DialWriteTimeout(time.Second),
		redis.DialDatabase(dbnum),
		redis.DialPassword(password),
	)

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
func CacheInit(servers []string, password string, dbnum int) error {
	master, _ := getMaster(servers, password, dbnum, servers[0], nil)
	if master == "" {
		return errors.New("no redis master found")
	}

	var lastServer atomic.Value
	lastServer.Store(master)

	rcPool = redis.Pool{
		MaxIdle:     64,               // max number of idle connections... TODO: move to config
		IdleTimeout: 30 * time.Second, // 30 seconds
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			// test not more than every 5 seconds
			if time.Since(t) < (time.Second * 5) {
				return nil
			}

			m, _, err := IsRedisMaster(c)
			if !m {
				return errors.New("not a master")
			}

			return err
		},
		Dial: func() (redis.Conn, error) {
			last := lastServer.Load().(string)
			c, _ := CacheDial(last, password, dbnum)
			m, data, _ := IsRedisMaster(c)
			if m {
				return c, nil
			}

			// real close connection, not just return to pool
			if nil != c {
				c.Do("QUIT")
				c.Close()
			}

			master, c := getMaster(servers, password, dbnum, last, data)
			if master == "" {
				return nil, errors.New("unable to find redis master")
			}
			lastServer.Store(master)
			return c, nil
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
	if timeout > 0 {
		res, err = c.Do("EXPIRE", key, timeout)
		if nil != err {
			return err
		}

		if rI, ok := res.(int64); ok {
			if rI != 1 {
				return errors.New("Unable to set EXPIRE ri / " + fmt.Sprintf("%#v %T", rI, rI))
			}
		} else if rS, ok := res.(string); ok && rS != "1" {
			return errors.New("Unable to set EXPIRE O / " + fmt.Sprintf("%#v %T", res, res))
		}
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
	if ok && rI != 1 {
		return errors.New("key not found in redis database")
	} else if res != "1" {
		return errors.New("key not found in redis database")
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
	if ok && rI != 1 {
		return errors.New("unable to prolong cache")
	} else if res != "1" {
		return errors.New("unable to prolong cache: " + res.(string))
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
		return errors.New("no cache value")
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
