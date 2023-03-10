package app

import (
	"errors"
	"fmt"

	"github.com/lazy-am/zvart/internal/core/contact"
	"github.com/lazy-am/zvart/internal/core/server"
	"github.com/lazy-am/zvart/internal/core/sound"
	"github.com/lazy-am/zvart/internal/core/tmes"
	"github.com/lazy-am/zvart/internal/storage"
	"github.com/lazy-am/zvart/internal/torl"
)

const (
	Version = "0.0.4 alpha"
)

type app struct {
	Db          storage.Storage
	Tor         *torl.Torlancher
	Server      *server.Server
	Clipboard   bool
	Sound       *sound.AppSound
	ErrorNotice chan error
}

var Zvart app

func (a *app) AddNewContact(link, himessage string) error {
	_, err := contact.NewFromLink(link, himessage, a.Db)
	if err != nil {
		return err
	}
	return nil
}

func (a *app) SendTextTo(index uint64, mes []string) error {
	text := ""
	for _, s := range mes {
		text += s + "\n"
	}
	c, err := contact.Load(a.Db, index)
	if err != nil {
		return errors.New("no contact selected")
	}
	m, err := tmes.Create(a.Db, c.DbMessagesTableName, text)
	if err != nil {
		return err
	}
	if c.FirstUnsentMessageId == nil {
		c.FirstUnsentMessageId = m.GetDBkey()
		c.Save(a.Db)
	}
	return nil
}

func (a *app) GetStatus() string {

	r := ""
	if err := a.Tor.GetError(); err != nil {
		r += fmt.Sprint(err)
	}
	if err := a.Db.GetError(); err != nil {
		r += fmt.Sprint(err)
	}
	if !a.Clipboard {
		r += "clipboard not available"
	}

	return r
}
