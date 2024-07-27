package teamvault

type Staging bool

func (s Staging) Bool() bool {
	return bool(s)
}
