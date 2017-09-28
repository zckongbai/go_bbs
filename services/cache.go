package services

import (
	"github.com/fpay/gopress"
	"github.com/garyburd/redigo/redis"
	. "go_bbs/conf"
)

const (
	// CacheServiceName is the identity of cache service
	CacheServiceName = "cache"
)

// CacheService type
type CacheService struct {
	// Uncomment this line if this service has dependence on other services in the container
	// c *gopress.Container
	Redis redis.Conn
}

// NewCacheService returns instance of cache service
func NewCacheService() *CacheService {
	var err error
	cache := new(CacheService)

	//cache.Redis, err = redis.Dial("tcp", "127.0.0.1:6379")
	cache.Redis, err = redis.Dial("tcp", Conf.Redis.Server)
	if err != nil {
		panic(err)
	}
	return cache
}

// ServiceName is used to implements gopress.Service
func (s *CacheService) ServiceName() string {
	return CacheServiceName
}

// RegisterContainer is used to implements gopress.Service
func (s *CacheService) RegisterContainer(c *gopress.Container) {
	// Uncomment this line if this service has dependence on other services in the container
	// s.c = c
}

func (s *CacheService) SampleMethod() {
}
