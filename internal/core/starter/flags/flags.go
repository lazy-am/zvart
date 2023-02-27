package flags

import (
	"flag"
	"path/filepath"
	"runtime"
)

type flags struct {
	Tor         string
	LibPath     string
	UseLibPath  bool
	DB          string
	ControlPort int
	OnionPort   int
	SocksPort   int // this parameter is specified in the torrc file, it cannot be overridden in the program yet
	Torrc       string
	TorWorkDir  string
}

var StartFlags flags
var DefFlags flags

// Preparing the configuration passed in the command line parameters
func ParseFlags() {

	var torD string
	if runtime.GOOS == "windows" {
		torD = filepath.Join("tor", "tor", "tor.exe")
	} else {
		torD = filepath.Join("tor", "tor", "tor")
	}

	DefFlags = flags{Tor: torD,
		DB:          "user.zvart",
		ControlPort: 6061,
		OnionPort:   6062,
		SocksPort:   6060,
		Torrc:       filepath.Join("tor", "torrc"),
		TorWorkDir:  "torworkdir",
		LibPath:     filepath.Join("tor", "tor"),
		UseLibPath:  true}

	tor := flag.String("tor", DefFlags.Tor, "location of your tor executable file")
	libp := flag.String("lib", DefFlags.LibPath, "shared libraries folder for tor(linux only)")
	libuse := flag.Bool("libuse", DefFlags.UseLibPath, "using of shared libraries folder for tor(linux only)")
	db := flag.String("db", DefFlags.DB, "location of your account database file")
	torcp := flag.Int("tp", DefFlags.ControlPort, "tor control port")
	torop := flag.Int("op", DefFlags.OnionPort, "onion service port")
	torpp := flag.Int("pp", DefFlags.SocksPort, "tor socks port(the same value should be in the torrc file)")
	torc := flag.String("tc", DefFlags.Torrc, "tor configuration file")
	twd := flag.String("twd", DefFlags.TorWorkDir, "tor work directory")

	flag.Parse()

	StartFlags = flags{Tor: *tor,
		DB:          *db,
		ControlPort: *torcp,
		OnionPort:   *torop,
		SocksPort:   *torpp,
		Torrc:       *torc,
		TorWorkDir:  *twd,
		LibPath:     *libp,
		UseLibPath:  *libuse}

}
