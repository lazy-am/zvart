package newuser

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/lazy-am/zvart/internal/core/app"
	"github.com/lazy-am/zvart/internal/core/starter/flags"
	"github.com/lazy-am/zvart/internal/core/user"
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
	step int
	name string
	pass string
}

func (w *Window) Init() error {
	w.step = 1
	w.G.Cursor = true
	w.G.Mouse = true
	w.G.SetManagerFunc(w.layout)
	if err := w.keybindings(); err != nil {
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

	if v, err := g.SetView(btn, maxX-11, 11, maxX-4, 13); err != nil &&
		err == gocui.ErrUnknownView {
		fmt.Fprintf(v, " NEXT")
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

	switch w.step {
	case 1:
		w.btnclik1()
	case 2:
		w.btnclik2()
	case 3:
		w.btnclik3()
	}

	return nil

}

func (w *Window) btnclik1() {

	w.G.Update(func(g2 *gocui.Gui) error {

		edit, _ := g2.View(edit)

		if len(edit.BufferLines()) == 0 {
			edit.Title = "(step 1/3) Nickname cannot be empty "
			return nil
		}

		if len(edit.BufferLines()) > 1 {
			edit.Title = "(step 1/3) Nickname should be one line, we cleared the input field, re-enter one line"
			edit.Clear()
			fmt.Fprintf(edit, "John Doe")
			edit.SetCursor(len("John Doe"), 0)
			return nil
		}

		nickname := edit.BufferLines()[0]
		if len(nickname) > user.NameMaxLen {
			edit.Title = fmt.Sprintf("(step 1/3) Nickname must be less than %d characters long, yours is %d characters long, shorten it", user.NameMaxLen, len(nickname))
			return nil
		}

		w.name = nickname
		// STEP 2 START
		w.step = 2
		w.print2()

		return nil
	})

}

func (w *Window) btnclik2() {

	w.G.Update(func(g2 *gocui.Gui) error {

		edit, _ := g2.View(edit)

		if len(edit.BufferLines()) == 0 {
			edit.Title = "(step 2/3) Password cannot be empty "
			return nil
		}

		if len(edit.BufferLines()) > 1 {
			edit.Title = "(step 2/3) The password cannot be more than one line"
			edit.Clear()
			edit.SetCursor(0, 0)
			return nil
		}

		pass := edit.BufferLines()[0]
		if len(pass) < user.MinPassLen {
			edit.Title = fmt.Sprintf("(step 2/3) Password must be more than %d characters long", user.MinPassLen)
			return nil
		}

		w.pass = pass
		//  STEP 3 start
		w.step = 3
		w.print3()

		return nil
	})

}

func (w *Window) btnclik3() {

	w.G.Update(func(g2 *gocui.Gui) error {

		edit, _ := g2.View(edit)

		if len(edit.BufferLines()) == 0 {
			edit.Title = "(step 3/3) Password cannot be empty "
			return nil
		}

		if len(edit.BufferLines()) > 1 {
			edit.Title = "(step 3/3) The password cannot be more than one line"
			edit.Clear()
			edit.SetCursor(0, 0)
			return nil
		}

		pass := edit.BufferLines()[0]
		if pass != w.pass {
			edit.Title = "(step 3/3) Passwords do not match"
			return nil
		}

		edit.Title = "Okay, it's done, I'm starting up."
		go w.initMain()

		return nil
	})

}

func (w *Window) initMain() {
	//init app
	if err := app.InitForNewUser(w.name, w.pass); err != nil {
		panic(err)
	}
	//show main window
	w2 := mainwindow.Window{G: w.G}
	w.G.Update(func(g2 *gocui.Gui) error {
		w2.Init()
		return nil
	})

}

func (w *Window) print1() {

	v, _ := w.G.View(edit)
	v.Editable = true
	v.Title = fmt.Sprintf("(step 1/3) Type your nickname (one string, less than %d characters)",
		user.NameMaxLen)
	fmt.Fprintf(v, "John Doe")
	v.SetCursor(len("John Doe"), 0)
	w.G.SetCurrentView(edit)

	v, _ = w.G.View(help)
	v.Wrap = true
	v.Title = "Zvart " + app.Version
	v.Clear()
	fmt.Fprintf(v, " \n")
	fmt.Fprintf(v, " New account wizard\n")
	fp, _ := filepath.Abs(flags.StartFlags.DB)
	fmt.Fprintf(v, " Everything will be saved in file \"%s\" \n",
		fp)
	fmt.Fprintf(v, " In the future make a backup of this file to save access to your account, this file is encrypted\n")

	fmt.Fprintf(v, strings.Repeat(" \n", 8))
	fmt.Fprintf(v, " Ctrl + Q - exit\n")
	fmt.Fprintf(v, " Ctrl + X - clear input box \n")
	fmt.Fprintf(v, " Ctrl + V - clear input box and paste from clipboard\n")
	fmt.Fprintf(v, " Enter - next step\n")
	fmt.Fprintf(v, " Run the program with the flag '-db file_name' to select a different account file\n")
}

func (w *Window) print2() {
	edit, _ := w.G.View(edit)
	edit.Clear()
	edit.Title = "(step 2/3) Enter a password to encrypt the file"
	edit.SetCursor(0, 0)
	edit.Mask = '*'

	v, _ := w.G.View(help)
	v.Clear()
	fmt.Fprintf(v, " \n")
	fmt.Fprintf(v, " %s, create a password \n", w.name)
	fmt.Fprintf(v, " Losing your password means losing your account, and access cannot be restored\n")
	fmt.Fprintf(v, " The best way to create and store passwords in a program like KeePass or similar\n")
	fmt.Fprintf(v, strings.Repeat(" \n", 8))
	fmt.Fprintf(v, " Ctrl + Q - exit\n")
	fmt.Fprintf(v, " Ctrl + X - clear input box \n")
	fmt.Fprintf(v, " Ctrl + V - clear input box and paste from clipboard\n")
	fmt.Fprintf(v, " Ctrl + Z - show/hide password\n")
	fmt.Fprintf(v, " Enter - next step\n")
}

func (w *Window) print3() {
	edit, _ := w.G.View(edit)
	edit.Clear()
	edit.Title = "(step 3/3) Enter your password again"
	edit.SetCursor(0, 0)
}
