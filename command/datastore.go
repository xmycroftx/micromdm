package command

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/go-kit/kit/log"
	"github.com/groob/plist"
	"github.com/micromdm/mdm"
)

var (
	// ErrNoKey is returned if there is no key in redis
	ErrNoKey = errors.New("There is no such key in redis.")
)

// Datastore manages MDM Payloads in redis
type Datastore interface {
	// Saves the payload in redis
	// SET CommandUUID plistData
	SavePayload(payload *mdm.Payload) error
	// Adds MDM commands to a queue in redis list
	// LPUSH deviceUDID commandUUID
	QueueCommand(deviceUDID, commandUUID string) error
	NextCommand(deviceUDID string) ([]byte, int, error)
	DeleteCommand(deviceUDID, commandUUID string) (int, error)
}

type redisDB struct {
	pool *redis.Pool
}

// NewDB creates a new databases connection
func NewDB(driver, conn string, options ...func(*config) error) Datastore {
	conf := &config{}
	defaultLogger := log.NewLogfmtLogger(os.Stderr)
	for _, option := range options {
		if err := option(conf); err != nil {
			defaultLogger.Log("err", err)
			os.Exit(1)
		}
	}
	switch driver {
	case "redis":
		return redisDB{pool: redisPool(conn, conf.logger)}
	default:
		conf.logger.Log("err", "unknown driver")
		os.Exit(1)
		return nil
	}
}

func redisPool(conn string, logger log.Logger) *redis.Pool {
	pool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", conn)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	checkRedisConn(pool, logger)
	return pool
}

func checkRedisConn(pool *redis.Pool, logger log.Logger) {
	conn := pool.Get()
	defer conn.Close()

	var dbError error
	maxAttempts := 20
	for attempts := 1; attempts <= maxAttempts; attempts++ {
		_, dbError = conn.Do("PING")
		if dbError == nil {
			break
		}
		logger.Log("msg", fmt.Sprintf("could not connect to redis: %v", dbError))
		time.Sleep(time.Duration(attempts) * time.Second)
	}
	if dbError != nil {
		logger.Log("err", dbError)
		os.Exit(1)
	}
}

func (rds redisDB) SavePayload(payload *mdm.Payload) error {
	var buf bytes.Buffer
	// get connection from redis pool
	conn := rds.pool.Get()
	defer conn.Close()
	// encode payload into a plist
	err := plist.NewEncoder(&buf).Encode(payload)
	if err != nil {
		return err
	}
	// create a commandUUID key with the plist as the value
	_, err = conn.Do("set", payload.CommandUUID, buf.String())
	if err != nil {
		return err
	}
	return nil
}

func (rds redisDB) QueueCommand(deviceUDID, commandUUID string) error {
	// get connection from redis pool
	conn := rds.pool.Get()
	defer conn.Close()
	_, err := conn.Do("lpush", deviceUDID, commandUUID)
	if err != nil {
		return err
	}
	return nil
}
func (rds redisDB) NextCommand(deviceUDID string) ([]byte, int, error) {
	// get connection from redis pool
	conn := rds.pool.Get()
	defer conn.Close()
	// pop the first command
	commandUUID, err := redis.String(conn.Do("lpop", deviceUDID))
	if err != nil && err != redis.ErrNil {
		return nil, 0, err
	}
	// if the list is empty
	if err == redis.ErrNil {
		return []byte{}, 0, nil
	}
	// push the redis command back to the end of the list
	_, err = conn.Do("rpush", deviceUDID, commandUUID)
	command, err := redis.String(conn.Do("get", commandUUID))
	if err == redis.ErrNil {
		return nil, 0, ErrNoKey
	}

	// get a command list length
	total, err := redis.Int(conn.Do("llen", deviceUDID))
	if err != nil {
		return nil, 0, err
	}
	return []byte(command), total, err
}

func (rds redisDB) DeleteCommand(deviceUDID, commandUUID string) (int, error) {
	// get connection from redis pool
	conn := rds.pool.Get()
	defer conn.Close()
	// remove from list
	_, err := conn.Do("lrem", deviceUDID, 0, commandUUID)
	if err != nil {
		return 0, err
	}
	// set the key to expire in an hour
	_, err = conn.Do("expire", commandUUID, 3600)
	if err != nil {
		return 0, err
	}
	// get a command list length
	total, err := redis.Int(conn.Do("llen", deviceUDID))
	if err != nil {
		return 0, err
	}
	return total, nil
}
