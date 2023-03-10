package app

import (
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/lazy-am/zvart/internal/core/server"
	"github.com/lazy-am/zvart/internal/core/sound"
	"github.com/lazy-am/zvart/internal/core/starter/flags"
	"github.com/lazy-am/zvart/internal/core/user"
	"github.com/lazy-am/zvart/internal/storage"
	"github.com/lazy-am/zvart/internal/torl"
	"golang.design/x/clipboard"
)

func init() {
	Zvart = app{}
	Zvart.Clipboard = true
	err := clipboard.Init()
	if err != nil {
		Zvart.Clipboard = false
	}
	Zvart.ErrorNotice = make(chan error, 1)
}

// Initialize storage at the beginning of the database opening dialog
func InitStorage() (storage.Storage, error) {
	return storage.InitFile(flags.StartFlags.DB)
}

func InitForOldUser(s storage.Storage) error {
	Zvart.Db = s
	return commonInit()
}

// Initializing the application when a new user is created
func InitForNewUser(name, pass string) error {

	// init storage
	s, err := storage.InitFile(flags.StartFlags.DB)
	if err != nil {
		return err
	}
	Zvart.Db = s
	Zvart.Db.SetPass(pass)

	// save user
	_, err = user.Create(name, Zvart.Db)
	if err != nil {
		return err
	}

	return commonInit()
}

func commonInit() error {

	Zvart.Sound, _ = sound.Init(filepath.Join("sounds", "notification-sound-7062.mp3"),
		filepath.Join("sounds", "stop-13692.mp3"))

	listener, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(flags.StartFlags.OnionPort))
	if err != nil {
		return err
	}

	if flags.StartFlags.UseLibPath && runtime.GOOS == "linux" {
		abs, err := filepath.Abs(flags.StartFlags.LibPath)
		if err == nil {
			os.Setenv("LD_LIBRARY_PATH", abs)
		}
	}
	//init tor
	Zvart.Tor, err = torl.Init(int16(flags.StartFlags.ControlPort),
		flags.StartFlags.Tor,
		int16(flags.DefFlags.OnionPort),
		Zvart.Db,
		flags.StartFlags.Torrc,
		flags.StartFlags.TorWorkDir, listener)
	if err != nil {
		return err
	}

	Zvart.Server, err = server.Init(listener, Zvart.Db, uint16(flags.StartFlags.SocksPort), Zvart.Sound)
	if err != nil {
		return err
	}
	return nil
}
