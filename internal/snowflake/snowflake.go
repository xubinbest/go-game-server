package snowflake

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.xubinbest.com/go-game-server/internal/cache"

	"github.com/redis/go-redis/v9"
)

const (
	epoch             = 1609459200000 // 2021-01-01 00:00:00 UTC
	workerIdBits      = 10            // 合并原来的5位datacenterId和5位workerId
	sequenceBits      = 12
	workerIdKey       = "snowflake:worker_ids"
	workerIdLockKey   = "snowflake:worker_lock"
	workerIdTTL       = 24 * time.Hour
	heartbeatInterval = 30 * time.Minute

	maxWorkerId    = -1 ^ (-1 << workerIdBits)
	maxSequence    = -1 ^ (-1 << sequenceBits)
	workerIdShift  = sequenceBits
	timestampShift = sequenceBits + workerIdBits
)

type Snowflake struct {
	mu            sync.Mutex
	lastTimestamp int64
	workerId      int64
	sequence      int64
	redisClient   cache.Cache
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewSnowflakeWithRedis(redisClient cache.Cache) (*Snowflake, error) {
	ctx, cancel := context.WithCancel(context.Background())
	sf := &Snowflake{
		redisClient: redisClient,
		ctx:         ctx,
		cancel:      cancel,
	}

	// Acquire workerId from Redis
	workerId, err := sf.acquireWorkerId()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire worker ID: %v", err)
	}
	sf.workerId = workerId

	// Start heartbeat to keep workerId alive
	go sf.heartbeat()

	return sf, nil
}

func (sf *Snowflake) acquireWorkerId() (int64, error) {
	// Try to get an available workerId
	workerId, err := sf.redisClient.SPop(sf.ctx, workerIdKey).Int64()
	if err == redis.Nil {
		// No available IDs, create new one
		workerId, err = sf.createNewWorkerId()
		if err != nil {
			return 0, err
		}
	} else if err != nil {
		return 0, err
	}

	// Lock the workerId
	ok, err := sf.redisClient.SetNX(sf.ctx,
		fmt.Sprintf("%s:%d", workerIdLockKey, workerId),
		os.Getpid(),
		workerIdTTL).Result()
	if err != nil || !ok {
		sf.redisClient.SAdd(sf.ctx, workerIdKey, workerId)
		return 0, fmt.Errorf("failed to lock worker ID: %d", workerId)
	}

	return workerId, nil
}

func (sf *Snowflake) createNewWorkerId() (int64, error) {
	// Get current max workerId
	maxId, err := sf.redisClient.Get(sf.ctx, "snowflake:max_worker_id").Int64()
	if err == redis.Nil {
		maxId = 0
	} else if err != nil {
		return 0, err
	}

	if maxId >= maxWorkerId {
		return 0, errors.New("no available worker IDs")
	}

	newId := maxId + 1
	err = sf.redisClient.Set(sf.ctx, "snowflake:max_worker_id", newId, 0)
	if err != nil {
		return 0, err
	}

	return newId, nil
}

func (sf *Snowflake) heartbeat() {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			lockKey := sf.getLockKey()
			sf.redisClient.Expire(sf.ctx, lockKey, workerIdTTL)
		case <-sf.ctx.Done():
			return
		}
	}
}

func (sf *Snowflake) Close() error {
	// Stop heartbeat
	sf.cancel()

	// Release workerId back to pool
	_, err := sf.redisClient.SAdd(sf.ctx, workerIdKey, sf.workerId).Result()
	if err != nil {
		return err
	}

	// Remove the lock
	lockKey := sf.getLockKey()
	err = sf.redisClient.Delete(sf.ctx, lockKey)
	return err
}

func (sf *Snowflake) NextID() (int64, error) {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	now := time.Now().UnixMilli() - epoch
	if now < 0 {
		return 0, errors.New("clock moved backwards")
	}

	if now == sf.lastTimestamp {
		sf.sequence = (sf.sequence + 1) & maxSequence
		if sf.sequence == 0 {
			for now <= sf.lastTimestamp {
				now = time.Now().UnixMilli() - epoch
			}
		}
	} else {
		sf.sequence = 0
	}

	sf.lastTimestamp = now

	id := (now << timestampShift) |
		(sf.workerId << workerIdShift) |
		sf.sequence

	return id, nil
}

func (sf *Snowflake) getLockKey() string {
	return fmt.Sprintf("%s:%d", workerIdLockKey, sf.workerId)
}
