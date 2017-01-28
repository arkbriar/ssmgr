package process

// Alive returns if the process is still alive
func Alive(pid int) bool {
	return alive(pid)
}
