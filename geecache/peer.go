package geecache

// locate the peer has the key
type PeerPicker interface{
	// choose the node according to the key
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// to be implemented by a peer
type PeerGetter interface {
	// find the cache value
	Get(group string, key string) ([]byte, error)
}