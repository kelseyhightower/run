package run

type cache struct {
	data map[string]string
}

func (c *cache) Set(key, value string) {
	c.data[key] = value
	return
}

func (c *cache) Get(key string) string {
	if value, ok := c.data[key]; ok {
		return value
	}

	return ""
}
