package teamvault

type Cache struct {
	Connector Connector
	Passwords map[Key]Password
	Users     map[Key]User
	Urls      map[Key]Url
	Files     map[Key]File
}

func NewCache(connector Connector) Connector {
	return &Cache{
		Connector: connector,
		Passwords: make(map[Key]Password),
		Users:     make(map[Key]User),
		Urls:      make(map[Key]Url),
		Files:     make(map[Key]File),
	}
}

func (c *Cache) Password(key Key) (Password, error) {
	value, ok := c.Passwords[key]
	if ok {
		return value, nil
	}
	value, err := c.Connector.Password(key)
	if err == nil {
		c.Passwords[key] = value
	}
	return value, err
}

func (c *Cache) User(key Key) (User, error) {
	value, ok := c.Users[key]
	if ok {
		return value, nil
	}
	value, err := c.Connector.User(key)
	if err == nil {
		c.Users[key] = value
	}
	return value, err
}

func (c *Cache) Url(key Key) (Url, error) {
	value, ok := c.Urls[key]
	if ok {
		return value, nil
	}
	value, err := c.Connector.Url(key)
	if err == nil {
		c.Urls[key] = value
	}
	return value, err
}

func (c *Cache) File(key Key) (File, error) {
	value, ok := c.Files[key]
	if ok {
		return value, nil
	}
	value, err := c.Connector.File(key)
	if err == nil {
		c.Files[key] = value
	}
	return value, err
}

func (c *Cache) Search(key string) ([]Key, error) {
	return c.Connector.Search(key)
}
