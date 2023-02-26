package starter

import (
	"github.com/lazy-am/zvart/internal/core/starter/flags"
	"github.com/lazy-am/zvart/internal/ui"
	"github.com/lazy-am/zvart/pkg/file"
	"github.com/lazy-am/zvart/pkg/service"
)

func StartConsole() {
	flags.ParseFlags()
	defer service.ClosingServices()
	i, err := ui.CreateConsoleUI()
	if err != nil {
		panic(err)
	}
	if file.FileExists(flags.StartFlags.DB) {
		i.InitOpenDBWindow()
	} else {
		i.InitNewUserWindow()
	}
	err = i.MainLoop()
	if err != nil {
		panic(err)
	}
}
