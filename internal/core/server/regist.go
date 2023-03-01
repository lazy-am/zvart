package server

import (
	"crypto/rsa"
	"encoding/binary"
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

func (s *Server) registSender(c *contact.Contact) error {
	//reload contact
	c, err := contact.Load(s.storage, binary.LittleEndian.Uint64(c.GetDBkey()))
	if err != nil {
		return err
	}
	c.ServerWork = true
	c.Save(s.storage)

	defer func() {
		c.ServerWork = false
		c.LastTryTime = time.Now()
		c.Save(s.storage)
	}()

	u, err := user.Load(s.storage)
	if err != nil {
		return err
	}

	onion, err := torl.GetSelfLink(s.storage)
	if err != nil {
		return err
	}
	req := registRequest{Name: u.Name,
		PubKey:   &u.PrivKey.PublicKey,
		Hello:    c.HelloMessage,
		OnionID:  onion,
		Pass:     c.SecretPass,
		RemoteId: c.GetDBkey()}

	js, err := json.Marshal(req)
	if err != nil {
		return err
	}

	res, err := socks5.SendViaTor(s.socksPort,
		c.OnionID, 80, registrationAddress, js)
	if err != nil {
		return err
	}
	c, err = contact.Load(s.storage, binary.LittleEndian.Uint64(c.GetDBkey()))
	if err != nil {
		return err
	}
	rs := registAnswer{}
	if err := json.Unmarshal([]byte(res), &rs); err == nil {
		c.LastCallTime = time.Now()
		if rs.Passed {
			c.RemoteId = rs.RemoteId
			c.PubKeySended = true
			c.LastCallTime = time.Now()
			c.NeedUpdateGuiInfo = true
		}
	}
	return nil
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

	if s.player != nil {
		s.player.PlaySound1()
	}

	ra.Passed = true
	ra.RemoteId = сon.GetDBkey()
	enc.Encode(ra)
}
