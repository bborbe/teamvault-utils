package teamvault

type Config struct {
	Url          Url      `json:"url"`
	User         User     `json:"user"`
	Password     Password `json:"pass"`
	Offline      bool     `json:"offline,omitempty"`
	CacheEnabled bool     `json:"cacheEnabled,omitempty"`
}
