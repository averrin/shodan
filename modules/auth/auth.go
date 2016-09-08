package auth

import (
	"fmt"

	ds "github.com/averrin/shodan/modules/datastream"
	uuid "github.com/satori/go.uuid"
)

type Auth struct {
	dataStream *ds.DataStream
}

func Connect(stream *ds.DataStream) *Auth {
	auth := Auth{}
	auth.dataStream = stream
	return &auth
}

func (auth *Auth) RenewToken() (token string) {
	token = fmt.Sprintf("%s", uuid.NewV4())
	auth.dataStream.SetValue("token", token)
	return token
}

func (auth *Auth) GetToken() (token string) {
	v := ds.Value{}
	auth.dataStream.Get("token", &v)
	if v.Value != nil {
		return v.Value.(string)
	}
	return ""
}

func (auth *Auth) Check(token string) bool {
	orig := auth.GetToken()
	return orig == token
}
