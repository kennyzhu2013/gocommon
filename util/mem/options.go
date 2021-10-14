package mem

import "time"

type XCacheOptions struct {
	nameEarlyTfs string
	nameStoreTfs string
	nameEarlyBlock string
	nameStoreBlock string

	limitsStoreTempFs int64
	limitsEarlyTempFs int64
	spanTempFs time.Duration
	spanCallTempFs time.Duration

	limitsStoreBlock int64
	limitsEarlyBlock int64
	spanBlock time.Duration
	spanCallBlock time.Duration
}

func DirEarlyTfs(dir string) Option {
	return func(o *XCacheOptions) {
		o.nameEarlyTfs = dir
	}
}

func DirStoreTfs(dir string) Option {
	return func(o *XCacheOptions) {
		o.nameStoreTfs = dir
	}
}

func DirEarlyBlock(dir string) Option {
	return func(o *XCacheOptions) {
		o.nameEarlyBlock = dir
	}
}

func DirStoreBlock(dir string) Option {
	return func(o *XCacheOptions) {
		o.nameStoreBlock = dir
	}
}

func LimitsEarlyFs(l int64) Option {
	return func(o *XCacheOptions) {
		o.limitsEarlyTempFs = l
	}
}

func LimitsStoreFs(l int64) Option {
	return func(o *XCacheOptions) {
		o.limitsStoreTempFs = l
	}
}

func LimitsEarlyBlock(l int64) Option {
	return func(o *XCacheOptions) {
		o.limitsEarlyBlock = l
	}
}

func LimitsStoreBlock(l int64) Option {
	return func(o *XCacheOptions) {
		o.limitsStoreBlock = l
	}
}

func SpanMemFs(s time.Duration) Option {
	return func(o *XCacheOptions) {
		o.spanTempFs = s
	}
}

func SpanBlock(s time.Duration) Option {
	return func(o *XCacheOptions) {
		o.spanBlock = s
	}
}

func SpanCallMemFs(s time.Duration) Option {
	return func(o *XCacheOptions) {
		o.spanCallTempFs = s
	}
}

func SpanCallBlock(s time.Duration) Option {
	return func(o *XCacheOptions) {
		o.spanCallBlock = s
	}
}
