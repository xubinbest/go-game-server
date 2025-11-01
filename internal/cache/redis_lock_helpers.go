package cache

import (
	crand "crypto/rand"
	"encoding/hex"
	"time"

	"github.com/redis/go-redis/v9"
)

// unlockLua：仅当持有者token匹配时才删除锁（原子释放）
var unlockLua = redis.NewScript(`
    if redis.call("GET", KEYS[1]) == ARGV[1] then
        return redis.call("DEL", KEYS[1])
    else
        return 0
    end
`)

// renewLua：仅当持有者token匹配时续期TTL
var renewLua = redis.NewScript(`
    if redis.call("GET", KEYS[1]) == ARGV[1] then
        return redis.call("PEXPIRE", KEYS[1], ARGV[2])
    else
        return 0
    end
`)

func (r *RedisCache) setTokenForKey(key, token string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lockTokens[key] = token
}

func (r *RedisCache) popTokenForKey(key string) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	tok, ok := r.lockTokens[key]
	if ok {
		delete(r.lockTokens, key)
	}
	return tok, ok
}

func generateToken() (string, error) {
	b := make([]byte, 16)
	if _, err := crand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func randomJitter(maxMs int) time.Duration {
	if maxMs <= 0 {
		return 0
	}
	var b [1]byte
	if _, err := crand.Read(b[:]); err != nil {
		return 0
	}
	n := int(b[0]) % (maxMs + 1)
	return time.Duration(n) * time.Millisecond
}
