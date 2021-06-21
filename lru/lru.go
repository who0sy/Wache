package lru

import "container/list"

// LRU缓存结构体，此结构并发访问是不安全的。
type Cache struct {
	maxBytes  int64                         // 允许使用的最大内存
	nbytes    int64                         // 已使用的内存量
	ll        *list.List                    // 双向链表
	cache     map[string]*list.Element      // 双向链表中对应节点的指针
	OnEvicted func(key string, value Value) // 某条记录被移除时的回调函数，可以为 nil。
}

// 链表节点存储的具体数据
type entry struct {
	key   string
	value Value // 为了通用性，允许值是实现了 Value 接口的任意类型
}

//使用Len来计算它需要多少字节
type Value interface {
	Len() int
}

// 创建新的缓存结构体
func New(maxBytes int64, OnEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: OnEvicted,
	}
}

// 获取查找键的值
func (c *Cache) Get(key string) (value Value, ok bool) {
	if element, ok := c.cache[key]; ok {
		// 将此元素移至双端链表的头部
		c.ll.MoveToFront(element)
		entry := element.Value.(*entry)
		return entry.value, true
	}
	return nil, false

}

// 这里的删除，实际上是缓存淘汰。即移除最近最少访问的节点（队尾）
func (c *Cache) RemoveOldest() bool {
	// 获取双端链表的尾部元素
	element := c.ll.Back()
	if element != nil {
		c.ll.Remove(element)
		entry := element.Value.(*entry)
		delete(c.cache, entry.key)
		c.nbytes -= int64(len(entry.key)) + int64(entry.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(entry.key, entry.value)
		}
	}
	return true
}

// 新增或更新
func (c *Cache) Add(key string, value Value) {

	// 如果存在则更新节点的值
	if element, ok := c.cache[key]; ok {
		entry := element.Value.(*entry)

		// 更新内存使用量
		c.nbytes += int64(value.Len()) - int64(entry.value.Len())

		// 更新值
		entry.value = value

		// 移至新元素到双端链表首位
		c.ll.MoveToFront(element)
	} else {
		element := c.ll.PushFront(&entry{key, value})
		c.cache[key] = element
		c.nbytes += int64(len(key)) + int64(value.Len())
	}

	// 如果超过了设定的最大值 c.maxBytes，则移除最少访问的节点即从尾部开始，直到可使用内存大于0
	for c.nbytes > 0 && c.nbytes > c.maxBytes {
		c.RemoveOldest()
	}
}


// 获取元素个数
func (c *Cache) Len() int {
	return c.ll.Len()
}
