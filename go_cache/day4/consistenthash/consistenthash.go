package consistenthash

import (
	"fmt"
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

//Map constains all hashed keys
type Map struct {
	hash     Hash
	replicas int
	keys     []int
	hashMap  map[int]string
}

//New creates a Map instance
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}

	return m
}

// add some keys to the hash
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			fmt.Printf("add idx %v-%v %v\n", key, i, hash)
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)

	for k, v := range m.hashMap {
		fmt.Printf(" %v %v ", k, v)
	}
	fmt.Printf("\n")
	for k, v := range m.keys {
		fmt.Printf(" %v %v ", k, v)
	}

	fmt.Printf("\n")
}

func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	//binary search for approoriate replica
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	fmt.Printf("key:%v hash:%v idx:%v index:%v\n", key, hash, idx, m.keys[idx%len(m.keys)])

	return m.hashMap[m.keys[idx%len(m.keys)]]
}
