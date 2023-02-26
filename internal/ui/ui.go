package ui

import (
	cui "github.com/lazy-am/zvart/internal/ui/gocui"
	"github.com/lazy-am/zvart/pkg/service"
)

type userInterface interface {
	InitOpenDBWindow()
	InitNewUserWindow()
	MainLoop() error
}

func CreateConsoleUI() (userInterface, error) {
	i, err := cui.InitGOCUI()
	if err != nil {
		return nil, err
	}
	service.AddService(i)
	return i, nil
}
