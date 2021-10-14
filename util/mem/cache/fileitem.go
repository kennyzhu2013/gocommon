package cache

import (
	"common/log/log"
	"common/util/file"
	"io/ioutil"
	"os"
	"sync"
	"time"
)


// Parameter data contains the user-set value in the cache.
// support file .
type FileItem struct {
	sync.RWMutex

	// The item's key, usually id.
	key string
	datapath string

	// The item's data for file.
	filepath string
	fileIdx  string
	// fData *os.File
	// fIdx *os.File

	// How long will the item live in the cache when not being accessed/kept alive.
	lifeSpan time.Duration

	// Creation timestamp.
	createdOn time.Time

	// Last access timestamp.
	accessedOn time.Time

	// How big the item was stored.
	accessSize int64
	accessCount int

	// Callback method triggered right before removing the item from the cache
	aboutToExpire func(*FileItem) (bDelete bool)
	aboutToDelete func(*FileItem) (bDelete bool)
}


// NewCacheItem returns a newly created CacheItem.
// Parameter lifeSpan determines after which time period without an access the item
// will get removed from the cache.
func NewFileItem(key, dataPath string, data []byte, lifeSpan time.Duration) *FileItem {
	t := time.Now()
	item := &FileItem{
		key:           key,
		datapath:      dataPath,
		lifeSpan:      lifeSpan,
		createdOn:     t,
		accessedOn:    t,
		accessSize:    int64(len(data)),
		accessCount:   0,
		aboutToExpire: nil,
		aboutToDelete: nil,
	}
	if err := item.createAndStoreToFile(data, dataPath); err != nil {
		log.Errorf("NewFileItem createAndStoreToFile key:%v failed:%v", key, err)
		return nil
	}

	return item
}

// lifeSpan is set by FileItems.
func LoadFileItem(key, dataPath string, lifeSpan time.Duration) *FileItem {
	//
	// t := time.Now()
	item := &FileItem{
		key:           key,
		lifeSpan:      lifeSpan,
		accessSize:    0,
		aboutToExpire: nil,
		aboutToDelete: nil,
	}

	item.datapath = dataPath
	item.filepath = dataPath + "/" + item.key
	item.fileIdx  = dataPath + "/" + item.key + ".idx"
	fData,err    := os.Open(item.fileIdx)
	fBody,_    := os.Open(item.filepath)
	if err != nil {
		// not exist idx, create
		// fSource,err := os.Open(item.filepath)
		// if err != nil {
		//	return  nil
		// }
		//
		// fBody,_    := os.Open(item.filepath)
		if fBody != nil {
			b, _ := ioutil.ReadAll( fBody )
			_ = fBody.Close()
			t := time.Now()
			item.createdOn = t
			item.accessedOn = t
			item.accessSize = int64(len(b))
			item.storeIdxFile( dataPath )
			fBody = nil
			fData = nil
			// log.Infof("LoadFileItem 1 item.createdOn:%v, err:%v", item.createdOn, err)
			return item
		}
		// log.Infof("LoadFileItem 2 item.createdOn:%v, err:%v", item.createdOn, err)
		return nil
	} else if fBody == nil {
		fData.Close()
		os.Remove(item.fileIdx)
		fBody = nil
		fData = nil
		return nil
	}

	body, _ := ioutil.ReadAll( fBody )
	_ = fBody.Close()

	b, err := ioutil.ReadAll( fData )
	if err != nil {
		fData.Close()
		return nil
	}
	fData.Close()
	fBody = nil
	fData = nil

	// len := Bytes2Octal(b[:4])
	lenB := len(body)

	err = item.createdOn.UnmarshalText(b)
	// log.Infof("LoadFileItem item.createdOn:%v, err:%v", item.createdOn, err)
	item.accessedOn = item.createdOn
	item.accessSize = int64(lenB)
	return item
}

// lifeSpan is set by FileItems.
// sourceDelete: control whether to delete source file
func (item *FileItem) delete(bCallBack bool, sourceDelete bool)  {
	item.Lock()
	defer item.Unlock()

	if bCallBack {
		if item.aboutToDelete != nil {
			// delete return decide whether to delete.
			sourceDelete = item.aboutToDelete(item)
		}
	}

	if sourceDelete {
		if err := os.Remove( item.filepath ); err != nil {
			log.Warnf("delete failed:%v :%v, may be moved.", item.key, err)
		}
	}

	if err := os.Remove( item.fileIdx ); err != nil {
		log.Errorf("delete failed:%v :%v", item.key, err)
	}
	item.key = ""
	item.accessSize = 0
}

// write here, config in future.
func (item *FileItem) createAndStoreToFile(data []byte, dataPath string) error {
	var err error
	var fData *os.File
	var fIdx *os.File

	_ = file.CreateSubDirIfNotExist(dataPath, "")
	item.filepath = dataPath + "/" + item.key
	item.fileIdx = dataPath + "/" + item.key + ".idx"
	fData,err = os.Create(item.filepath)
	if err != nil {
		return err
	}

	_, err = fData.Write(data)
	if err != nil {
		_ = os.Remove( item.filepath )
		return err
	}
	fData.Close()

	fIdx,err = os.Create(item.fileIdx)
	if err != nil {
		os.Remove( item.filepath )
		return err
	}

	// write access file time, only write createdOn time.
	// _, err = fIdx.Write([]byte(Octal2bytes(item.accessSize)))
	b,_ := item.accessedOn.MarshalText()
	_, err = fIdx.Write(b)
	if err != nil {
		os.Remove( item.filepath )
		os.Remove( item.fileIdx )
		return err
	}
	fIdx.Close()

	return nil
}


// write here, config in future.
func (item *FileItem) storeIdxFile(dataPath string) error {
	var err error
	var fIdx *os.File

	_ = file.CreateSubDirIfNotExist(dataPath, "")
	item.fileIdx = dataPath + "/" + item.key + ".idx"
	fIdx,err = os.Create(item.fileIdx)
	if err != nil {
		os.Remove( item.filepath )
		return err
	}

	// write access file time, only write createdOn time.
	// _, err = fIdx.Write([]byte(Octal2bytes(item.accessSize)))
	b,_ := item.accessedOn.MarshalText()
	_, err = fIdx.Write(b)
	if err != nil {
		log.Errorf("storeIdxFile write failed, key:%v..", item.key)
		os.Remove( item.filepath )
		os.Remove( item.fileIdx )
		return err
	}
	fIdx.Close()

	return nil
}


// KeepAlive marks an item to be kept for another expireDuration period.
func (item *FileItem) AddCount() {
	item.accessCount++
	// item.accessCount++
}

// LifeSpan returns this item's expiration duration.
func (item *FileItem) LifeSpan() time.Duration {
	// immutable
	return item.lifeSpan
}

// AccessedOn returns when this item was last accessed.
func (item *FileItem) AccessedOn() time.Time {
	return item.accessedOn
}

// CreatedOn returns when this item was added to the cache.
func (item *FileItem) CreatedOn() time.Time {
	// immutable
	return item.createdOn
}

// AccessCount returns how often this item has been accessed.
func (item *FileItem) AccessSize() int64 {
	return item.accessSize
}

func (item *FileItem) AccessCount() int {
	return item.accessCount
}

// Key returns the key of this cached item.
func (item *FileItem) Key() string {
	// immutable
	return item.key
}

func (item *FileItem) DataPath() string {
	// immutable
	return item.datapath
}

func (item *FileItem) FilePath() string {
	// immutable
	return item.filepath
}

// Data returns the value of this cached item.
func (item *FileItem) Data() ([]byte, error) {
	// load from file.
	fData,err := os.Open(item.filepath)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll( fData )
	if err != nil {
		return nil, err
	}
	fData.Close()

	return b, nil
}

// SetAboutToExpireCallback configures a callback, which will be called right
// before the item is about to be removed from the cache.
// AddAboutToExpireCallback appends a new callback to the AboutToExpire queue
func (item *FileItem) SetAboutToExpireCallback(f func(*FileItem) bool) {
	item.Lock()
	defer item.Unlock()
	item.aboutToExpire = f
}

func (item *FileItem) SetAboutToDeleteCallback(f func(*FileItem) bool) {
	item.Lock()
	defer item.Unlock()
	item.aboutToDelete = f
}

func Octal2bytes(row int64) (bs []byte) {
	bs = make([]byte, 0)
	for i := 0; i < 8; i++ {
		r := row >> uint((7-i)*8) // 12, 8, 4, 0
		bs = append(bs, byte(r))
	}
	return
}

// bb is 8 bytes.
func Bytes2Octal(bb []byte) (value int64) {
	value = int64(0x0000)
	for i, b := range bb {
		ii := uint(b) << uint((7-i)*8)
		value = value | int64(ii)
	}
	return
}
