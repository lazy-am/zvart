package mainwindow

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
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
	G                 *gocui.Gui
	status            *status.Status
	ctx               context.Context
	concelF           context.CancelFunc
	activeContact     int
	oldPrintedContact *contact.Contact
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

	const contactsX = 15

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
					edit, _ := w.G.View(mesEditView)
					fmt.Fprint(edit, b)
					x, _ := edit.Cursor()
					edit.SetCursor(len(b)+x, 0)
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

	// err = w.G.SetKeybinding(mesEdtView, //
	// 	gocui.MouseLeft,
	// 	gocui.ModNone,
	// 	func(g *gocui.Gui, v *gocui.View) error {
	// 		w.G.SetCurrentView(messageEdit)
	// 		return nil
	// 	})
	// if err != nil {
	// 	return err
	// }

	// err = g.SetKeybinding("main", gocui.KeyCtrlV, gocui.ModNone, print_help)
	// if err != nil {
	// 	return err
	// }

	// err = g.SetKeybinding("main", gocui.MouseWheelUp, gocui.ModNone, up)
	// if err != nil {
	// 	return err
	// }
	// err = g.SetKeybinding("main", gocui.MouseWheelDown, gocui.ModNone, down)
	// if err != nil {
	// 	return err
	// }

	return nil
}

func (w *Window) sendClick(g *gocui.Gui, v *gocui.View) error {

	w.G.Update(func(g2 *gocui.Gui) error {

		edit, _ := g2.View(mesEditView)
		if len(edit.BufferLines()) == 0 {
			return nil
		}

		if fs := edit.BufferLines()[0]; fs[0] == ':' && len(edit.BufferLines()) == 1 {
			app.CmdDecode(fs)
		} else {
			if w.activeContact < 0 {
				return nil
			}
			app.Zvart.SendTextTo(uint64(w.activeContact), edit.BufferLines())
		}

		edit.Clear()
		edit.SetCursor(0, 0)
		return nil

	})
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
		tordesc = "connected"
	} else {
		tordesc = "not connected"
	}
	var onidesc string
	if app.Zvart.Tor.OnionConnected {
		onidesc = "connected"
	} else {
		onidesc = "not connected"
	}
	v, _ := w.G.View(aboutContactView)
	v.Clear()
	fmt.Fprintf(v, " You >> %s ID >> %s \n Public key >> %s \n Tor %s | Onion %s",
		u.Name,
		formats.FormatKey(string(id)),
		formats.FormatKey(base64.StdEncoding.EncodeToString(u.PrivKey.PublicKey.N.Bytes())),
		tordesc,
		onidesc)
}

func (w *Window) printAboutContact(c *contact.Contact) {

	name := c.ReportedName
	if name == "" {
		name = "unknown"
	}

	var key string
	if c.PubKey == nil {
		key = "not received"
	} else {
		key = formats.FormatKey(base64.StdEncoding.EncodeToString(c.PubKey.N.Bytes()))
	}

	v, _ := w.G.View(aboutContactView)
	v.Clear()
	fmt.Fprintf(v, " Contact >> %s | Created %s \n Key >> %s | ID >> %s (ctrl+j - copy) \n Last connection %s | Last try %s",
		name,
		formats.FormatTime(&c.CreationTime),
		key,
		formats.FormatKey(string(c.OnionID)),
		formats.FormatTime(&c.LastCallTime),
		formats.FormatTime(&c.LastTryTime))
}

func (w *Window) PrintAllMessages(c *contact.Contact) {
	regist := ""
	h1 := "\n -You cannot write messages to this contact until he/she replies to you "
	h2 := "\n -The connection has not yet been established."
	h3 := "\n -Until you answer him, he can't write to you anymore. "
	if c.CreatedByMe && c.PubKeySended {
		regist = fmt.Sprintf("Added by you, message \"%s\" sent"+h1, c.HelloMessage)
	} else if c.CreatedByMe && !c.PubKeySended {
		regist = fmt.Sprintf("Added by you, but message \"%s\" has not yet sent"+h2, c.HelloMessage)
	} else if !c.CreatedByMe && !c.PubKeySended {
		regist = fmt.Sprintf("Added from the network, message received \"%s\""+h3, c.HelloMessage)
	} else if !c.CreatedByMe && c.PubKeySended {
		regist = fmt.Sprintf("Added from the network, with a message \"%s\"", c.HelloMessage)
	}

	v, _ := w.G.View(mesListView)
	v.Clear()
	fmt.Fprint(v, " "+formats.FormatTime(&c.CreationTime)+"\n")
	fmt.Fprint(v, " -"+regist+"\n")

	max, err := tmes.GetMaxId(app.Zvart.Db, c.DbMessagesTableName)
	if err != nil && err.Error() != "empty table" {
		return
	}
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, max)
	c.LastViewedMessageId = buf
	c.Save(app.Zvart.Db)

	ml, err := tmes.LoadLast(app.Zvart.Db, c.DbMessagesTableName, onStartLoadMessageNumber)
	if err != nil {
		return
	}
	for _, m := range ml {
		w.printMessage(&m)
	}

}

func (w *Window) UpdateMessages(c *contact.Contact) {

	viewed := binary.LittleEndian.Uint64(c.LastViewedMessageId)

	max, err := tmes.GetMaxId(app.Zvart.Db, c.DbMessagesTableName)
	if err != nil {
		return
	}

	if viewed < max {
		viewed++
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, viewed)
		ml, err := tmes.LoadListFromId(app.Zvart.Db, c.DbMessagesTableName, buf)
		if err != nil {
			return
		}
		for _, m := range ml {
			w.printMessage(m)
		}
		binary.LittleEndian.PutUint64(buf, max)
		c.LastViewedMessageId = buf
		c.Save(app.Zvart.Db)
	}
}

func (w *Window) printMessage(m *tmes.TextMessage) {
	v, _ := w.G.View(mesListView)
	fmt.Fprint(v, " "+formats.FormatTime(&m.SendedTime)+"\n")
	fmt.Fprint(v, " -"+m.Text+"\n")
}

func (w *Window) printMessages(c *contact.Contact) {
	if !c.Equal(w.oldPrintedContact) {
		w.PrintAllMessages(c)
		w.oldPrintedContact = c
	} else {
		w.UpdateMessages(c)
	}
}

func (w *Window) printHelpInMesList() {
	w.oldPrintedContact = nil
	v, _ := w.G.View(mesListView)
	v.Clear()
	fmt.Fprint(v, "Help help help")
}

func (w *Window) rebuildContacts() {

	cl, err := contact.LoadList(app.Zvart.Db)
	if err != nil {
		return
	}

	w.G.Update(func(g2 *gocui.Gui) error {

		if len(cl) >= w.activeContact && (w.activeContact != -1) {
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
			fmt.Fprint(v, " "+cline)
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
	// u, err := user.Load(app.Zvart.Db)
	// if err != nil {
	// 	return
	// }
	w.G.Update(func(g2 *gocui.Gui) error {

		// s := app.Zvart.GetStatus()
		w.status.Notes(app.Zvart.Notifications)
		t, _ := g2.View(titleView)
		t.Clear()
		fmt.Fprintf(t, " Zvart %s | %s",
			app.Version,
			w.status.Get())
		// if s == "" {
		//
		// } else {
		// 	fmt.Fprintf(t, " Zvart %s | %s | Error: %s",
		// 		app.Version,
		// 		u.Name,
		// 		s)
		// }

		// t2, _ := g2.View(actLogView)
		// t2.Clear()
		// fmt.Fprint(t2, w.status.Get())

		return nil
	})
}
