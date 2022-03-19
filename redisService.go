package main

import (
	"time"

	"github.com/go-redis/redis"
)

var rdb *redis.Client

func initRedis() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     REDIS_ADDR,
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

func modifyBlocklist(channelID string, listType string, state bool) (err error) {
	if state {
		status := rdb.Set(channelID+":"+listType, "1", 0)
		err = status.Err()
	} else {
		status := rdb.Del(channelID + ":" + listType)
		err = status.Err()
	}
	return
}

func checkBlocklist(channelID string, listType string) (bool, error) {
	exists := rdb.Exists(channelID + ":" + listType)
	if exists.Err() != nil {
		return false, exists.Err()
	}

	if exists.Val() == 1 {
		return true, nil
	}

	return false, nil
}
