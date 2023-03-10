package mainwindow

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/jroimartin/gocui"
	"github.com/lazy-am/zvart/internal/core/app"
	"github.com/lazy-am/zvart/internal/core/contact"
	"github.com/lazy-am/zvart/internal/core/tmes"
	"github.com/lazy-am/zvart/internal/core/user"
	"github.com/lazy-am/zvart/internal/torl"
	"github.com/lazy-am/zvart/internal/ui/status"
	"github.com/lazy-am/zvart/pkg/formats"
	"golang.design/x/clipboard"
)

const (
	titleView                = "titl"
	contactListView          = "cntlist"
	aboutContactView         = "cntabout"
	mesEditView              = "edit"
	mesListView              = "msg"
	onStartLoadMessageNumber = 100
)

type Window struct {
	G                   *gocui.Gui
	status              *status.Status
	ctx                 context.Context
	concelF             context.CancelFunc
	activeContact       int
	oldPrintedContact   *contact.Contact
	standartHelpPrinted bool
}

func (w *Window) Init() {

	w.G.Cursor = true
	w.G.Mouse = true
	w.G.SetManagerFunc(w.layout)
	w.keybindings()
	w.ctx, w.concelF = context.WithCancel(context.Background())
	w.status = status.Build("")
	w.activeContact = -1
	go w.updateTitle()
	go w.updateContacts()

}

func (w *Window) layout(g *gocui.Gui) error {

	const contactsX = 20

	maxX, maxY := g.Size()

	if v, err := g.SetView(titleView, -1, -1, maxX, 2); err != nil &&
		err == gocui.ErrUnknownView {
		v.Frame = false
	}

	if v, err := g.SetView(contactListView, -1, 1, contactsX, maxY); err != nil &&
		err == gocui.ErrUnknownView {
		v.Editable = false
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
	}

	if v, err := g.SetView(aboutContactView, contactsX, 1, maxX, 5); err != nil &&
		err == gocui.ErrUnknownView {
		v.Frame = false
		v.Wrap = true
	}

	if v, err := g.SetView(mesListView, contactsX, 5, maxX, maxY-7); err != nil &&
		err == gocui.ErrUnknownView {
		v.Frame = true
		v.Editable = false
		v.Autoscroll = true
		v.Wrap = true
	}

	if v, err := g.SetView(mesEditView, contactsX, maxY-7, maxX, maxY); err != nil &&
		err == gocui.ErrUnknownView {
		v.Frame = true
		v.Editable = true
		w.G.SetCurrentView(mesEditView)
	}

	return nil
}

func (w *Window) keybindings() error {

	err := w.G.SetKeybinding("", // ctrl + Q - exit
		gocui.KeyCtrlQ,
		gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			w.concelF()
			return gocui.ErrQuit
		})
	if err != nil {
		return err
	}

	err = w.G.SetKeybinding("", // ctrl + I - copy ID
		gocui.KeyCtrlI,
		gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			if app.Zvart.Clipboard {
				clipboard.Write(clipboard.FmtText, []byte(app.Zvart.Tor.GetHost()))
				w.status.Set(" Your id copied to the clipboard", 10)
			} else {
				w.status.Set(" The clipboard does not work on your system(ctrl+i error)", 10)
			}
			return nil
		})
	if err != nil {
		return err
	}

	err = w.G.SetKeybinding("", // ctrl + V - paste in message edit
		gocui.KeyCtrlV,
		gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			if app.Zvart.Clipboard {
				b := string(clipboard.Read(clipboard.FmtText))
				if len(b) > 0 {
					text := strings.Replace(b, "\r\n", "\n", -1)
					edit, _ := w.G.View(mesEditView)
					fmt.Fprint(edit, text)
					x, y := edit.Cursor()
					strCount := strings.Count(text, "\n")
					lastEndString := strings.LastIndex(text, "\n")
					var xDelta int
					if lastEndString == -1 {
						xDelta = len(text) + x
					} else {
						xDelta = len(text) - (lastEndString + 1)
					}
					edit.SetCursor(xDelta, y+strCount)
					w.status.Set("The text from the clipboard is copied to the input field", 10)
				} else {
					w.status.Set("Nothing on the clipboard", 10)
				}
			} else {
				w.status.Set(" The clipboard does not work on your system(ctrl+v error)", 10)
			}
			return nil
		})
	if err != nil {
		return err
	}

	err = w.G.SetKeybinding("", // ctrl + X - clear input
		gocui.KeyCtrlQ,
		gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			edit, _ := w.G.View(mesEditView)
			edit.Clear()
			edit.SetCursor(0, 0)
			return nil
		})
	if err != nil {
		return err
	}

	err = w.G.SetKeybinding(mesEditView,
		gocui.KeyCtrlTilde,
		gocui.ModNone,
		w.clearActiveContact)
	if err != nil {
		return err
	}

	if err := w.G.SetKeybinding(contactListView,
		gocui.MouseLeft,
		gocui.ModNone,
		w.contactListClk); err != nil {
		return err
	}

	if err := w.G.SetKeybinding(mesEditView,
		gocui.KeyEnter,
		gocui.ModNone,
		w.sendClick); err != nil {
		return err
	}

	err = w.G.SetKeybinding(mesListView, gocui.MouseWheelUp, gocui.ModNone, w.up)
	if err != nil {
		return err
	}
	err = w.G.SetKeybinding(mesListView, gocui.MouseWheelDown, gocui.ModNone, w.down)
	if err != nil {
		return err
	}

	return nil
}

func (w *Window) down(g *gocui.Gui, v *gocui.View) error {
	v.Autoscroll = false
	ox, oy := v.Origin()
	v.SetOrigin(ox, oy+1)
	return nil
}

func (w *Window) up(g *gocui.Gui, v *gocui.View) error {
	v.Autoscroll = false
	ox, oy := v.Origin()
	v.SetOrigin(ox, oy-1)
	return nil
}

func (w *Window) sendClick(g *gocui.Gui, v *gocui.View) error {

	w.G.Update(func(g2 *gocui.Gui) error {

		edit, _ := g2.View(mesEditView)
		lines := edit.BufferLines()

		if len(lines) == 0 {
			return nil
		}

		finalLines := []string{}
		for _, l := range lines {
			if l == "" {
				l = "\n"
			}
			finalLines = append(finalLines, l)
		}

		if fs := finalLines[0]; fs[0] == ':' && len(finalLines) == 1 {
			if err := app.Zvart.CmdApply(fs, w.activeContact); err != nil {
				w.status.Set("Command error: "+err.Error(), 10)
			} else {
				edit.Clear()
				edit.SetCursor(0, 0)
			}
		} else {
			if w.activeContact == -1 {
				w.status.Set("No contact selected", 10)
				return nil
			}
			if err := app.Zvart.SendTextTo(uint64(w.activeContact), finalLines); err != nil {
				w.status.Set("Send text error: "+err.Error(), 10)
			} else {
				edit.Clear()
				edit.SetCursor(0, 0)
			}
		}

		return nil

	})
	return nil
}

func (w *Window) clearActiveContact(g *gocui.Gui, v *gocui.View) error {
	w.status.Set(" Back to home screen", 10)
	w.activeContact = -1
	return nil
}

func (w *Window) contactListClk(g *gocui.Gui, v *gocui.View) error {
	_, y := v.Cursor()
	_, yo := v.Origin()
	w.activeContact = y + yo + 1
	return nil
}

func (w *Window) updateContacts() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			w.rebuildContacts()
		}
	}
}

func (w *Window) printSelfInfo() {
	u, err := user.Load(app.Zvart.Db)
	if err != nil {
		return
	}
	var id []byte
	id, err = torl.GetSelfLink(app.Zvart.Db)
	if err != nil {
		id = []byte("(wait)")
	}
	var tordesc string
	if app.Zvart.Tor.Connected {

		tordesc = "\033[32;1mconnected\033[0m"
	} else {
		tordesc = "\033[31;1mnot connected\033[0m"
	}
	var onidesc string
	if app.Zvart.Tor.OnionConnected {
		onidesc = "\033[32;1mconnected\033[0m"
	} else {
		onidesc = "\033[31;1mnot connected\033[0m"
	}
	v, _ := w.G.View(aboutContactView)
	v.Clear()
	fmt.Fprintf(v, " %s >>> %s \n Part of public key >> %s \n Tor %s | Onion %s",
		u.Name,
		id,
		formats.FormatKey(base64.StdEncoding.EncodeToString(u.PrivKey.PublicKey.N.Bytes())),
		tordesc,
		onidesc)
}

func (w *Window) printAboutContact(c *contact.Contact) {

	name := c.ReportedName
	if name == "" {
		name = "\033[31;1munknown\033[0m"
	}

	var key string
	if c.PubKey == nil {
		key = "\033[31;1mnot received\033[0m"
	} else {
		key = formats.FormatKey(base64.StdEncoding.EncodeToString(c.PubKey.N.Bytes()))
	}

	v, _ := w.G.View(aboutContactView)
	v.Clear()
	fmt.Fprintf(v, " Contact >> %s | Created %s \n Part of key >> %s | ID >> %s \n Last connection %s | Last try %s",
		name,
		formats.FormatTime(&c.CreationTime),
		key,
		c.OnionID,
		formats.FormatTime(&c.LastCallTime),
		formats.FormatTime(&c.LastTryTime))
}

func (w *Window) PrintAllMessages(c *contact.Contact) {
	regist := ""
	h1 := ""
	if c.PubKey == nil {
		h1 = "\n -You cannot write messages to this contact until he/she replies to you \n"
	} else {
		h1 = "\n"
	}

	h2 := "\n -The connection has not yet been established."
	h3 := "\n -\033[33;1mUntil you answer him, he can't write to you anymore.\033[0m "
	if c.CreatedByMe && c.PubKeySended {
		regist = fmt.Sprintf("Added by you, \033[32;1mmessage \"%s\" sent\033[0m"+h1, c.HelloMessage)
	} else if c.CreatedByMe && !c.PubKeySended {
		regist = fmt.Sprintf("Added by you, but \033[31;1mmessage \"%s\" has not yet sent\033[0m"+h2, c.HelloMessage)
	} else if !c.CreatedByMe && !c.PubKeySended {
		regist = fmt.Sprintf("Added from the network, message received \"\033[33;1m%s\033[0m\""+h3, c.HelloMessage)
	} else if !c.CreatedByMe && c.PubKeySended {
		regist = fmt.Sprintf("Added from the network, with a message \"%s\"", c.HelloMessage)
	}

	v, _ := w.G.View(mesListView)
	v.SetOrigin(0, 0)
	v.Autoscroll = true
	v.Clear()
	fmt.Fprint(v, " "+formats.FormatTime(&c.CreationTime)+"\n")
	fmt.Fprint(v, " -"+regist+"\n")

	max, err := tmes.GetMaxId(app.Zvart.Db, c.DbMessagesTableName)
	if err != nil && err.Error() != "empty table" {
		return
	}
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, max)
	// reload and save contact
	c, err = contact.Load(app.Zvart.Db, binary.LittleEndian.Uint64(c.GetDBkey()))
	if err != nil {
		return
	}
	c.LastViewedMessageId = buf
	c.Save(app.Zvart.Db)
	//

	ml, err := tmes.LoadLast(app.Zvart.Db, c.DbMessagesTableName, onStartLoadMessageNumber)
	if err != nil {
		return
	}
	for i, m := range ml {
		if i >= 1 && m.CreationTime.Day() == ml[i-1].CreationTime.Day() {
			w.printMessage(c, m, false)
		} else {
			w.printMessage(c, m, true)
		}
	}

}

func (w *Window) UpdateMessages(c *contact.Contact) {

	viewed := binary.LittleEndian.Uint64(c.LastViewedMessageId)

	max, err := tmes.GetMaxId(app.Zvart.Db, c.DbMessagesTableName)
	if err != nil {
		return
	}

	if viewed < max {
		w.PrintAllMessages(c)
	}
}

func (w *Window) printMessage(c *contact.Contact, m *tmes.TextMessage, longDate bool) {
	var (
		date    string
		t       *time.Time
		recText string
	)
	if m.CreatedByMe {
		t = &m.CreationTime
	} else {
		t = &m.SendedTime
	}
	if longDate {
		date = formats.FormatTime(t)
	} else {
		date = formats.FormatTimeShort(t)
	}
	v, _ := w.G.View(mesListView)
	if m.CreatedByMe {
		if m.IsSended {
			recText = "\033[32;1mreceived\033[0m"
		} else {
			recText = "\033[31;1mnot received\033[0m"
		}
		fmt.Fprint(v, " - \033[33;1mYou\033[0m("+recText+") - "+date+"\n")
	} else {
		fmt.Fprint(v, " - \033[33;1m"+c.ReportedName+"\033[0m - "+date+"\n")
	}

	for _, s := range strings.Split(m.Text, "\n") {
		fmt.Fprint(v, "  "+s+"\n")
	}

	//fmt.Fprint(v, " ------------- \n")
}

func (w *Window) printMessages(c *contact.Contact) {
	if c.NeedUpdateGuiInfo || !c.Equal(w.oldPrintedContact) {
		if c.NeedUpdateGuiInfo {
			// reload and save contact
			c, err := contact.Load(app.Zvart.Db, binary.LittleEndian.Uint64(c.GetDBkey()))
			if err != nil {
				return
			}
			c.NeedUpdateGuiInfo = false
			c.Save(app.Zvart.Db)
			//
			app.Zvart.Sound.PlaySound2()
		}
		w.PrintAllMessages(c)
		w.oldPrintedContact = c
	} else {
		w.UpdateMessages(c)
	}
}

func (w *Window) printHelpInMesList() {
	if w.standartHelpPrinted {
		return
	}
	w.standartHelpPrinted = true
	w.oldPrintedContact = nil
	v, _ := w.G.View(mesListView)
	v.SetOrigin(0, 0)
	v.Clear()
	v.Autoscroll = false
	fmt.Fprint(v, " Welcome to the world of anonymity\n")
	fmt.Fprint(v, " This is one of the first builds of the program and you will meet a lot of bugs\n")
	fmt.Fprint(v, " After the Tor is fully connected you will see a long link to your account\n")
	fmt.Fprint(v, " You can copy this link to the clipboard by pressing \033[33;1mCTRL + I\033[0m\n")
	fmt.Fprint(v, " Share this link with the people you want to talk to\n")
	fmt.Fprint(v, " To add someone to your contact list, enter the following command in the input field below\n")
	fmt.Fprint(v, " \033[33;1m:nc link <some hello message>\033[0m\n")
	fmt.Fprint(v, " Where \"link\" is the same long link received from your friend\n")
	fmt.Fprint(v, " To insert a link into the input box, use \033[33;1mCTRL + V\033[0m\n")
	fmt.Fprint(v, " After entering the command you will have a \"contact\" \n")
	fmt.Fprint(v, " And after a while your respondent will be notified\n")
	fmt.Fprint(v, " He must write you back first\n")
	fmt.Fprint(v, "\n You can help develop the program (translation, programming, donations)\n")
	fmt.Fprint(v, " Visit \033[33;1mgithub.com/lazy-am/zvart\033[0m for details\n")
	fmt.Fprint(v, " Mail \033[33;1mLazyOnPascal@proton.me\033[0m\n")
	fmt.Fprint(v, " --Keyboard shortcuts--\n")
	fmt.Fprint(v, " \033[33;1mCTRL + Tilde\033[0m Exit from the chat room and return to the this screen\n")
	fmt.Fprint(v, " \033[33;1mCTRL + Q\033[0m Program exit\n")
}

func (w *Window) rebuildContacts() {

	cl, err := contact.LoadList(app.Zvart.Db)
	if err != nil {
		return
	}

	w.G.Update(func(g2 *gocui.Gui) error {

		if len(cl) >= w.activeContact && (w.activeContact != -1) {
			w.standartHelpPrinted = false
			w.printAboutContact(cl[w.activeContact-1])
			w.printMessages(cl[w.activeContact-1])
		} else {
			w.printSelfInfo()
			w.printHelpInMesList()
		}

		v, _ := g2.View(contactListView)
		v.Clear()

		for _, c := range cl {

			cline := c.ReportedName
			if cline == "" {
				cline = string(c.OnionID)
			}

			mescount, err := tmes.GetMaxId(app.Zvart.Db, c.DbMessagesTableName)
			if err != nil {
				mescount = 0
			}
			viewed := binary.LittleEndian.Uint64(c.LastViewedMessageId)
			if viewed < mescount {
				fmt.Fprintf(v, " "+cline+"(\033[31;1m%d\033[0m)\n", mescount-viewed)
			} else {
				fmt.Fprint(v, " "+cline+"\n")
			}
		}

		return nil
	})
}

func (w *Window) updateTitle() {
	w.rebuildTitle()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			w.rebuildTitle()
		case <-w.status.Rebuild:
			w.rebuildTitle()
		}
	}
}

func (w *Window) rebuildTitle() {
	if len(app.Zvart.ErrorNotice) > 0 {
		w.status.Set((<-app.Zvart.ErrorNotice).Error(), 10)
	}
	w.G.Update(func(g2 *gocui.Gui) error {
		t, _ := g2.View(titleView)
		t.Clear()
		fmt.Fprintf(t, " Zvart %s | %s",
			app.Version,
			w.status.Get())
		return nil
	})
}
