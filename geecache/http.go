package geecache

import (
	"fmt"
	"geecache/consistenthash"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

/*

	http.ListenAndServe(_address, _handler)

	type Handler interface {
    	ServeHTTP(w ResponseWriter, r *Request)
	}

	=> an object that implement ServeHTTP can be _handler

*/

const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)

// Http Server
// implement PeerPicker for a pool of Http peers
/*
	1. Provide http service
	2. Based on the key, create http client to get the cache value from remote node
*/
type HttpPool struct {
	// record address: hostname / ip, port: e.g. "https://example.net:8000"
	self string
	// prefix of address: /_geecache/
	basePath string
	// => http://example.com/_geecache/ 
	mu sync.Mutex // guard peers and httpGetters
	peers *consistenthash.Map // consistenthash.go: choose node based on key
	httpGetters map[string]*httpGetter // mapping remote node to httpGetter
}

func NewHttpPool(self string) *HttpPool {
	return &HttpPool{
		self: self,
		basePath: defaultBasePath,
	}
}

// implement the consistent hash algorithm and add all the peers
// and create a http client for all the peers
func (p *HttpPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseUrl: peer + p.basePath}
	}
}

// capsulate the Get() in consistent hash, return the http client of the peer
func (p *HttpPool) PickPeer(key string) (peer PeerGetter, ok bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	// peer != p.self: target node is not self node
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

var _ PeerPicker = (*HttpPool)(nil)

// Log info with server name
func (p *HttpPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// implement ServeHTTP to handle all http requests
func (p *HttpPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// /<basepath>/<groupname>/<key>
	// check prefix
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)

	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: " + groupName, http.StatusNotFound)
		return
	}

	v, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	// write cache byteview as body of httpRespose
	w.Write(v.ByteSlice())
}

// Http Client
// implement PeerGetter
type httpGetter struct{
	// the address of the remote node: e.g. http://example.com/_geecache/
	baseUrl string
}

func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf("%v%v/%v", h.baseUrl, url.QueryEscape(group), url.QueryEscape(key))

	// Get issues a GET to the specified URL => ServeHTTP
	res, err := http.Get(u)

	if (err != nil) {
		return nil, err
	}

	// The response body is streamed on demand as the Body field is read.
	// It is the caller's responsibility to close Body. 
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Server return: %v", res.Status)
	}

	bytes, err := io.ReadAll(res.Body)

	if (err != nil) {
		return nil, fmt.Errorf("Response body: %v", err)
	}

	return bytes, nil

}

var _ PeerGetter = (*httpGetter)(nil)