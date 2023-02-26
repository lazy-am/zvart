package cui

import (
	"github.com/jroimartin/gocui"
	"github.com/lazy-am/zvart/internal/ui/gocui/newuser"
	"github.com/lazy-am/zvart/internal/ui/gocui/openuser"
)

type consoleUI struct {
	CUI *gocui.Gui
}

func InitGOCUI() (*consoleUI, error) {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return nil, err
	}
	return &consoleUI{g}, nil
}

func (ui *consoleUI) Close() {
	ui.CUI.Close()
}

func (ui *consoleUI) InitOpenDBWindow() {
	w := openuser.Window{G: ui.CUI}
	if err := w.Init(); err != nil {
		panic(err)
	}
}

func (ui *consoleUI) InitNewUserWindow() {
	w := newuser.Window{G: ui.CUI}
	if err := w.Init(); err != nil {
		panic(err)
	}
}

func (ui *consoleUI) MainLoop() error {
	if err := ui.CUI.MainLoop(); err != nil && err != gocui.ErrQuit {
		return err
	}
	return nil
}
