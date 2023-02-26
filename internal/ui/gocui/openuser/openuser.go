package openuser

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/lazy-am/zvart/internal/core/app"
	"github.com/lazy-am/zvart/internal/core/starter/flags"
	"github.com/lazy-am/zvart/internal/storage"
	"github.com/lazy-am/zvart/internal/ui/gocui/mainwindow"
	"golang.design/x/clipboard"
)

const (
	help = "titl_top"
	edit = "edit1"
	btn  = "next_btn"
)

type Window struct {
	G    *gocui.Gui
	s    storage.Storage
	step int
	name string
	pass string
}

func (w *Window) Init() error {
	w.G.Cursor = true
	w.G.Mouse = true
	w.G.SetManagerFunc(w.layout)
	if err := w.keybindings(); err != nil {
		return err
	}
	var err error
	if w.s, err = app.InitStorage(); err != nil {
		return err
	}
	return nil
}

func (w *Window) layout(g *gocui.Gui) error {

	maxX, _ := g.Size()

	if _, err := g.SetView(help, 1, 1, maxX-2, 20); err != nil &&
		err == gocui.ErrUnknownView {

	}

	if _, err := g.SetView(edit, 3, 8, maxX-4, 10); err != nil &&
		err == gocui.ErrUnknownView {
		w.print1()

	}

	if v, err := g.SetView(btn, maxX-13, 11, maxX-4, 13); err != nil &&
		err == gocui.ErrUnknownView {
		fmt.Fprintf(v, " Unlock")
	}

	return nil
}

func (w *Window) keybindings() error {
	err := w.G.SetKeybinding("", // ctrl + Q - exit
		gocui.KeyCtrlQ,
		gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			return gocui.ErrQuit
		})
	if err != nil {
		return err
	}

	err = w.G.SetKeybinding("", // ctrl + X - clear
		gocui.KeyCtrlX,
		gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			edit, _ := w.G.View(edit)
			edit.Clear()
			edit.SetCursor(0, 0)
			return nil
		})
	if err != nil {
		return err
	}

	err = w.G.SetKeybinding("", // ctrl + V - clear and paste
		gocui.KeyCtrlV,
		gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			edit, _ := w.G.View(edit)
			edit.Clear()
			st := string(clipboard.Read(clipboard.FmtText))
			fmt.Fprintf(edit, st)
			edit.SetCursor(len(st), 0)

			return nil
		})
	if err != nil {
		return err
	}

	err = w.G.SetKeybinding("", // ctrl + Z - mask input
		gocui.KeyCtrlZ,
		gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			edit, _ := w.G.View(edit)
			edit.Mask ^= '*'

			return nil
		})
	if err != nil {
		return err
	}

	if err := w.G.SetKeybinding(btn,
		gocui.MouseLeft,
		gocui.ModNone,
		w.btnclik); err != nil {
		return err
	}

	err = w.G.SetKeybinding("",
		gocui.KeyEnter,
		gocui.ModNone,
		w.btnclik)
	if err != nil {
		return err
	}

	return nil
}

func (w *Window) btnclik(g1 *gocui.Gui, v1 *gocui.View) error {

	w.G.Update(func(g2 *gocui.Gui) error {

		edit, _ := g2.View(edit)

		if len(edit.BufferLines()) == 0 {
			edit.Title = "Password cannot be empty "
			return nil
		}

		if len(edit.BufferLines()) > 1 {
			edit.Title = "Password should be one line, we cleared the input field, re-enter one line"
			edit.Clear()
			edit.SetCursor(0, 0)
			return nil
		}

		pass := edit.BufferLines()[0]
		edit.Title = "Starting up"
		go w.testPass(pass)

		return nil
	})

	return nil

}

func (w *Window) testPass(p string) {
	if w.s.TestPass(p) {
		//init app
		if err := app.InitForOldUser(w.s); err != nil {
			panic(err)
		}
		//show main window
		w.G.Update(func(g2 *gocui.Gui) error {
			w2 := mainwindow.Window{G: w.G}
			w2.Init()
			return nil
		})

	} else {
		w.G.Update(func(g2 *gocui.Gui) error {
			edit, _ := g2.View(edit)
			edit.Title = "Password does not match, try again"
			return nil
		})
	}
}

func (w *Window) print1() {

	v, _ := w.G.View(edit)
	v.Editable = true
	v.Mask = '*'
	v.Title = fmt.Sprintf("Enter password")
	v.SetCursor(0, 0)
	w.G.SetCurrentView(edit)

	v, _ = w.G.View(help)
	v.Title = "Zvart " + app.Version
	v.Clear()
	v.Wrap = true
	fmt.Fprintf(v, " \n")
	fp, _ := filepath.Abs(flags.StartFlags.DB)
	fmt.Fprintf(v, " Open file \"%s\" \n",
		fp)

	fmt.Fprintf(v, strings.Repeat(" \n", 8))
	fmt.Fprintf(v, " Ctrl + Q - exit\n")
	fmt.Fprintf(v, " Ctrl + X - clear input box \n")
	fmt.Fprintf(v, " Ctrl + V - clear input box and paste from clipboard\n")
	fmt.Fprintf(v, " Ctrl + Z - show/hide password\n")
	fmt.Fprintf(v, " Enter - apply password\n")
	fmt.Fprintf(v, " Run the program with the flag '-db file_name' to select a different account file\n")
}
