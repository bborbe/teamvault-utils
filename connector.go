package teamvault

type Connector interface {
	Password(key Key) (Password, error)
	User(key Key) (User, error)
	Url(key Key) (Url, error)
	File(key Key) (File, error)
	Search(name string) ([]Key, error)
}
