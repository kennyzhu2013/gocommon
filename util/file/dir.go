/*
@Time : 2019/7/2 11:45
@Author : kenny zhu
@File : dir
@Software: GoLand
@Others:
*/
package file

import (
	"common/log/log"
	"container/list"
	"os"
	"path/filepath"
	"time"
)

const (
	IsDirectory = iota
	IsRegular
	IsSymlink
)

type SysFile struct {
	FType  int
	FName  string
	fLink  string
	FSize  int64
	FMtime time.Time
	fPerm  os.FileMode
}

type TreeInfos struct {
	Files []*SysFile
}

func (tree *TreeInfos) Visit(path string, f os.FileInfo, err error) error {
	if f == nil {
		return err
	}
	var tp int
	if f.IsDir() {
		tp = IsDirectory
	} else if (f.Mode() & os.ModeSymlink) > 0 {
		tp = IsSymlink
	} else {
		tp = IsRegular
	}
	inoFile := &SysFile{
		FName:  path,
		FType:  tp,
		fPerm:  f.Mode(),
		FMtime: f.ModTime(),
		FSize:  f.Size(),
	}
	tree.Files = append(tree.Files, inoFile)
	return nil
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsExist(err) {
		return true, nil
	}

	return false, err
}

//
func CreateSubDirIfNotExist(dir, sub string) string {
	// 年月日..
	// nowData := time.Now().Format("2006-01-02")
	subDir := dir + "/" + sub
	if bExist, _ := PathExists(subDir); bExist {
		return subDir
	}

	// create mkdir -p
	_, _, _ = ShellCmd("mkdir -p", subDir, "")

	// change mode
	_, _, _ = ShellCmd("chmod -R 777", subDir, "")
	// _ = os.Mkdir(subDir, os.ModePerm)
	return subDir
}

// fast walk
// Walk walks the file tree rooted at root, calling walkFn for each file or
// directory in the tree, including root. All errors that arise visiting files
// and directories are filtered by walkFn. The files are walked in lexical
// order, which makes the output deterministic but means that for very
// large directories Walk can be inefficient.
// Walk does not follow symbolic links.
func Walk(root string, walkFn filepath.WalkFunc) error {
	info, err := os.Lstat(root)
	if err != nil {
		err = walkFn(root, nil, err)
	} else {
		err = walkFiles(root, info, walkFn)
	}
	if err == filepath.SkipDir {
		return nil
	}
	return err
}

// walk use for loop descends path, calling walkFn.
func walkFiles(path string, info os.FileInfo, walkFn filepath.WalkFunc) error {
	if !info.IsDir() {
		return walkFn(path, info, nil)
	}

	names, err := readDirNames(path)
	err1 := walkFn(path, info, err)
	// If err != nil, walk can't walk into this directory.
	// err1 != nil means walkFn want walk to skip this directory or stop walking.
	// Therefore, if one of err and err1 isn't nil, walk will return.
	if err != nil || err1 != nil {
		// The caller's behavior is controlled by the return value, which is decided
		// by walkFn. walkFn may ignore err and return nil.
		// If walkFn returns SkipDir, it will be handled by the caller.
		// So walk should return whatever walkFn returns.
		return err1
	}
	stack := list.New()
	// for _, name := range names {
	stack.PushBack(names)
	for stack.Len() > 0 {
		curNames := stack.Front()
		stack.Remove(curNames)
		log.Infof("walkFiles names:%v", curNames)
		for _, name := range curNames.Value.([]string) {
			filename := filepath.Join(path, name)
			fileInfo, err := os.Lstat(filename)
			if err == filepath.SkipDir || fileInfo == nil {
				// log.Infof("walkFiles ignored:%v", filename)
				continue
			}

			if fileInfo.IsDir() {
				_ = walkFn(filename, fileInfo, err)

				namesTemp, err := readDirNames(filename)
				if err != nil || len(namesTemp) < 1 {
					continue
				}

				for key, _ := range namesTemp { // add parent.
					namesTemp[key] = name + "/" + namesTemp[key]
				}
				// log.Infof("walkFiles push dir:%v", namesTemp)
				stack.PushBack(namesTemp)
			} else {
				// log.Infof("walkFiles push file:%v", filename)
				_ = walkFn(filename, fileInfo, err)
			}
		}
	}

	return nil
}

// readDirNames reads the directory named by dirname and returns
// a sorted list of directory entries.
func readDirNames(dirname string) ([]string, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return nil, err
	}

	// not need sort.
	return names, nil
}
