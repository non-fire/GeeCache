package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// self-defind Hash function
type Hash func(data []byte) uint32

type Map struct {
	// Dependency Injection
	hash Hash
	// multiple of virtual node
	replicas int
	// hash ring
	key []int
	// mapping vn to rn
	hashMap map[int]string
}

func New(replicas int, hash Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash: hash,
		hashMap: make(map[int]string),
	}
	if hash == nil {
		hash = crc32.ChecksumIEEE
	}
	return m
}

// accept 1 or n node names
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		// for 1 real node, make m.replicas replicas named "strconv.Itoa(i) + key"
		for i := range m.replicas {
			// caculate the hase value of "strconv.Itoa(i) + key"
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			// add to the hash ring
			m.key = append(m.key, hash)
			// add the mapping of hash value to real node name
			m.hashMap[hash] = key
		}
	}
	// rank on the hash ring
	sort.Ints(m.key)
}

func (m *Map) Get(key string) string {
	if len(key) == 0 {
		return ""
	}
	// calculate the hash value of key
	hash := int(m.hash([]byte(key)))
	// binary search for the first node whose hash value is larger than target's, return idx
	/*
		sort.Search returns the first true index. If there is no such index, Search returns n.
	*/
	idx := sort.Search(len(m.key), func(i int) bool {
		return m.key[i] >= hash
	})
	// if idx == len(m.key) => choose m.key[0]
	// as m.key is a ring
	// idx % len(m.key) => the real idx of the node
	// m.key[idx % len(m.key)] => the hash value of the node
	// => return the node name
	return m.hashMap[m.key[idx % len(m.key)]]
}