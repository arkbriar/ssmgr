package manager

// Slave provides interfaces for managing the remote slave.
type Slave interface {}

// slave is the true object of remote slave process. It implements the
// `Slave` interface.
type slave struct {}
