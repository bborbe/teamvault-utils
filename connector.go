package teamvault

import "context"

//go:generate counterfeiter -o mocks/connector.go --fake-name Connector . Connector
type Connector interface {
	Password(ctx context.Context, key Key) (Password, error)
	User(ctx context.Context, key Key) (User, error)
	Url(ctx context.Context, key Key) (Url, error)
	File(ctx context.Context, key Key) (File, error)
	Search(ctx context.Context, name string) ([]Key, error)
}
