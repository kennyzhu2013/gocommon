package cache

import `errors`

var (
	ErrKeyNotFound = errors.New("Key not found in cache ")
	ErrKeyNotFoundOrLoadable = errors.New("Key not found and could not be loaded into cache. ")

	// file errors
	ErrFileOpsFailed = errors.New("File create or write failed. ")
)

