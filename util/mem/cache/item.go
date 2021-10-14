/*
 *   base cache for block manager.
 *   author: kennyzhu
 */

package cache

import (
	`sync`
	`time`
)

// Parameter data contains the user-set value in the cache.
// support cache.
type CacheItem struct {
	sync.RWMutex
	
	// The item's key, usually id.
	key string
	
	// The item's data for file.
	data []byte
	
	// How long will the item live in the cache when not being accessed/kept alive.
	lifeSpan time.Duration
	
	// Creation timestamp.
	createdOn time.Time
	
	// Last access timestamp.
	accessedOn time.Time
	
	// How big the item was stored.
	accessSize int
	
	// Callback method triggered right before removing the item from the cache
	aboutToExpire []func(key string)
}

// NewCacheItem returns a newly created CacheItem.
// Parameter lifeSpan determines after which time period without an access the item
// will get removed from the cache.
func NewCacheItem(key string, data []byte, lifeSpan time.Duration) *CacheItem {
	t := time.Now()
	return &CacheItem{
		key:           key,
		lifeSpan:      lifeSpan,
		createdOn:     t,
		accessedOn:    t,
		accessSize:    len(data),
		aboutToExpire: nil,
		data:          data,
	}
}

// KeepAlive marks an item to be kept for another expireDuration period.
func (item *CacheItem) KeepAlive() {
	item.Lock()
	defer item.Unlock()
	item.accessedOn = time.Now()
	// item.accessCount++
}

// LifeSpan returns this item's expiration duration.
func (item *CacheItem) LifeSpan() time.Duration {
	// immutable
	return item.lifeSpan
}

// AccessedOn returns when this item was last accessed.
func (item *CacheItem) AccessedOn() time.Time {
	item.RLock()
	defer item.RUnlock()
	return item.accessedOn
}

// CreatedOn returns when this item was added to the cache.
func (item *CacheItem) CreatedOn() time.Time {
	// immutable
	return item.createdOn
}

// AccessCount returns how often this item has been accessed.
func (item *CacheItem) AccessSize() int {
	item.RLock()
	defer item.RUnlock()
	return item.accessSize
}

// Key returns the key of this cached item.
func (item *CacheItem) Key() interface{} {
	// immutable
	return item.key
}

// Data returns the value of this cached item.
func (item *CacheItem) Data() interface{} {
	// immutable
	return item.data
}

// SetAboutToExpireCallback configures a callback, which will be called right
// before the item is about to be removed from the cache.
func (item *CacheItem) SetAboutToExpireCallback(f func(string)) {
	if len(item.aboutToExpire) > 0 {
		item.RemoveAboutToExpireCallback()
	}
	item.Lock()
	defer item.Unlock()
	item.aboutToExpire = append(item.aboutToExpire, f)
}

// AddAboutToExpireCallback appends a new callback to the AboutToExpire queue
func (item *CacheItem) AddAboutToExpireCallback(f func(string)) {
	item.Lock()
	defer item.Unlock()
	item.aboutToExpire = append(item.aboutToExpire, f)
}

// RemoveAboutToExpireCallback empties the about to expire callback queue
func (item *CacheItem) RemoveAboutToExpireCallback() {
	item.Lock()
	defer item.Unlock()
	item.aboutToExpire = nil
}
