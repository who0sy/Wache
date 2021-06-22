package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Map struct {
	// 自定义hash函数
	hash Hash
	// 虚拟节点数
	replicas int
	// hash环
	keys []int
	// 虚拟节点和真实节点映射关闭
	HashMap map[int]string
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		HashMap:  make(map[int]string),
	}
	if fn == nil {
		m.hash = crc32.ChecksumIEEE
	}

	return m
}

func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.HashMap[hash] = key
		}
	}
	sort.Ints(m.keys)

}

func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	// 通过二分搜索法寻找第一个匹配的虚拟节点的下标 idx
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	// 如果 idx == len(m.keys)，说明应选择 m.keys[0]，因为 m.keys 是一个环状结构，所以用取余数的方式来处理这种情况。
	return m.HashMap[m.keys[idx%len(m.keys)]]
}
