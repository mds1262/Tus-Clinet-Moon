package lib

import (
	"github.com/eventials/go-tus"
	"github.com/go-redis/redis"
	"log"
)

type TusStore struct {
	*redis.Client
}

//type TusRedisStore interface {
//	Get(key string) (string, bool)
//	Set(key, val string)
//	Delete(key string)
//	Close()
//}

func NewTusStore(redisClient *redis.Client) tus.Store {
	return &TusStore{redisClient}
}

func (s *TusStore) Get(key string) (string, bool) {
	result, err := s.HMGet(key, REDISTUSHASHKEY).Result()

	if err != nil || result[0] == nil {
		log.Print(err)
		return "", false
	}

	return result[0].(string), true
}

func (s *TusStore) Set(key, val string) {
	var err error

	err = s.HMSet(key, map[string]interface{}{
			REDISTUSHASHKEY: val,
			REDISTUSDOWNHASHKEY:val,
			REDISTUSREMOVEHASHKEY: val,
	}).Err()
	if err != nil {
		log.Println(err)
	}

	//err = s.HMSet(key, map[string]interface{}{REDISTUSDOWNHASHKEY: val}).Err()
	//if err != nil {
	//	log.Println(err)
	//}
}

func (s *TusStore) Delete(key string) {
	err := s.HDel(key, REDISTUSHASHKEY).Err()

	if err != nil {
		log.Println(err)
	}
	//delete(s.m, fingerprint)
}

func (s *TusStore) Close() {
	s.Close()
}
