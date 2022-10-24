package model

import (
	"fmt"
	"github.com/lixiangyun/go-ntlm"
	"testing"
)

func TestInsert(t *testing.T) {
	session, err := ntlm.CreateClientSession(ntlm.Version1, ntlm.ConnectionlessMode)
	if err != nil {
		return
	}
	session.SetUserInfo("auth.Ntlm.Username", "auth.Ntlm.Password", "auth.Ntlm.Domain")
	negotiate, err := session.GenerateNegotiateMessage()
	fmt.Println(negotiate)

}
