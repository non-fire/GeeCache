package lru

import "container/list"

type Cache struct {
	// max bytes allowed to be used
	maxBytes int64
	// bytes used currently
	nBytes int64
	ll     *list.List
	cache  map[string]*list.Element
	// callback func when an entry is removed, can be nil
	OnEvicted func(key string, value Value)
}

// datatype of the list node
type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

// the constructor of Cache
func New(maxBytes int64, OnEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: OnEvicted,
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	// find the element by key
	// if exist
	if element, ok := c.cache[key]; ok {
		// move this element to the front of ll
		c.ll.MoveToFront(element)
		kv := element.Value.(*entry) // convert the element.Value to *entry type
		return kv.value, ok
	}
	return
}

func (c *Cache) RemoveOldest() {
	element := c.ll.Back()
	// if ll has a back element
	if element != nil {
		// remove this element from ll
		c.ll.Remove(element)
		// get the key and value of the elemnt
		kv := element.Value.(*entry)
		// delete it from the map
		delete(c.cache, kv.key)
		// modify the current byte
		c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		// run the callback func
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	// if exist, modify value
	if element, ok := c.cache[key]; ok {
		kv := element.Value.(*entry)
		c.nBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
		c.ll.MoveToFront(element)
		// if is not exist, add to the front
	} else {
		newElement := c.ll.PushFront(&entry{key, value})
		c.cache[key] = newElement
		c.nBytes += int64(len(key)) + int64(value.Len())
	}
	// remove the back element until current bytes <= maxBytes
	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
