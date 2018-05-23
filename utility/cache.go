package utility

import (
	"container/list"
	"sync"
	"time"
)

const (
	GCInterval = time.Second * 1
)

type CacheElement interface {
	Key() string
	Invalid() bool
	Erase()
}

type Cache struct {
	mutex    sync.Mutex
	elements map[string]*list.Element
	list     *list.List
}

func NewCache() *Cache {
	cache := &Cache{
		elements: make(map[string]*list.Element),
		list:     list.New(),
	}

	go cache.GC()

	return cache
}

func (cache *Cache) Add(key string, element CacheElement) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	elmt := cache.list.PushBack(element)
	cache.elements[key] = elmt
}

func (cache *Cache) Lookup(key string) (CacheElement, bool) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	elmt, ok := cache.elements[key]
	if !ok {
		return nil, false
	}

	element := elmt.Value.(CacheElement)

	return element, true
}

func (cache *Cache) GC() {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	var next *list.Element
	iter := cache.list.Front()

	for iter != nil {
		next = iter.Next()

		element := iter.Value.(CacheElement)

		if element.Invalid() {
			delete(cache.elements, element.Key())

			cache.list.Remove(iter)

			element.Erase()
		}

		iter = next
	}

	time.AfterFunc(GCInterval, cache.GC)
}

type ReferenceCounter struct {
	keyString        string
	lifeTime         time.Duration
	referenceCount   int32
	invalidTimestamp int64
	mutex            sync.Mutex
}

func NewReferenceCounter(key string, lifeTime time.Duration) *ReferenceCounter {
	return &ReferenceCounter{
		keyString:        key,
		lifeTime:         lifeTime,
		referenceCount:   0,
		invalidTimestamp: time.Now().Unix(),
	}
}

func (rc *ReferenceCounter) Key() string {
	return rc.keyString
}

func (rc *ReferenceCounter) Invalid() bool {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	return (rc.referenceCount == 0) && (time.Now().Unix()-rc.invalidTimestamp > int64(rc.lifeTime/time.Second))
}

func (rc *ReferenceCounter) AddReference() {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	rc.referenceCount += 1
}

func (rc *ReferenceCounter) DelReference() {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	rc.referenceCount -= 1

	if rc.referenceCount == 0 {
		rc.invalidTimestamp = time.Now().Unix()
	}
}
