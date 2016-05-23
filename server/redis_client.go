package server

import (
	"fmt"

	redis "gopkg.in/redis.v3"
)

// NewRedisClient is a factory method that returns a new redis connection.
func NewRedisClient(hostname string, port int, password string,
	database int) (*redis.Client, error) {
	r := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", hostname, port),
		Password: password,
		DB:       int64(database),
	})
	if _, err := r.Ping().Result(); err != nil {
		return nil, err
	}
	return r, nil
}
