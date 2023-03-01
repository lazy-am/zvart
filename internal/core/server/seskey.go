package server

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/lazy-am/zvart/internal/core/contact"
	"github.com/lazy-am/zvart/internal/core/user"
	"github.com/lazy-am/zvart/pkg/cipher"
	"github.com/lazy-am/zvart/pkg/socks5"
)

const (
	seskeyAddress = "seskey"
)

type sessionKeySetUpRequest struct {
	RemoteId []byte
	Pass     []byte
	SesKey   []byte
}

func (s *Server) checkSessionKey(c *contact.Contact) (*contact.Contact, error) {
	generated := c.CheckSesKey(s.storage)
	if generated {
		return s.sendSesKey(c)
	}
	return c, nil
}

func (s *Server) sendSesKey(c *contact.Contact) (*contact.Contact, error) {

	sksr := sessionKeySetUpRequest{RemoteId: c.RemoteId,
		Pass:   c.SecretPass,
		SesKey: c.SessionKey}

	js, err := json.Marshal(sksr)
	if err != nil {
		return nil, err
	}

	enc, err := cipher.RSAEncrypt(js, c.PubKey)
	if err != nil {
		return nil, err
	}

	js2, err := json.Marshal(enc)
	if err != nil {
		return nil, err
	}

	c.LastTryTime = time.Now()
	c.Save(s.storage)
	res, err := socks5.SendViaTor(s.socksPort,
		c.OnionID, 80, seskeyAddress, js2)
	if err != nil {
		return nil, err
	}
	c, err = contact.Load(s.storage, binary.LittleEndian.Uint64(c.GetDBkey()))
	if err != nil {
		return nil, err
	}
	allgood := false
	if err := json.Unmarshal([]byte(res), &allgood); err != nil {
		return nil, err
	}
	c.LastCallTime = time.Now()
	if allgood {
		return c, nil
	}
	return nil, errors.New("rebuttal")
}

func (s *Server) sesKeyHandler(w http.ResponseWriter, req *http.Request) {
	allgood := false
	enc := json.NewEncoder(w)

	encryptedBytes := []byte{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&encryptedBytes)
	if err != nil {
		enc.Encode(allgood)
		return
	}

	u, err := user.Load(s.storage)
	if err != nil {
		enc.Encode(allgood)
		return
	}

	decripted, err := cipher.RSADecrypt(encryptedBytes, u.PrivKey)
	if err != nil {
		enc.Encode(allgood)
		return
	}

	sksr := sessionKeySetUpRequest{}
	err = json.Unmarshal(decripted, &sksr)
	if err != nil {
		enc.Encode(allgood)
		return
	}

	err = contact.SetSessionKey(s.storage, sksr.RemoteId, sksr.Pass, sksr.SesKey)
	if err != nil {
		enc.Encode(allgood)
		return
	}

	allgood = true
	enc.Encode(allgood)
}
