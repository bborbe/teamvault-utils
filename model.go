package teamvault

type VariableName string

func (v VariableName) String() string {
	return string(v)
}

type Key string

func (t Key) String() string {
	return string(t)
}

type SourceDirectory string

func (s SourceDirectory) String() string {
	return string(s)
}

type TargetDirectory string

func (t TargetDirectory) String() string {
	return string(t)
}

type Url string

func (t Url) String() string {
	return string(t)
}

type User string

func (t User) String() string {
	return string(t)
}

type Password string

func (t Password) String() string {
	return string(t)
}

type TeamvaultCurrentRevision string

func (t TeamvaultCurrentRevision) String() string {
	return string(t)
}
