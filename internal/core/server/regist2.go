package server

import (
	"crypto/rsa"
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
	contregistAddress = "regist2"
)

type continuingRegistration struct {
	Name     string
	RemoteId []byte
	Pass     []byte
	PubKey   *rsa.PublicKey
}

func (s *Server) sendPubKey(c *contact.Contact) error {

	u, err := user.Load(s.storage)
	if err != nil {
		return err
	}
	if c.PubKey != nil {
		cr := continuingRegistration{PubKey: &u.PrivKey.PublicKey, Name: u.Name}
		cr.Pass, err = cipher.RSAEncrypt(c.SecretPass, c.PubKey)
		if err != nil {
			return err
		}
		cr.RemoteId, err = cipher.RSAEncrypt(c.RemoteId, c.PubKey)
		if err != nil {
			return err
		}
		js, err := json.Marshal(cr)
		if err != nil {
			return err
		}
		c.LastTryTime = time.Now()
		c.Save(s.storage)
		res, err := socks5.SendViaTor(s.socksPort,
			c.OnionID, 80, contregistAddress, js)
		if err != nil {
			return err
		}
		c, err = contact.Load(s.storage, binary.LittleEndian.Uint64(c.GetDBkey()))
		if err != nil {
			return err
		}
		var allgood bool
		if err := json.Unmarshal([]byte(res), &allgood); err == nil {
			c.LastCallTime = time.Now()
			if allgood {
				c.PubKeySended = true
				c.Save(s.storage)
				return nil
			} else {
				c.Save(s.storage)
				return errors.New("failed attempt to send a key")
			}
		}
	} else {
		return errors.New("the public key is not known")
	}
	return nil
}

func (s *Server) regist2Handler(w http.ResponseWriter, req *http.Request) {
	allgood := false
	enc := json.NewEncoder(w)

	u, err := user.Load(s.storage)
	if err != nil {
		enc.Encode(allgood)
		return
	}

	cr := continuingRegistration{}
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&cr)
	if err != nil {
		enc.Encode(allgood)
		return
	}

	rid, err := cipher.RSADecrypt(cr.RemoteId, u.PrivKey)
	if err != nil {
		enc.Encode(allgood)
		return
	}

	pass, err := cipher.RSADecrypt(cr.Pass, u.PrivKey)
	if err != nil {
		enc.Encode(allgood)
		return
	}

	err = contact.AddPubKey(s.storage, cr.Name, rid, pass, cr.PubKey)
	if err != nil {
		enc.Encode(allgood)
		return
	}

	time.Sleep(time.Millisecond * 200)

	c, err := contact.Load(s.storage, binary.LittleEndian.Uint64(rid))
	if err != nil || c.PubKey == nil || c.ReportedName != cr.Name {
		enc.Encode(allgood)
		return
	}

	if s.player != nil {
		s.player.PlaySound1()
	}

	allgood = true
	enc.Encode(allgood)
}
