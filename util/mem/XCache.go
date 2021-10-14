package mem

import (
	"common/util/mem/cache"
	"time"
)

// 基于slab内存管理的内存块管理..
// store to file if memory is full.
// 第一步先把文件存到本地根目录下，直接key命名..文件若不在则载入内存..
// 小内存，对于小于100k的文件都分配100k?..
// 文件再考虑直接存到本地硬盘..
// TODO: 此文件需要移到xmedia目录下..

var (
	DefaultCache *XCache
)

type Option func(*XCacheOptions)

type XCache struct {
	TempFsStoreCache *cache.FileItems
	BlockStoreCache  *cache.FileItems

	// early media..
	TempFsEarlyCache *cache.FileItems
	BlockEarlyCache  *cache.FileItems

	// for total control.
	XCacheOptions
}



func NewXCache(bLoad bool, opts ...Option) *XCache {
	options := XCacheOptions{
		nameEarlyTfs: "/tmp/earlymedia",
		nameStoreTfs: "/tmp/storemedia",
		nameEarlyBlock: "/data/earlymedia",
		nameStoreBlock: "/data/storemedia",

		limitsStoreTempFs: 14*1024*1024*1024,
		limitsEarlyTempFs: 2*1024*1024*1024,
		spanTempFs: 10 * time.Minute,
		spanCallTempFs: 10 * time.Minute,

		limitsStoreBlock: 140*1024*1024*1024,
		limitsEarlyBlock: 20*1024*1024*1024,
		spanBlock: 20 * time.Minute, // 2 hours.
		spanCallBlock: 120 * time.Minute,
	}

	for _, o := range opts {
		o(&options)
	}

	X := &XCache{
		XCacheOptions: options,
	}

	// init all
	X.TempFsEarlyCache = cache.LoadFileItems(options.nameEarlyTfs, options.limitsEarlyTempFs, options.spanTempFs, bLoad)
	X.TempFsStoreCache = cache.LoadFileItems(options.nameStoreTfs, options.limitsStoreTempFs, options.spanCallTempFs, bLoad)

	X.BlockEarlyCache = cache.LoadFileItems(options.nameEarlyBlock, options.limitsEarlyBlock, options.spanBlock, bLoad)
	X.BlockStoreCache = cache.LoadFileItems(options.nameStoreBlock, options.limitsStoreBlock, options.spanCallBlock, bLoad)
	return X
}


func (X *XCache) AddEarlyTfsFromFile(key, datapath string) *cache.FileItem {
	return X.TempFsEarlyCache.AddFromFile(key, datapath, X.XCacheOptions.spanTempFs)
}

func (X *XCache) AddStoreTfsFromFile(key, datapath string) *cache.FileItem {
	return X.TempFsStoreCache.AddFromFile(key, datapath, X.XCacheOptions.spanCallTempFs)
}

func (X *XCache) AddEarlyBlockFromFile(key, datapath string) *cache.FileItem {
	return X.BlockEarlyCache.AddFromFile(key, datapath, X.XCacheOptions.spanBlock)
}

func (X *XCache) AddStoreBlockFromFile(key, datapath string) *cache.FileItem {
	return X.BlockStoreCache.AddFromFile(key, datapath, X.XCacheOptions.spanCallBlock)
}

// ========================================下面这些性能待优化==========================================================
// Load from tfs first.
func (X *XCache) GetEarlyFromFile(key string) (*cache.FileItem,error) {
	if item, err := X.TempFsEarlyCache.Value(key); item != nil {
		return item, err
	}

	return X.BlockEarlyCache.Value(key)
}

func (X *XCache) GetStoreFromFile(key string) (*cache.FileItem,error) {
	if item, err := X.TempFsStoreCache.Value(key); item != nil {
		return item, err
	}

	return X.BlockStoreCache.Value(key)
}

// Load from tfs first.
func (X *XCache) DeleteEarlyFromFile(key string) (*cache.FileItem,error) {
	if item, err := X.TempFsEarlyCache.Delete(key, false); item != nil {
		return item, err
	}

	return X.BlockEarlyCache.Delete(key, false)
}

func (X *XCache) DeleteStoreFromFile(key string) (*cache.FileItem,error) {
	if item, err := X.TempFsStoreCache.Delete(key, true); item != nil {
		return item, err
	}

	return X.BlockStoreCache.Delete(key, true)
}
