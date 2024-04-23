package redis

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()
var Rdb *redis.Client

// Redis 연결
func ConnectRedis() error {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // user default DB
	})

	log.Println("Connected Redis")
	return nil
}

// Redis에 token 저장
func SetToken(key, value string, exp int64) error {
	// Time To Live
	ttl := time.Until(time.Unix(exp, 0))
	if err := Rdb.Set(Ctx, key, value, ttl).Err(); err != nil {
		return err
	} else {
		return nil
	}
}

// Redis에서 refresh-token 조회
func GetRefreshToken(account string) string { // refreshToken, error 반환

	// account로 Redis 조회해서 redis-token 가져오기
	token, err := Rdb.Get(Ctx, account).Result()

	if err == redis.Nil {
		return ""
	}

	return token
}

// Redis에서 access-token 조회 > BlackList인지 확인
func IsBlackList(accessToken string) bool { // AccesshToken, error 반환
	// accessToken Redis 조회해서 blackList인지 아닌지 확인
	result, _ := Rdb.Get(Ctx, accessToken).Result()
	if result != "BlackList" {
		return false
	} else {
		return true
	}
}

// RefreshToken 삭제
func DeleteRefreshToken(account string) error {
	if err := Rdb.Del(Ctx, account).Err(); err != nil {
		return err
	} else {
		return nil
	}
}
