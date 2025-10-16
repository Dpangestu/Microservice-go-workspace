package mfa

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/redis/go-redis/v9"
)

type OTPService struct {
	rdb *redis.Client
}

func NewOTPService(rdb *redis.Client) *OTPService {
	return &OTPService{rdb: rdb}
}

func (s *OTPService) Generate(ctx context.Context, key string, ttl time.Duration) (string, error) {
	code := fmt.Sprintf("%06d", randInt(100000, 999999))
	if err := s.rdb.Set(ctx, "otp:"+key, code, ttl).Err(); err != nil {
		return "", err
	}
	return code, nil
}

func (s *OTPService) Verify(ctx context.Context, key, code string) bool {
	stored, err := s.rdb.Get(ctx, "otp:"+key).Result()
	if err != nil {
		return false
	}
	return stored == code
}

func randInt(min, max int) int {
	nBig, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	return int(nBig.Int64()) + min
}
