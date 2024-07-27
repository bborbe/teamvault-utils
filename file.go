package teamvault

import "encoding/base64"

type File string

func (t File) String() string {
	return string(t)
}

func (t File) Content() ([]byte, error) {
	return base64.StdEncoding.DecodeString(t.String())
}
