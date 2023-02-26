package server

import (
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"time"

	"github.com/lazy-am/zvart/internal/core/contact"
	"github.com/lazy-am/zvart/internal/core/user"
	"github.com/lazy-am/zvart/internal/torl"
	"github.com/lazy-am/zvart/pkg/socks5"
)

const (
	registrationAddress = "regist"
)

type registRequest struct {
	Name     string
	Hello    string
	OnionID  []byte
	Pass     []byte
	PubKey   *rsa.PublicKey
	RemoteId []byte
}

type registAnswer struct {
	Passed   bool
	RemoteId []byte
}

func (s *Server) registSender(c *contact.Contact) {
	c.InWork = true
	c.Save(s.storage)

	defer func() {
		c.InWork = false
		c.LastTryTime = time.Now()
		c.Save(s.storage)
	}()

	u, err := user.Load(s.storage)
	if err != nil {
		return
	}

	onion, err := torl.GetSelfLink(s.storage)
	if err != nil {
		return
	}
	req := registRequest{Name: u.Name,
		PubKey:   &u.PrivKey.PublicKey,
		Hello:    c.HelloMessage,
		OnionID:  onion,
		Pass:     c.SecretPass,
		RemoteId: c.GetDBkey()}

	js, err := json.Marshal(req)
	if err != nil {
		return
	}

	res, err := socks5.SendViaTor(s.socksPort,
		c.OnionID, 80, registrationAddress, js)
	if err == nil {
		rs := registAnswer{}
		if err := json.Unmarshal([]byte(res), &rs); err == nil {
			c.LastCallTime = time.Now()
			if rs.Passed {
				c.RemoteId = rs.RemoteId
				c.PubKeySended = true
				c.LastCallTime = time.Now()
			}
		}
	}
}

// accepting server-side registration requests
func (s *Server) registHandler(w http.ResponseWriter, req *http.Request) {
	rr := registRequest{}
	ra := registAnswer{Passed: false}
	enc := json.NewEncoder(w)

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&rr)
	if err != nil {
		enc.Encode(ra)
		return
	}

	сon, err := contact.NewAccepted(s.storage,
		rr.OnionID,
		rr.PubKey,
		rr.Hello,
		rr.Name,
		rr.Pass,
		rr.RemoteId)
	if err != nil {
		enc.Encode(ra)
		return
	}

	ra.Passed = true
	ra.RemoteId = сon.GetDBkey()
	enc.Encode(ra)
}
