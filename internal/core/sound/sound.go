package sound

import (
	"bytes"
	"io"
	"os"
	"time"

	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto/v2"
	"github.com/lazy-am/zvart/pkg/service"
)

type AppSound struct {
	noError     bool
	soundCtx    *oto.Context
	sound1      oto.Player
	sound1Ready bool
	sound2      oto.Player
	sound2Ready bool
	readyChan   chan struct{}
}

func Init(s1, s2 string) (*AppSound, error) {

	s := AppSound{noError: false}

	samplingRate := 44100
	numOfChannels := 2
	audioBitDepth := 2
	var err error
	s.soundCtx, s.readyChan, err = oto.NewContext(samplingRate, numOfChannels, audioBitDepth)
	if err != nil {
		return &s, err
	}

	// "./sounds/notification-sound-7062.mp3"
	fileBytes, err := os.ReadFile(s1)
	if err != nil {
		return &s, err
	}
	fileBytesReader := bytes.NewReader(fileBytes)
	decodedMp3, err := mp3.NewDecoder(fileBytesReader)
	if err != nil {
		return &s, err
	}
	s.sound1 = s.soundCtx.NewPlayer(decodedMp3)
	s.sound1Ready = true

	//sound 2
	fileBytes, err = os.ReadFile(s2)
	if err != nil {
		return &s, err
	}
	fileBytesReader = bytes.NewReader(fileBytes)
	decodedMp3, err = mp3.NewDecoder(fileBytesReader)
	if err != nil {
		return &s, err
	}
	s.sound2 = s.soundCtx.NewPlayer(decodedMp3)
	s.sound2Ready = true

	service.AddService(&s)

	s.noError = true
	return &s, nil
}

func (s *AppSound) PlaySound2() {
	go s.playSound2()
}

func (s *AppSound) playSound2() {
	if !s.noError || !s.sound2Ready {
		return
	}
	<-s.readyChan
	s.sound2Ready = false
	s.sound2.Play()
	for s.sound2.IsPlaying() {
		time.Sleep(time.Millisecond * 100)
	}
	_, err := s.sound2.(io.Seeker).Seek(0, io.SeekStart)
	if err != nil {
		s.noError = false
		return
	}
	s.sound2Ready = true
}

func (s *AppSound) PlaySound1() {
	go s.playSound1()
}

func (s *AppSound) playSound1() {
	if !s.noError || !s.sound1Ready {
		return
	}
	<-s.readyChan
	s.sound1Ready = false
	s.sound1.Play()
	for s.sound1.IsPlaying() {
		time.Sleep(time.Millisecond * 100)
	}
	_, err := s.sound1.(io.Seeker).Seek(0, io.SeekStart)
	if err != nil {
		s.noError = false
		return
	}
	s.sound1Ready = true
}

func (s *AppSound) Close() {
	if !s.noError {
		return
	}
	s.sound1.Close()
}
