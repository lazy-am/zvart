package torl

import (
	"context"
	"crypto/ed25519"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"

	"github.com/cretz/bine/tor"
	"github.com/lazy-am/zvart/pkg/cipher"
	"github.com/lazy-am/zvart/pkg/service"
)

type torlancherStorage interface {
	Set(table []byte, subtable []byte, key []byte, value []byte) error
	Get(table []byte, subtable []byte, key []byte) ([]byte, error)
}

const (
	bdtableName     = "user"
	bdonionKeyName  = "onionKey"
	bdonionLinkName = "onionLink"
)

type Torlancher struct {
	tor            *tor.Tor
	ctx            context.Context
	cancelFunc     context.CancelFunc
	creationError  bool
	runtimeError   bool
	Connected      bool
	onionError     bool
	OnionConnected bool

	onion    *tor.OnionService
	listener net.Listener
	keyStor  torlancherStorage
}

func Init(ctrlport int16,
	torexe string,
	portonion int16,
	ks torlancherStorage,
	torrc string,
	workdir string,
	listener net.Listener,
) (*Torlancher, error) {

	t := &Torlancher{}
	t.listener = listener
	t.keyStor = ks
	t.ctx, t.cancelFunc = context.WithCancel(context.Background())

	var err error
	t.tor, err = tor.Start(t.ctx, &tor.StartConf{DataDir: workdir,
		ExePath:         torexe,
		GeoIPFileReader: geoipFileReader,
		TorrcFile:       torrc,
		ControlPort:     int(ctrlport),
		NoAutoSocksPort: true,
		EnableNetwork:   false})
	if err != nil {
		t.creationError = true
		return nil, err
	}
	go waitTorConnect(t, portonion)
	service.AddService(t)
	return t, nil
}

func (t *Torlancher) GetHost() string {
	if t.OnionConnected {
		return t.onion.ID
	}
	return "not connected"
}

func (t *Torlancher) Close() {
	if t.onion != nil {
		t.onion.Close()
	}
	t.tor.Close()
	t.cancelFunc()
}

func (t *Torlancher) GetError() error {
	er := ""
	if t.creationError {
		er += " torCreationError "
	}
	if t.runtimeError {
		er += " torError "
	}
	if t.onionError {
		er += " onionError "
	}
	if er != "" {
		return errors.New(er)
	}
	return nil
}

func (t *Torlancher) startOnion(local_port int16) (*tor.OnionService, error) {

	k, _ := t.readOnionKey()
	if k == nil {
		var err error
		k, err = cipher.GeneratePrivEd25519()
		if err != nil {
			return nil, err
		}
		t.saveOnionKey(k)
	}

	onion, err := t.tor.Listen(t.ctx,
		&tor.ListenConf{RemotePorts: []int{80},
			LocalPort:     int(local_port),
			Key:           k,
			NoWait:        false,
			LocalListener: t.listener})

	if err != nil {
		return nil, err
	}

	return onion, nil

}

func (t *Torlancher) readOnionKey() (ed25519.PrivateKey, error) {
	return t.keyStor.Get([]byte(bdtableName), nil, []byte(bdonionKeyName))
}

func (t *Torlancher) saveOnionKey(key ed25519.PrivateKey) error {
	return t.keyStor.Set([]byte(bdtableName), nil, []byte(bdonionKeyName), key)
}

func geoipFileReader(ipv6 bool) (io.ReadCloser, error) {

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	p := filepath.Join(wd, "tor", "data", "geoip")
	path, err := filepath.Abs(p)
	if err != nil {
		return nil, err
	}
	if ipv6 {
		path += "6"
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func waitTorConnect(t *Torlancher, onion_port int16) error {
	//wait tor connected
	err := t.tor.EnableNetwork(t.ctx, true)
	if err != nil {
		t.runtimeError = true
		return err
	}
	t.Connected = true

	//init onion
	t.onion, err = t.startOnion(onion_port)
	if err != nil {
		t.onionError = true
		return err
	}
	t.OnionConnected = true
	t.saveSelfLink(t.onion.ID)
	return nil
}

func (t *Torlancher) saveSelfLink(link string) error {
	return t.keyStor.Set([]byte(bdtableName), nil, []byte(bdonionLinkName), []byte(link))
}

func GetSelfLink(ks torlancherStorage) ([]byte, error) {
	return ks.Get([]byte(bdtableName), nil, []byte(bdonionLinkName))
}
