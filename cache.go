package regolt

type Cacheable interface {
	GetKey() ULID
}

type Cache1CheckContext[T Cacheable] struct {
	Cache  *Cache1[T]
	Entity *T
}

type Cache1Checker[T Cacheable] interface {
	CanCache1(*Cache1CheckContext[T]) bool
}

type Cache1[T Cacheable] struct {
	Checker              Cache1Checker[T]
	DontInsertIfOverflow bool
	MaxSize              int
	Cache                map[ULID]*T
}

func (c *Cache1[T]) resize(e *T) bool {
	if c.MaxSize == 0 {
		return false
	} else if c.MaxSize > 0 {
		if c.Checker != nil && !c.Checker.CanCache1(&Cache1CheckContext[T]{Cache: c, Entity: e}) {
			return false
		}
		if len(c.Cache) >= c.MaxSize {
			if c.DontInsertIfOverflow {
				return false
			}
			for k := range c.Cache {
				delete(c.Cache, k)
				break
			}
		}
		c.Cache[(*e).GetKey()] = e
	}
	// user wants infinite count of entities, if c.MaxSize is less than 0
	return true
}

func (c *Cache1[T]) Get(id ULID) *T {
	return c.Cache[id]
}

func (c *Cache1[T]) Del(id ULID) {
	delete(c.Cache, id)
}

func (c *Cache1[T]) PartiallyUpdate(id ULID, updater func(m *T)) {
	x, ok := c.Cache[id]
	if !ok {
		return
	}
	updater(x)
}

func (c *Cache1[T]) Set(v *T) {
	if !c.resize(v) {
		return
	}
	c.Cache[(*v).GetKey()] = v
}

func (c *Cache1[T]) Size() int {
	return len(c.Cache)
}

type Cache2CheckContext[T Cacheable] struct {
	Cache  *Cache2[T]
	Parent ULID
	Entity *T
}

type Cache2Checker[T Cacheable] interface {
	CanCache2(*Cache2CheckContext[T]) bool
}

type Cache2[T Cacheable] struct {
	Checker              Cache2Checker[T]
	MaxSize1             int
	MaxSize2             int
	TotalMaxSize         int
	total                int
	DontInsertIfOverflow bool
	Cache                map[ULID]map[ULID]*T
}

func (c *Cache2[T]) del(parent, id ULID) {
	p := c.Cache[parent]
	_, j := p[id]
	delete(p, id)
	if j {
		c.total--
	}
}

func (c *Cache2[T]) ins(parent ULID, e *T) {
	m, ok := c.Cache[parent]
	if !ok {
		c.Cache[parent] = map[ULID]*T{(*e).GetKey(): e}
		c.total++
		return
	}
	i := (*e).GetKey()
	_, ok = m[i]
	m[i] = e
	if !ok {
		c.total++
	}
}

func (c *Cache2[T]) resize(parent ULID, e *T) bool {
	switch {
	case c.MaxSize1 == 0 && c.MaxSize2 == 0 && c.TotalMaxSize == 0:
		return false
	case c.TotalMaxSize < 0 || c.MaxSize1 > 0:
		if c.Checker != nil && !c.Checker.CanCache2(&Cache2CheckContext[T]{Cache: c, Parent: parent, Entity: e}) {
			return false
		}
		if c.MaxSize1 == 0 && c.MaxSize2 == 0 && c.total == c.TotalMaxSize {
			if c.DontInsertIfOverflow {
				return false
			}
			for k1, v1 := range c.Cache {
				for k2 := range v1 {
					c.del(k1, k2)
					break
				}
			}
		}
		o, k := c.Cache[parent]
		if k && c.MaxSize2 != 0 && len(o) >= c.MaxSize2 {
			for k := range o {
				c.del(parent, k)
				return true
			}
		}
		if c.MaxSize1 != 0 && len(c.Cache) >= c.MaxSize1 && !k {
			for k := range c.Cache {
				c.total -= len(c.Cache[k])
				delete(c.Cache, k)
				return true
			}
		}
	}
	return true
}

func (c *Cache2[T]) Get(parent, id ULID) *T {
	m, ok := c.Cache[parent]
	if ok {
		return m[id]
	}
	return nil
}

func (c *Cache2[T]) Del(parent, id ULID) {
	c.del(parent, id)
}

func (c *Cache2[T]) DelGroup(parent ULID) {
	m, ok := c.Cache[parent]
	if ok {
		c.total -= len(m)
		delete(c.Cache, parent)
	}
}

func (c *Cache2[T]) GroupsCount() int {
	return len(c.Cache)
}

func (c *Cache2[T]) Size() int {
	return c.total
}

func (c *Cache2[T]) PartiallyUpdate(parent, id ULID, updater func(m *T)) {
	x, ok := c.Cache[parent]
	if !ok {
		return
	}
	y, ok := x[id]
	if ok {
		updater(y)
	}
}

func (c *Cache2[T]) Set(parent ULID, v *T) {
	if c.resize(parent, v) {
		c.ins(parent, v)
	}
}

func (c *Cache1[T]) init() {
	if c.Cache == nil {
		c.Cache = map[ULID]*T{}
	}
}

func (c *Cache2[T]) init() {
	if c.Cache == nil {
		c.Cache = map[ULID]map[ULID]*T{}
	}
}

type GenericCache struct {
	Channels *Cache1[OptimizedChannel]
	Emojis   *Cache1[OptimizedCustomEmoji]
	Messages *Cache2[OptimizedMessage]
	Members  *Cache2[Member]
	Roles    *Cache2[OptimizedRole]
	Servers  *Cache1[OptimizedServer]
	Users    *Cache1[OptimizedUser]
	Webhooks *Cache1[OptimizedWebhook]
}

func (gc *GenericCache) init() {
	if gc.Channels == nil {
		gc.Channels = &Cache1[OptimizedChannel]{}
	}
	if gc.Emojis == nil {
		gc.Emojis = &Cache1[OptimizedCustomEmoji]{}
	}
	if gc.Messages == nil {
		gc.Messages = &Cache2[OptimizedMessage]{}
	}
	if gc.Members == nil {
		gc.Members = &Cache2[Member]{}
	}
	if gc.Roles == nil {
		gc.Roles = &Cache2[OptimizedRole]{}
	}
	if gc.Servers == nil {
		gc.Servers = &Cache1[OptimizedServer]{}
	}
	if gc.Users == nil {
		gc.Users = &Cache1[OptimizedUser]{}
	}

	gc.Channels.init()
	gc.Emojis.init()
	gc.Messages.init()
	gc.Members.init()
	gc.Roles.init()
	gc.Servers.init()
	gc.Users.init()
}

const InfiniteCache int = -1
