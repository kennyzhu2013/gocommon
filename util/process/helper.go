/*
@Time : 2019/9/25 11:54
@Author : kenny zhu
@File : helper
@Software: GoLand
@Others:
*/
package process

import (
	"os/exec"
	"strings"
)

// find
func IsProcessExist(appName string) bool {
	const shellToUse = "bash"
	command := "ps  -C " + appName
	cmd := exec.Command(shellToUse, "-c", command)
	// cmd := exec.Command("ps", "-C", appName)
	output, _ := cmd.Output()

	fields := strings.Fields(string(output))

	for _, v := range fields {
		if v == appName {
			return true
		}
	}

	return false
}

// find
func ProcessCount(appName string) int {
	const shellToUse = "bash"
	result := 0
	command := "ps  -C " + appName
	cmd := exec.Command(shellToUse, "-c", command)
	// cmd := exec.Command("ps", "-C", appName)
	output, _ := cmd.Output()

	fields := strings.Fields(string(output))

	for _, v := range fields {
		if v == appName {
			result++
		}
	}

	return result
}

func ExecProcess(path, appName string) ([]byte, error) {
	const shellToUse = "bash"
	command := "cd " + path + ";" + "./" + appName + " &"
	// localPath := "./"    // app路径
	// cmd := exec.Command(localPath + appName + "&")
	cmd := exec.Command(shellToUse, "-c", command)
	return cmd.Output()
}

// channel的安全关闭.
func SafeCloseNull(e *chan interface{}) {
	if e == nil || *e == nil {
		return
	}

	select {
	case _, OK := <-*e: // closed already.
		if !OK {
			return
		}
	default:
		// null to do.
	}
	close(*e)
	_, OK := <-*e
	if !OK {
		*e = nil
	}
}
