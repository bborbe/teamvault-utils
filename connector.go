package teamvault

type Connector interface {
	Password(key TeamvaultKey) (TeamvaultPassword, error)
	User(key TeamvaultKey) (TeamvaultUser, error)
	Url(key TeamvaultKey) (TeamvaultUrl, error)
	File(key TeamvaultKey) (TeamvaultFile, error)
	Search(name string) ([]TeamvaultKey, error)
}
