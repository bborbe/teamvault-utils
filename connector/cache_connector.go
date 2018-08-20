package connector

import (
	"github.com/bborbe/teamvault-utils"
)

type Cache struct {
	Connector teamvault.Connector
	Passwords map[teamvault.Key]teamvault.Password
	Users     map[teamvault.Key]teamvault.User
	Urls      map[teamvault.Key]teamvault.Url
	Files     map[teamvault.Key]teamvault.File
}

func NewCache(connector teamvault.Connector) *Cache {
	return &Cache{
		Connector: connector,
		Passwords: make(map[teamvault.Key]teamvault.Password),
		Users:     make(map[teamvault.Key]teamvault.User),
		Urls:      make(map[teamvault.Key]teamvault.Url),
		Files:     make(map[teamvault.Key]teamvault.File),
	}
}

func (c *Cache) Password(key teamvault.Key) (teamvault.Password, error) {
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

func (c *Cache) User(key teamvault.Key) (teamvault.User, error) {
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

func (c *Cache) Url(key teamvault.Key) (teamvault.Url, error) {
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

func (c *Cache) File(key teamvault.Key) (teamvault.File, error) {
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

func (c *Cache) Search(key string) ([]teamvault.Key, error) {
	return c.Connector.Search(key)
}
