package cache

import (
	"common/log/log"
	logdebug "common/util/debug"
	"common/util/file"
	"os"
	"path/filepath"
	"strings"
	`sync`
	`time`
)

// FileItems is a table within the cache
// only support the same lift time for all items.
type FileItems struct {
	totalSizeMutex sync.Mutex
	// expirationLock sync.Mutex
	
	// The table's name, which is a directory.
	name string

	// Timer responsible for triggering cleanup.
	cleanupTimer *time.Timer

	// Current timer duration.
	cleanupInterval time.Duration
	
	// Callback method triggered when trying to load a non-existing key.
	loadData func(key string, args ...interface{}) *FileItem
	// Callback method triggered when adding a new item to the cache.
	addedItem func(item *FileItem)
	// Callback method triggered before deleting an item from the cache.
	// aboutToDeleteItem func(item *FileItem)

	totalSize int64
	limitSize int64 // bytes..
	lifeSpan time.Duration // as all default life span for all items.

	// All cached items. not consider gc
	// consider gc must shard.
	// items map[string]*CacheItem
	items *sync.Map // [string]*FileItem

	overSize func(key, datapath string, lifeSpan time.Duration, data []byte) (bDelete bool)

	// Callback method triggered right before removing the item from the cache
	aboutToExpire func(*FileItem) (bDelete bool)
	aboutToDelete func(*FileItem) (bDelete bool)
}

// init from base, name is directory.
// 注意:这一步耗时比较严重.
func LoadFileItems(name string, limits int64, span time.Duration, bLoad bool ) *FileItems {
	t := &FileItems{
		name:          name,
		lifeSpan:      span,
		limitSize:     limits,
		totalSize:     0,
		cleanupTimer:  nil,
		cleanupInterval: 0*time.Second,
		items: new(sync.Map),
	}

	defer func() {
		go t.expirationCheck()
	}()

	log.Infof("LoadFileItems enter, name:%v, span:%v ", name, span)
	source := file.TreeInfos{
		Files: make([]*file.SysFile, 0),
	}
	err := filepath.Walk(name, func(path string, f os.FileInfo, err error) error {
		return source.Visit(path, f, err)
	})

	if err != nil {
		log.Errorf("clearEmptyFiles filePath.Walk() returned:%v", err)
		return t
	}

	// 遍历下面所有的文件，不合法的过期的都删除..
	if !bLoad {
		return t
	}

	for _, v := range source.Files {
		// 判断文件.
		if b,_ := file.PathExists(v.FName); !b {
			continue
		}

		if v.FType == file.IsDirectory {
			continue
		}

		if strings.Contains(v.FName, ".idx") {
			filename := v.FName[:len(v.FName)-4]
			if b,_ := file.PathExists(filename); !b {
				os.Remove( v.FName )
				continue
			} else {
				// to
				keys := strings.Split(filename,"/")
				key := keys[len(keys)-1]
				path := filename[:len(filename) - len(key)]
				_ = t.AddFromFile(key, path, t.lifeSpan)
			}
		} else {
			filename := v.FName + ".idx"
			if b,_ := file.PathExists(filename); !b {
				os.Remove( v.FName )
				continue
			} else {
				keys := strings.Split(v.FName,"/")
				key := keys[len(keys)-1]
				path := filename[:len(v.FName) - len(key)]
				_ = t.AddFromFile(key, path, t.lifeSpan)
			}
		}
	}

	// add expiration check.
	return t
}

func (table *FileItems) SetAboutToExpireCallback(f func(*FileItem) bool) {
	// table.Lock()
	// defer table.Unlock()
	table.aboutToExpire = f
}

func (table *FileItems) SetAboutToDeleteCallback(f func(*FileItem) bool) {
	// table.Lock()
	// defer table.Unlock()
	table.aboutToDelete = f
}

func (table *FileItems) TotalSize() int64 {
	return table.totalSize
}

func (table *FileItems) addTotalSize(l int64)  {
	//table.totalSizeMutex.Lock()
	//table.totalSize += l
	//table.totalSizeMutex.Unlock()
}

func (table *FileItems) subTotalSize(l int64)  {
	//table.totalSizeMutex.Lock()
	//table.totalSize -= l
	//table.totalSizeMutex.Unlock()
}

func (table *FileItems) SetLimitSize(l int64)  {
	table.totalSizeMutex.Lock()
	table.limitSize = l
	table.totalSizeMutex.Unlock()
}

// Foreach all items
func (table *FileItems) Foreach(ops func(key, value interface{}) bool) {
	table.items.Range(ops)
}

// SetDataLoader configures a data-loader callback, which will be called when
// trying to access a non-existing key. The key and 0...n additional arguments
// are passed to the callback function.
func (table *FileItems) SetDataLoader(f func(string, ...interface{}) *FileItem) {
	table.loadData = f
}

// SetAddedItemCallback configures a callback, which will be called every time
// a new item is added to the cache.
func (table *FileItems) SetAddedItemCallback(f func(*FileItem)) {
	// table.Lock()
	// defer table.Unlock()
	table.addedItem = f
}

// SetAboutToDeleteItemCallback configures a callback, which will be called
// every time an item is about to be removed from the cache.
//func (table *FileItems) SetAboutToDeleteItemCallback(f func(*FileItem)) {
//	table.Lock()
//	defer table.Unlock()
//	table.aboutToDeleteItem = f
//}

func (table *FileItems) SetOverSizeCallback(f func(key, datapath string, lifeSpan time.Duration, data []byte) bool ) {
	table.overSize = f
}

// Expiration check loop, triggered by a self-adjusting timer.
// 注意:如果没有任何item的情况..
// TODO: expirationCheck定时器精度太高了，改成秒超时定时器...
func (table *FileItems) expirationCheck() {
	// attention: new goroutine here must add defer function.
	defer func() {
		if v := recover(); v != nil {
			// debug.PrintStack()
			log.Errorf("expirationCheck table name[%v] panic:%v", table.name, v)
			logdebug.LogLocalStacks()
		}
	}()
	//table.expirationLock.Lock()
	//defer table.expirationLock.Unlock()

	if table.cleanupTimer != nil {
		table.cleanupTimer.Stop() // for another?..
	}
	if table.cleanupInterval > 0 {
		log.Infof("Expiration check triggered after %v for table %v", table.cleanupInterval, table.name)
	}
	//} else {
	//	log.Infof("Expiration check installed for table %v", table.name)
	//}

	// To be more accurate with timers, we would need to update 'now' on every
	// loop iteration. Not sure it's really efficient though.
	now := time.Now()
	smallestDuration := 0 * time.Second

	var itemStrings []string
	ops := func(key, value interface{}) bool {
		item := value.(*FileItem)
		// Cache values so we don't keep blocking the mutex.
		// item.RLock()
		lifeSpan := item.lifeSpan
		accessedOn := item.accessedOn

		if now.Sub(accessedOn) >= lifeSpan {
			itemStrings = append(itemStrings, key.(string))
		} else {
			// Find the item chronologically closest to its end-of-lifespan.
			if smallestDuration == 0 || lifeSpan-now.Sub(accessedOn) < smallestDuration {
				smallestDuration = lifeSpan - now.Sub(accessedOn)
			}
		}
		return true
	}
	table.items.Range(ops)

	// delete all outdated items
	for _, key := range itemStrings {
		table.deleteInternal(key)
	}
	
	// Setup the interval for the next cleanup run, set the smallest Duration .
	table.cleanupInterval = smallestDuration
	if smallestDuration > 0 {
		table.cleanupTimer = time.AfterFunc(smallestDuration, func() {
			// go table.expirationCheck()
			table.expirationCheck()
		})
	} else {
		// check for 5 seconds idle.
		table.cleanupTimer = time.AfterFunc(time.Second * 5, func() {
			// go table.expirationCheck()
			table.expirationCheck()
		})
	}
}

// add one item internally, not lock
func (table *FileItems) addInternal(item *FileItem) {
	log.Infof("Adding item with key %v and lifespan of %v to table %v", item.key, item.lifeSpan, table.name)
	item.SetAboutToExpireCallback(table.aboutToExpire)
	item.SetAboutToDeleteCallback(table.aboutToDelete)
	table.items.Store(item.key, item)
	// table.items[item.key] = item
	
	// Cache values so we don't keep blocking the mutex.
	// expDur := table.cleanupInterval
	addedItem := table.addedItem
	
	// Trigger callback after adding an item to cache.
	if addedItem != nil {
		addedItem(item)
	}
	
	// If we haven't set up any expiration check timer or found a more imminent item.
	//if item.lifeSpan > 0 && (expDur == 0 || item.lifeSpan < expDur) {
	//	go table.expirationCheck()
	//}
}

// Add adds a key/value pair to the cache.
// Parameter key is the item's cache-key.
// Parameter lifeSpan determines after which time period without an access the item
// will get removed from the cache.
// Parameter data is the item's value.
func (table *FileItems) Add(key, datapath string, lifeSpan time.Duration, data []byte) *FileItem {
	// Add item to cache.
	return table.AddUnsafe(key, datapath, lifeSpan, data)
}

func (table *FileItems) AddUnsafe(key, datapath string, lifeSpan time.Duration, data []byte) *FileItem {
	// Add item to cache.
	if _, ok := table.items.Load(key); ok {
		log.Infof("FileItems key exist :%v table name:%v", key, table.name)

		return nil
	}

	len := int64(len(data))
	//newTotal := table.totalSize + len
	//if newTotal >= table.limitSize {
	//	if table.overSize != nil {
	//		if table.overSize(key, datapath, lifeSpan, data) {
	//
	//		}
	//	}
	//	log.Infof("FileItems oversize :%v table name:%v", newTotal, table.name)
	//	return nil
	//}

	item := NewFileItem(key, datapath, data, lifeSpan)
	if item == nil {
		return nil
	}
	table.addInternal(item)
	// table.totalSize = newTotal
	table.addTotalSize(len)
	return item
}

func (table *FileItems) AddFromFile(key, datapath string, lifeSpan time.Duration) *FileItem {
	// Add item to cache.
	if _, ok := table.items.Load(key); ok {
		log.Infof("FileItems key exist :%v table name:%v", key, table.name)
		return nil
	}

	now := time.Now()
	item := LoadFileItem(key, datapath, lifeSpan)
	if item == nil {
		log.Errorf("FileItems LoadFileItem failed key:%v", key)
		return nil
	}

	// 超时..
	if now.Sub(item.accessedOn) >= lifeSpan  {
		bDelete := true
		if item.aboutToExpire != nil {
			log.Info("AddFromFile item aboutToExpire ball back.")
			bDelete = item.aboutToExpire(item)
		} else if table.aboutToExpire != nil {
			log.Info("AddFromFile table aboutToExpire ball back.")
			bDelete = table.aboutToExpire(item)
		}

		log.Infof("FileItems over time :%v key:%v now:%v accessedOn:%v", now.Sub(item.accessedOn), key, now, item.accessedOn)
		item.delete(false, bDelete)
		return nil
	}

	//newTotal := table.totalSize + item.accessSize
	//if newTotal >= table.limitSize {
	//	if table.overSize != nil {
	//		if table.overSize(key, datapath, lifeSpan, nil) {
	//			item.delete(false, true)
	//		}
	//	}
	//	log.Infof("FileItems oversize :%v table name:%v", newTotal, table.name)
	//	return nil
	//}

	table.addInternal(item)
	table.addTotalSize(item.accessSize)
	return item
}

// delete the internally.
func (table *FileItems) deleteInternal(key string) (*FileItem, error) {
	r, ok := table.items.Load(key)
	if !ok {
		return nil, ErrKeyNotFound
	}
	
	// Cache value so we don't keep blocking the mutex.
	aboutToExpire := table.aboutToExpire
	
	// Trigger callbacks before deleting an item from cache.
	bDelete := true
	item := r.(*FileItem)
	if aboutToExpire != nil {
		bDelete = aboutToExpire(item)
	}

	// r.delete()
	log.Infof("deleteInternal item with key %v created on %v and hit %v times from table %v", key, item.createdOn, item.accessCount, table.name)
	table.subTotalSize(item.accessSize)
	item.delete(false, bDelete)
	table.items.Delete(key)
	// delete(table.items, key)
	
	return item, nil
}

// Delete an item from the cache.
func (table *FileItems) Delete(key string, bDeleteCallBack bool) (*FileItem, error) {
	// table.Lock()
	// defer table.Unlock()
	r,ok := table.items.Load(key)
	var item *FileItem
	if ok && r != nil {
		item = r.(*FileItem)
		log.Infof("Delete item with key %v created on %v and hit %v times from table %v", key, item.createdOn, item.accessCount, table.name)
		item.delete(bDeleteCallBack, true)
		// table.totalSize -= int64(item.accessSize)
		table.subTotalSize(item.accessSize)
		table.items.Delete(key)
		// delete(table.items, key)
	}
	return item,ErrKeyNotFound
}

//func (table *FileItems) OnlyDelete(item *FileItem) {
//	item.delete(false)
//}

// Exists returns whether an item exists in the cache. Unlike the Value method
// Exists neither tries to fetch data via the loadData callback nor does it
// keep the item alive in the cache.
func (table *FileItems) Exists(key string) bool {
	_, ok := table.items.Load(key)
	
	return ok
}

// NotFoundAdd checks whether an item is not yet cached. Unlike the Exists
// method this also adds data if the key could not be found.
func (table *FileItems) NotFoundAdd(key string, lifeSpan time.Duration, data []byte) bool {
	
	if _, ok := table.items.Load(key); ok {
		return false
	}
	
	item := NewFileItem(key, table.name, data, lifeSpan)
	table.addInternal(item)
	return true
}

// Value returns an item from the cache and marks it to be kept alive. You can
// pass additional arguments to your DataLoader callback function.
func (table *FileItems) Value(key string, args ...interface{}) (*FileItem, error) {
	r, ok := table.items.Load(key)
	loadData := table.loadData
	
	if ok {
		// Update access counter and timestamp.
		item := r.(* FileItem)
		item.AddCount()
		return item, nil
	}
	
	// Item doesn't exist in FileItem. Try and fetch it with a data-loader. may be in another FileItems
	if loadData != nil {
		item := loadData(key, args...)
		if item != nil {
			// TODO: 先直接返回，以后需要再做重新load进这层缓存处理.
			// table.Add(key, item.lifeSpan, item.data)
			// not load any more, just get and return.
			return item, nil
		}
		
		return nil, ErrKeyNotFoundOrLoadable
	}
	
	return nil, ErrKeyNotFound
}

// Flush deletes all items from this cache table.
func (table *FileItems) Flush() {
	log.Infof("Flushing table %v", table.name)
	
	table.items = new(sync.Map)
	table.cleanupInterval = 0
	if table.cleanupTimer != nil {
		table.cleanupTimer.Stop()
	}
	table.totalSize = 0
}
