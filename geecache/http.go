package geecache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)
/*

	http.ListenAndServe(_address, _handler)

	type Handler interface {
    	ServeHTTP(w ResponseWriter, r *Request)
	}
	
	=> an object that implement ServeHTTP can be _handler

*/

const defaultBasePath = "/_geecache/"

type HttpPool struct {
	// record address: hostname / ip, port: e.g. "https://example.net:8000"
	self string
	// prefix of address: /_geecache/
	basePath string
	// => http://example.com/_geecache/ 
}

func NewHttpPool(self string) *HttpPool {
	return &HttpPool{
		self: self,
		basePath: defaultBasePath,
	}
}

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