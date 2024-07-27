package teamvault

import "context"

type cacheConnector struct {
	Connector Connector
	Passwords map[Key]Password
	Users     map[Key]User
	Urls      map[Key]Url
	Files     map[Key]File
}

func NewCacheConnector(connector Connector) Connector {
	return &cacheConnector{
		Connector: connector,
		Passwords: make(map[Key]Password),
		Users:     make(map[Key]User),
		Urls:      make(map[Key]Url),
		Files:     make(map[Key]File),
	}
}

func (c *cacheConnector) Password(ctx context.Context, key Key) (Password, error) {
	value, ok := c.Passwords[key]
	if ok {
		return value, nil
	}
	value, err := c.Connector.Password(ctx, key)
	if err == nil {
		c.Passwords[key] = value
	}
	return value, err
}

func (c *cacheConnector) User(ctx context.Context, key Key) (User, error) {
	value, ok := c.Users[key]
	if ok {
		return value, nil
	}
	value, err := c.Connector.User(ctx, key)
	if err == nil {
		c.Users[key] = value
	}
	return value, err
}

func (c *cacheConnector) Url(ctx context.Context, key Key) (Url, error) {
	value, ok := c.Urls[key]
	if ok {
		return value, nil
	}
	value, err := c.Connector.Url(ctx, key)
	if err == nil {
		c.Urls[key] = value
	}
	return value, err
}

func (c *cacheConnector) File(ctx context.Context, key Key) (File, error) {
	value, ok := c.Files[key]
	if ok {
		return value, nil
	}
	value, err := c.Connector.File(ctx, key)
	if err == nil {
		c.Files[key] = value
	}
	return value, err
}

func (c *cacheConnector) Search(ctx context.Context, key string) ([]Key, error) {
	return c.Connector.Search(ctx, key)
}
