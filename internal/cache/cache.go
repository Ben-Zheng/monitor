package cache

import (
	"context"
	"fmt"
	"github.com/patrickmn/go-cache"
	"sync"
	"time"
)

const (
	DefaultExpiration      = 5 * time.Minute  // 默认缓存过期时间
	DefaultCleanupInterval = 10 * time.Minute // 默认清理间隔
)

// Cache 接口定义
type Cache interface {
	Get(ctx context.Context, key string) (interface{}, bool)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration)
	SetDefault(ctx context.Context, key string, value interface{})
	Delete(ctx context.Context, key string)
	Flush(ctx context.Context)
	Items() map[string]cache.Item
	ItemCount() int
	GetOrSet(ctx context.Context, key string, loader LoaderFunc, ttl time.Duration) (interface{}, error)
	GetItemsWithValues() map[string]interface{}
}

// LoaderFunc 缓存加载函数
type LoaderFunc func(ctx context.Context) (interface{}, error)

// goCache 包装 go-cache 的实现
type goCache struct {
	cache     *cache.Cache
	loaderMap *sync.Map
}

// New 创建缓存实例
func New() Cache {
	c := cache.New(DefaultExpiration, DefaultCleanupInterval)
	return &goCache{
		cache:     c,
		loaderMap: new(sync.Map),
	}
}

// NewWithConfig 创建带自定义配置的缓存
func NewWithConfig(defaultExpiration, cleanupInterval time.Duration) Cache {
	c := cache.New(defaultExpiration, cleanupInterval)
	return &goCache{
		cache:     c,
		loaderMap: new(sync.Map),
	}
}

func (c *goCache) Get(ctx context.Context, key string) (interface{}, bool) {
	return c.cache.Get(key)
}

func (c *goCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) {
	c.cache.Set(key, value, ttl)
}

func (c *goCache) SetDefault(ctx context.Context, key string, value interface{}) {
	c.cache.SetDefault(key, value)
}

func (c *goCache) Delete(ctx context.Context, key string) {
	c.cache.Delete(key)
}

func (c *goCache) Flush(ctx context.Context) {
	c.cache.Flush()
}

// 修正：正确返回 go-cache 的 Items 结构
func (c *goCache) Items() map[string]cache.Item {
	return c.cache.Items()
}

func (c *goCache) ItemCount() int {
	return c.cache.ItemCount()
}

// 新增方法：获取所有缓存项的键值对（不包含元数据）
func (c *goCache) GetItemsWithValues() map[string]interface{} {
	items := c.cache.Items()
	result := make(map[string]interface{}, len(items))
	for k, item := range items {
		result[k] = item.Object
	}
	return result
}

func (c *goCache) GetOrSet(ctx context.Context, key string, loader LoaderFunc, ttl time.Duration) (interface{}, error) {
	// 尝试直接获取缓存
	if val, found := c.Get(ctx, key); found {
		return val, nil
	}

	// 创建或获取加载锁
	var mu *sync.Mutex
	val, _ := c.loaderMap.LoadOrStore(key, &sync.Mutex{})
	mu = val.(*sync.Mutex)
	mu.Lock()
	defer func() {
		mu.Unlock()
		c.loaderMap.Delete(key)
	}()

	// 再次检查缓存（防止并发期间已被其他协程填充）
	if val, found := c.Get(ctx, key); found {
		return val, nil
	}

	// 执行加载函数
	value, err := loader(ctx)
	if err != nil {
		return nil, fmt.Errorf("loader error: %w", err)
	}

	// 设置缓存
	c.Set(ctx, key, value, ttl)
	return value, nil
}

// 获取项目并更新过期时间
func (c *goCache) GetAndRefresh(ctx context.Context, key string, ttl time.Duration) (interface{}, bool) {
	value, found := c.Get(ctx, key)
	if found {
		c.Set(ctx, key, value, ttl)
	}
	return value, found
}

// 获取项目过期时间
func (c *goCache) GetExpiration(ctx context.Context, key string) (time.Time, bool) {
	items := c.Items()
	if item, exists := items[key]; exists {
		// go-cache 的 Expiration 字段是 int64 纳秒时间戳
		return time.Unix(0, item.Expiration), true
	}
	return time.Time{}, false
}

// 手动删除过期项目
func (c *goCache) DeleteExpired(ctx context.Context) {
	now := time.Now().UnixNano()
	items := c.Items()
	for key, item := range items {
		if now > item.Expiration {
			c.Delete(ctx, key)
		}
	}
}

// 添加项目（仅当不存在时）
func (c *goCache) Add(ctx context.Context, key string, value interface{}, ttl time.Duration) bool {
	if _, exists := c.Get(ctx, key); exists {
		return false
	}
	c.Set(ctx, key, value, ttl)
	return true
}
