package sound_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto/v2"
	"github.com/lazy-am/zvart/internal/core/sound"
	"github.com/lazy-am/zvart/pkg/service"
)

func TestSound(t *testing.T) {
	// Read the mp3 file into memory

	fileBytes, err := os.ReadFile(filepath.Join("..",
		"..",
		"..",
		"sounds",
		"notification-sound-7062.mp3"))
	if err != nil {
		t.Fatal("reading my-file.mp3 failed: " + err.Error())
	}

	// Convert the pure bytes into a reader object that can be used with the mp3 decoder
	fileBytesReader := bytes.NewReader(fileBytes)

	// Decode file
	decodedMp3, err := mp3.NewDecoder(fileBytesReader)
	if err != nil {
		t.Fatal("mp3.NewDecoder failed: " + err.Error())
	}

	// Prepare an Oto context (this will use your default audio device) that will
	// play all our sounds. Its configuration can't be changed later.

	// Usually 44100 or 48000. Other values might cause distortions in Oto
	samplingRate := 44100

	// Number of channels (aka locations) to play sounds from. Either 1 or 2.
	// 1 is mono sound, and 2 is stereo (most speakers are stereo).
	numOfChannels := 2

	// Bytes used by a channel to represent one sample. Either 1 or 2 (usually 2).
	audioBitDepth := 2

	// Remember that you should **not** create more than one context
	otoCtx, readyChan, err := oto.NewContext(samplingRate, numOfChannels, audioBitDepth)
	if err != nil {
		t.Fatal("oto.NewContext failed: " + err.Error())
	}
	// It might take a bit for the hardware audio devices to be ready, so we wait on the channel.
	<-readyChan

	// Create a new 'player' that will handle our sound. Paused by default.
	player := otoCtx.NewPlayer(decodedMp3)

	// Play starts playing the sound and returns without waiting for it (Play() is async).
	player.Play()

	// We can wait for the sound to finish playing using something like this
	for player.IsPlaying() {
		time.Sleep(time.Millisecond)
	}

	// Now that the sound finished playing, we can restart from the beginning (or go to any location in the sound) using seek
	// newPos, err := player.(io.Seeker).Seek(0, io.SeekStart)
	// if err != nil{
	//     panic("player.Seek failed: " + err.Error())
	// }
	// println("Player is now at position:", newPos)
	// player.Play()

	// If you don't want the player/sound anymore simply close
	err = player.Close()
	if err != nil {
		t.Fatal("player.Close failed: " + err.Error())
	}
}

func TestSoundObjects(t *testing.T) {
	defer service.ClosingServices()

	sounds, err := sound.Init(filepath.Join("..",
		"..",
		"..",
		"sounds",
		"notification-sound-7062.mp3"),
		filepath.Join("..",
			"..",
			"..",
			"sounds",
			"stop-13692.mp3"))
	if err != nil {
		t.Fatal(err)
	}
	sounds.PlaySound1()
	time.Sleep(time.Second)
	sounds.PlaySound2()
	time.Sleep(time.Second)
	sounds.PlaySound1()
	time.Sleep(time.Second)
	sounds.PlaySound2()

}
