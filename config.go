package teamvault

type TeamvaultConfig struct {
	Url      Url      `json:"url"`
	User     User     `json:"user"`
	Password Password `json:"pass"`
	Offline  bool     `json:"offline"`
}
