package psp

// blockLRU is a tiny fixed-capacity cache of decoded CISO blocks, keyed by block
// number. ISO directory and file reads are strongly sequential, so even a small
// cache avoids re-inflating a block that was just touched. Eviction is
// approximate LRU: a FIFO of insertion order, evicting the oldest.

type blockLRU struct {
	cap   int
	m     map[int][]byte
	order []int // insertion order, oldest first
}

func newBlockLRU(capacity int) *blockLRU {
	return &blockLRU{cap: capacity, m: make(map[int][]byte, capacity)}
}

func (c *blockLRU) get(n int) ([]byte, bool) {
	b, ok := c.m[n]
	return b, ok
}

func (c *blockLRU) put(n int, b []byte) {
	if _, ok := c.m[n]; ok {
		c.m[n] = b
		return
	}
	if len(c.order) >= c.cap {
		oldest := c.order[0]
		c.order = c.order[1:]
		delete(c.m, oldest)
	}
	c.m[n] = b
	c.order = append(c.order, n)
}
