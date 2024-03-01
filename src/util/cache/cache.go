package cache

import (
	"context"
	"strconv"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/manishchauhan/dugguGo/models/roomModel"
)

// RedisClient encapsulates a Redis client along with related functions.
type RedisClient struct {
	Client *redis.Client
}

var (
	redisClients map[string]*RedisClient
	mutex        sync.Mutex
)

// NewRedisClient initializes and returns a new Redis client instance.
func NewRedisClient(addr, password string, db int) *RedisClient {
	rdbInstance := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &RedisClient{Client: rdbInstance}
}

// AddMessageToRedisStream adds a message to a Redis stream using the specified client instance.
func (r *RedisClient) AddMessageToRedisStream(ctx context.Context, messageObject *roomModel.IFWebsocketMessage) error {
	_, err := r.Client.XAdd(ctx, &redis.XAddArgs{
		Stream: strconv.Itoa(messageObject.RoomId),
		Values: map[string]interface{}{
			"data": messageObject.Data,
			"time": messageObject.Time,
			"user": messageObject.User,
		},
	}).Result()
	return err
}

// InitializeRedisClient initializes and stores a new Redis client instance.
func InitializeRedisClient(instanceName, addr, password string, db int) {
	mutex.Lock()
	defer mutex.Unlock()

	if redisClients == nil {
		redisClients = make(map[string]*RedisClient)
	}
	redisClients[instanceName] = NewRedisClient(addr, password, db)
}

// GetRedisClient returns the Redis client instance with the specified name.
func GetRedisClient(instanceName string) (*RedisClient, bool) {
	mutex.Lock()
	defer mutex.Unlock()

	Client, ok := redisClients[instanceName]
	return Client, ok
}

// GetDataFromRedisStream retrieves data from a Redis stream based on the specified stream ID.
func (r *RedisClient) GetDataFromRedisStream(ctx context.Context, streamID string) ([]redis.XMessage, error) {
	return r.Client.XRange(ctx, streamID, "-", "+").Result()
}
