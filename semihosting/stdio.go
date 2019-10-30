package semihosting

// These three file descriptors are connected to the host stdin/stdout/stderr,
// and can be used for logging.
var (
	Stdin  = File{fd: 0}
	Stdout = File{fd: 1}
	Stderr = File{fd: 2}
)

// File represents a semihosting file descriptor.
type File struct {
	fd uintptr
}

// Write writes the given data buffer to the file descriptor, returning an error
// if the write could not complete successfully.
func (f *File) Write(buf []byte) error {
	return Write(f.fd, buf)
}
