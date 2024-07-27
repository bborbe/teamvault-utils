package teamvault

import "context"

func NewCacheConnector(connector Connector) Connector {
	return &cacheConnector{
		connector: connector,
		passwords: make(map[Key]Password),
		users:     make(map[Key]User),
		urls:      make(map[Key]Url),
		files:     make(map[Key]File),
	}
}

type cacheConnector struct {
	connector Connector
	passwords map[Key]Password
	users     map[Key]User
	urls      map[Key]Url
	files     map[Key]File
}

func (c *cacheConnector) Password(ctx context.Context, key Key) (Password, error) {
	value, ok := c.passwords[key]
	if ok {
		return value, nil
	}
	value, err := c.connector.Password(ctx, key)
	if err == nil {
		c.passwords[key] = value
	}
	return value, err
}

func (c *cacheConnector) User(ctx context.Context, key Key) (User, error) {
	value, ok := c.users[key]
	if ok {
		return value, nil
	}
	value, err := c.connector.User(ctx, key)
	if err == nil {
		c.users[key] = value
	}
	return value, err
}

func (c *cacheConnector) Url(ctx context.Context, key Key) (Url, error) {
	value, ok := c.urls[key]
	if ok {
		return value, nil
	}
	value, err := c.connector.Url(ctx, key)
	if err == nil {
		c.urls[key] = value
	}
	return value, err
}

func (c *cacheConnector) File(ctx context.Context, key Key) (File, error) {
	value, ok := c.files[key]
	if ok {
		return value, nil
	}
	value, err := c.connector.File(ctx, key)
	if err == nil {
		c.files[key] = value
	}
	return value, err
}

func (c *cacheConnector) Search(ctx context.Context, key string) ([]Key, error) {
	return c.connector.Search(ctx, key)
}
