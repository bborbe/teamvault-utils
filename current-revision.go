package teamvault

type CurrentRevision string

func (t CurrentRevision) String() string {
	return string(t)
}
