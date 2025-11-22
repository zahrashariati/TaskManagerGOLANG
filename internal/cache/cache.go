// stores and retrieves tasks from Redis cache
package cache

import (
	"context"                                //for redis operations
	"encoding/json"                          //for marshalling and unmarshalling
	"errors"                                 //for creating errors
	"strconv"                                //for converting int to string
	"task_manager/internal/models"           //task struct
	"time"                                   //for expiration time

	"github.com/redis/go-redis/v9" //redis client library
)

// Cache keys defined outside functions
var (
	tasksAllKey       = "tasks:all"
	tasksIncompleteKey = "tasks:incomplete"
)

type Cache struct {
	client *redis.Client //holds redis connection- pointer to redis.Client struct
}

func NewCache(client *redis.Client) *Cache { //pointer to Cache struct
	return &Cache{client: client}
}

func (c *Cache) GetTasks(showCompleted bool) ([]*models.Task, error) {
	// Local error definitions
	var (
		ErrCacheGetFailed      = errors.New("cache get failed")
		ErrCacheUnmarshalFailed = errors.New("cache unmarshal failed")
	)

	key := tasksAllKey
	if !showCompleted {
		key = tasksIncompleteKey
	}
	val, err := c.client.Get(context.Background(), key).Result() //redis operations need context for cancellation/timeouts - here no timeout(default context)
	if err != nil {
		return nil, ErrCacheGetFailed //return error if redis operation fails
	}
	var tasks []*models.Task
	err = json.Unmarshal([]byte(val), &tasks) //unmarshal json string to tasks slice because redis stores strings
	if err != nil {
		return nil, ErrCacheUnmarshalFailed //return error if unmarshalling fails
	}
	return tasks, nil
}

//marshalling: converting Go structs to JSON strings
//unmarshalling: converting JSON strings to Go structs

func (c *Cache) SetTasks(showCompleted bool, tasks []*models.Task) error {
	// Local error definitions
	var (
		ErrCacheMarshalFailed = errors.New("cache marshal failed")
		ErrCacheSetFailed     = errors.New("cache set failed")
	)

	key := tasksAllKey
	if !showCompleted {
		key = tasksIncompleteKey
	}
	data, err := json.Marshal(tasks)
	if err != nil {
		return ErrCacheMarshalFailed
	}
	err = c.client.Set(context.Background(), key, data, time.Minute*5).Err()
	if err != nil {
		return ErrCacheSetFailed
	}
	return nil
}

func (c *Cache) Invalidate() error {
	return c.client.Del(context.Background(), tasksAllKey, tasksIncompleteKey).Err()
}

func (c *Cache) GetTask(id int) (*models.Task, error) {
	// Local error definitions
	var (
		ErrCacheGetFailed      = errors.New("cache get failed")
		ErrCacheUnmarshalFailed = errors.New("cache unmarshal failed")
	)

	key := "task:" + strconv.Itoa(id)
	val, err := c.client.Get(context.Background(), key).Result()
	if err != nil {
		return nil, ErrCacheGetFailed
	}
	var task models.Task
	err = json.Unmarshal([]byte(val), &task)
	if err != nil {
		return nil, ErrCacheUnmarshalFailed
	}
	return &task, nil
}

func (c *Cache) SetTask(id int, task *models.Task) error {
	// Local error definitions
	var ErrCacheMarshalFailed = errors.New("cache marshal failed")

	key := "task:" + strconv.Itoa(id)
	data, err := json.Marshal(task)
	if err != nil {
		return ErrCacheMarshalFailed
	}
	return c.client.Set(context.Background(), key, data, time.Minute*5).Err()
}

func (c *Cache) InvalidateTask(id int) error {
	return c.client.Del(context.Background(), "task:"+strconv.Itoa(id)).Err()
}
