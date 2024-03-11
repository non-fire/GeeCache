package geecache

// a read-only data struct to indicate cache value
type ByteView struct {
	// the real value is saved in b
	// use []bytes to support many data type(string, image...)
	b []byte
}

func (v ByteView) Len() int {
	return len(v.b)
}

// return a copy of data as a form of slices to prevent modifying
// because ByteView is read-only
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b);
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

// return a copy of data as a form of string
func (v ByteView) String() string {
	return string(v.b)
}