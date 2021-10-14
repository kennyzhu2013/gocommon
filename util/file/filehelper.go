/*
@Time : 2019/7/2 11:51
@Author : kenny zhu
@File : filehelper
@Software: GoLand
@Others:
*/
package file

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// recursively search file, return full path
func SearchFileByName(dir, fileName string) (error, string)  {
	source := TreeInfos{
		Files: make([]*SysFile, 0),
	}

	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		return source.Visit(path, f, err)
	})
	if err != nil {
		// fmt.Printf("filepath.Walk() returned %v\n", err)
		return err, ""
	}

	for _, v := range source.Files {
		if v.FType == IsRegular {
			// source name must write here.
			tmpName := strings.Split(v.FName, "/")
			sourcename := tmpName[len(tmpName)-1]
			if strings.Compare(sourcename, fileName) == 0 {
				return nil, v.FName
			}

		}
	}

	return errors.New("File not found. "), ""
}

// cp -f , rm -f ..
func ShellCmd(shell, source, dest string) (error, string, string) {
	const shellToUse = "bash"
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command := fmt.Sprintf("%s %s %s", shell, source, dest)
	// log.Debug("ShellCmd:", command)
	cmd := exec.Command(shellToUse, "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil{
		errorString := err.Error()
		if errorString == "exit status 1" {
			return nil, stdout.String(), stderr.String()
		}
	}

	return err, stdout.String(), stderr.String()
}

// shell command expand
func ShellCmdEx(shellFormat string, args ...string) (error, string, string) {
	const shellToUse = "bash"
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	// command := fmt.Sprintf("%s %s %s", shell, source, dest)
	command :=  fmt.Sprintf(shellFormat, args)
	// log.Debug("ShellCmd:", command)
	cmd := exec.Command(shellToUse, "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil{
		errorString := err.Error()
		if errorString == "exit status 1" {
			return nil, stdout.String(), stderr.String()
		}
	}

	return err, stdout.String(), stderr.String()
}
