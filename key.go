package teamvault

type Key string

func (t Key) String() string {
	return string(t)
}
