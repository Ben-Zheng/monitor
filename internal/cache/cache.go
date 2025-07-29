package cache

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// 全局缓存结构
type SharedCache struct {
	data          sync.Map      // 核心并发安全存储
	expirationMap sync.Map      // 过期时间管理
	cleanupTicker *time.Ticker  // 清理定时器
	shutdownChan  chan struct{} // 关闭信号
}

// 缓存值（可根据需要扩展）
type CacheValue struct {
	Value  interface{} `json:"value"`  // 实际存储的值
	Expire int64       `json:"expire"` // Unix时间戳，0表示永不过期
}

func NewSharedCache() *SharedCache {
	cache := &SharedCache{
		shutdownChan: make(chan struct{}),
	}

	// 启动定期清理任务
	cache.cleanupTicker = time.NewTicker(500 * time.Minute)
	go cache.startCleanupTask()

	// 注册优雅退出
	cache.setupGracefulShutdown()

	return cache
}

// Set 存储键值对
func (c *SharedCache) Set(key string, value interface{}, ttl time.Duration) {
	expire := int64(0)
	if ttl > 0 {
		expire = time.Now().Add(ttl).Unix()
	}

	cv := CacheValue{
		Value:  value,
		Expire: expire,
	}

	c.data.Store(key, cv)
	if expire > 0 {
		c.expirationMap.Store(key, expire)
	}
}

// SetWithoutExpire 存储永不过期的键值对
func (c *SharedCache) SetWithoutExpire(key string, value interface{}) {
	c.Set(key, value, 0)
}

// Get 获取键值
func (c *SharedCache) Get(key string) (interface{}, bool) {
	value, found := c.data.Load(key)
	if !found {
		return nil, false
	}

	cv, ok := value.(CacheValue)
	if !ok {
		return nil, false
	}

	// 检查是否过期
	if cv.Expire > 0 && cv.Expire < time.Now().Unix() {
		c.data.Delete(key)
		c.expirationMap.Delete(key)
		return nil, false
	}
	return cv.Value, true
}

// GetAndUpdate 获取并更新值
func (c *SharedCache) GetAndUpdate(key string, updater func(interface{}) interface{}) {
	value, found := c.data.Load(key)

	var cv CacheValue
	if found {
		cv = value.(CacheValue)
	} else {
		cv = CacheValue{}
	}

	newValue := updater(cv.Value)
	cv.Value = newValue

	c.data.Store(key, cv)
}

// Delete 删除键
func (c *SharedCache) Delete(key string) {
	c.data.Delete(key)
	c.expirationMap.Delete(key)
}

// Cleanup 定期清理过期项
func (c *SharedCache) Cleanup() {
	now := time.Now().Unix()
	count := 0
	c.expirationMap.Range(func(key, value interface{}) bool {
		expire := value.(int64)
		if expire > 0 && expire < now {
			c.data.Delete(key)
			c.expirationMap.Delete(key)
			count++
		}
		return true
	})
	if count > 0 {
		log.Printf("清理了 %d 个过期项", count)
	}
}

// startCleanupTask 启动清理定时任务
func (c *SharedCache) startCleanupTask() {
	for {
		select {
		case <-c.cleanupTicker.C:
			c.Cleanup()
		case <-c.shutdownChan:
			c.cleanupTicker.Stop()
			log.Println("清理任务已停止")
			return
		}
	}
}

// setupGracefulShutdown 设置优雅退出
func (c *SharedCache) setupGracefulShutdown() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sig
		log.Println("收到停止信号，准备关闭...")
		// 停止清理任务
		close(c.shutdownChan)

		log.Println("服务已停止")
		os.Exit(0)
	}()
}
