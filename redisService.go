package main

import (
	"time"

	"github.com/go-redis/redis"
)

var rdb *redis.Client

func initRedis() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}

func checkUserHasPosted(userID string, serverID string) (bool, error) {
	exists := rdb.Exists(userID + ":" + serverID)
	if exists.Err() != nil {
		return false, exists.Err()
	}

	if exists.Val() == 1 {
		return true, nil
	}

	err := rdb.Set(userID+":"+serverID, "1", time.Minute*1).Err()
	if err != nil {
		return false, err
	}
	return false, nil

}
