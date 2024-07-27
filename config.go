package teamvault

type Config struct {
	Url          Url      `json:"url"`
	User         User     `json:"user"`
	Password     Password `json:"pass"`
	CacheEnabled bool     `json:"cacheEnabled,omitempty"`
}
