package server

import (
	"encoding/binary"
	"encoding/json"
	"net/http"
	"time"

	"github.com/lazy-am/zvart/internal/core/contact"
	"github.com/lazy-am/zvart/internal/core/tmes"
	"github.com/lazy-am/zvart/pkg/cipher"
	"github.com/lazy-am/zvart/pkg/socks5"
)

const (
	textmessageAddress = "message"
)

type textMessage struct {
	IDGlobal     []byte
	CreationTime time.Time
	Text         string
}

type encryptedRequest struct {
	RemoteId    []byte
	MessagesEnc []byte
}

func (s *Server) sendMessages(c *contact.Contact) error {
	//reload contact
	c, err := contact.Load(s.storage, binary.LittleEndian.Uint64(c.GetDBkey()))
	if err != nil {
		return err
	}
	c.ServerWork = true
	c.Save(s.storage)

	defer func() {
		c.ServerWork = false
		c.Save(s.storage)
	}()

	//If the public key is not sent, send
	if !c.PubKeySended {
		err = s.sendPubKey(c)
		if err != nil {
			return err
		}
		//reload contact
		c, err = contact.Load(s.storage, binary.LittleEndian.Uint64(c.GetDBkey()))
		if err != nil {
			return err
		}
	}

	//If there is no session key or it has expired, then set
	generated := c.CheckSesKey(s.storage)
	if generated {
		err := s.sendSesKey(c)
		if err != nil {
			return err
		}
		//reload contact
		c, err = contact.Load(s.storage, binary.LittleEndian.Uint64(c.GetDBkey()))
		if err != nil {
			return err
		}
	}

	ml, err := tmes.LoadListFromId(s.storage, c.DbMessagesTableName, c.FirstUnsentMessageId)
	if err != nil {
		return err
	}

	sml := []textMessage{}
	for _, m := range ml {
		if m.CreatedByMe && !m.IsSended {
			sm := textMessage{IDGlobal: m.IDGlobal,
				CreationTime: m.CreationTime,
				Text:         m.Text}
			sml = append(sml, sm)
		}
	}

	jsml, err := json.Marshal(sml)
	if err != nil {
		return err
	}

	encr, err := cipher.AESEncript(c.SessionKey, jsml)
	if err != nil {
		return err
	}

	req := encryptedRequest{RemoteId: c.RemoteId, MessagesEnc: encr}
	jreq, err := json.Marshal(req)
	if err != nil {
		return err
	}

	c.LastTryTime = time.Now()
	c.Save(s.storage)
	res, err := socks5.SendViaTor(s.socksPort,
		c.OnionID, 80, textmessageAddress, jreq)
	if err != nil {
		return err
	}
	c, err = contact.Load(s.storage, binary.LittleEndian.Uint64(c.GetDBkey()))
	if err != nil {
		return err
	}

	allgood := false
	if err := json.Unmarshal([]byte(res), &allgood); err == nil {
		c.LastCallTime = time.Now()
		if allgood {
			for _, m := range ml {
				if m.CreatedByMe && !m.IsSended {
					m.IsSended = true
					m.SendedTime = time.Now()
					m.Save(s.storage)
				}
			}
			//
			ml, err = tmes.LoadListFromId(s.storage, c.DbMessagesTableName, c.FirstUnsentMessageId)
			if err != nil {
				return err
			}
			c.FirstUnsentMessageId = nil
			for _, m := range ml {
				if m.CreatedByMe && !m.IsSended {
					c.FirstUnsentMessageId = m.GetDBkey()
					break
				}
			}
			c.NeedUpdateGuiInfo = true
		}
	}

	return nil
}

// accepting server-side registration requests
func (s *Server) textmesHandler(w http.ResponseWriter, req *http.Request) {
	allgood := false
	enc := json.NewEncoder(w)

	encReq := encryptedRequest{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&encReq)
	if err != nil {
		enc.Encode(allgood)
		return
	}
	id := binary.LittleEndian.Uint64(encReq.RemoteId)
	contact, err := contact.Load(s.storage, id)
	if err != nil {
		enc.Encode(allgood)
		return
	}

	decr, err := cipher.AESDecript(contact.SessionKey, encReq.MessagesEnc)
	if err != nil {
		enc.Encode(allgood)
		return
	}

	rml := []textMessage{}
	err = json.Unmarshal(decr, &rml)
	if err != nil {
		enc.Encode(allgood)
		return
	}

	for _, mr := range rml {
		_, err := tmes.Accepted(s.storage, contact.DbMessagesTableName, mr.IDGlobal, mr.CreationTime, mr.Text)
		if err != nil {
			enc.Encode(allgood)
			return
		}
	}

	s.SoundMessagePlay()

	allgood = true
	enc.Encode(allgood)
}
