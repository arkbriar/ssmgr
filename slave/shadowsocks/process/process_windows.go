// +build windows

package process

import "os"

func alive(pid int) bool {
	p, err := os.FindProcess(pid)
	return err == nil && p != nil
}
