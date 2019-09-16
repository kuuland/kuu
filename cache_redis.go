package kuu

// CacherRedis
type CacherRedis struct {
}

// NewCacherRedis
func NewCacherRedis() *CacherRedis {
	return &CacherRedis{}
}

// SetString
func (c *CacherRedis) SetString(key, val string) {
}

// GetString
func (c *CacherRedis) GetString(key string) (val string) {
	return
}

// SetInt
func (c *CacherRedis) SetInt(key string, val int) {
}

// GetInt
func (c *CacherRedis) GetInt(key string) (val int) {
	return
}

// DelCache
func (c *CacherRedis) DelCache(keys ...string) {
	return
}

// Set
func (c *CacherRedis) Set(id string, value string) {
	c.SetString(id, value)
	return
}

// Get
func (c *CacherRedis) Get(id string, clear bool) (val string) {
	val = c.GetString(id)
	if clear {
		DelCache(id)
	}
	return
}

// Close
func (c *CacherRedis) Close() {
}
