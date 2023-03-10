package app

import (
	"errors"
	"strings"

	"github.com/lazy-am/zvart/internal/core/contact"
)

func (a *app) CmdApply(cmd string, contactIndex int) error {
	c := cmd[1:]
	sc := strings.Split(c, " ")
	switch sc[0] {
	case "nc":
		return a.addNewContactCmd(sc[1:])
	case "clear":
		return a.clearMessages(contactIndex)
	default:
		return errors.New("unknown command")
	}
}

func (a *app) clearMessages(contactIndex int) error {
	if contactIndex < 1 {
		return errors.New("no contact specified")
	}
	c, err := contact.Load(a.Db, uint64(contactIndex))
	if err != nil {
		return err
	}
	c.ClearMessages(a.Db, a.ErrorNotice)
	return nil
}

func (a *app) addNewContactCmd(cmd []string) error {
	if len(cmd) < 1 {
		return errors.New("no link to create a contact")
	}
	return a.AddNewContact(cmd[0], strings.Join(cmd[1:], " "))
}
