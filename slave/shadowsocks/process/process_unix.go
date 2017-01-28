// +build linux darwin freebsd

package process

import "syscall"

// Unix kill 0, check if process is alive
func alive(pid int) bool {
	return syscall.Kill(pid, syscall.Signal(0)) == nil
}
