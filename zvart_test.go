package app_test

import (
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/lazy-am/zvart/internal/core/contact"
	"github.com/lazy-am/zvart/internal/core/server"
	"github.com/lazy-am/zvart/internal/core/tmes"
	"github.com/lazy-am/zvart/internal/core/user"
	"github.com/lazy-am/zvart/internal/storage"
	"github.com/lazy-am/zvart/internal/torl"
	"github.com/lazy-am/zvart/pkg/service"
)

func TestZvart(t *testing.T) {

	defer service.ClosingServices()

	os.Remove("test_storage1.zvart")
	os.Remove("test_storage2.zvart")

	const (
		onionPort1   = 6060
		controlPort1 = 6061
		socksPort1   = 6062

		onionPort2   = 6063
		controlPort2 = 6064
		socksPort2   = 6065
	)

	torrc1 := filepath.Join("tor", "torrcTest1")
	workDir1 := "torworkdirTest1"

	torrc2 := filepath.Join("tor", "torrcTest2")
	workDir2 := "torworkdirTest2"

	torExe := "tor"
	if runtime.GOOS == "windows" {
		torExe = filepath.Join("tor", "tor", "tor.exe")
	} else {
		torExe = filepath.Join("tor", "tor", "tor")
		abs, err := filepath.Abs(filepath.Join("tor", "tor"))
		if err != nil {
			t.Fatal("it was not possible to calculate the absolute path of the folder with shared libraries")
		}
		os.Setenv("LD_LIBRARY_PATH", abs)
	}

	var (
		storage1 storage.Storage
		storage2 storage.Storage
		tor1     *torl.Torlancher
		tor2     *torl.Torlancher
		contact1 *contact.Contact
		contact2 *contact.Contact
		message2 *tmes.TextMessage
	)
	t.Run("init storage", func(t *testing.T) {
		var err error
		storage1, err = storage.InitFile("test_storage1.zvart")
		if err != nil {
			t.Fatal("storage1 error " + err.Error())
		}
		storage2, err = storage.InitFile("test_storage2.zvart")
		if err != nil {
			t.Fatal("storage2 error " + err.Error())
		}
		storage1.SetPass("pass1")
		storage2.SetPass("pass2")
	})
	t.Run("create users", func(t *testing.T) {
		_, err := user.Create("user1", storage1)
		if err != nil {
			t.Fatal("user1 creation error " + err.Error())
		}
		_, err = user.Create("user2", storage2)
		if err != nil {
			t.Fatal("user2 creation error " + err.Error())
		}
	})
	t.Run("init servers", func(t *testing.T) {
		//# 1
		listener1, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(onionPort1))
		if err != nil {
			t.Fatal("listener1 creation error " + err.Error())
		}
		tor1, err = torl.Init(int16(controlPort1), torExe,
			int16(onionPort1),
			storage1,
			torrc1,
			workDir1,
			listener1)
		if err != nil {
			t.Fatal("tor1 creation error " + err.Error())
		}
		_, err = server.Init(listener1, storage1, socksPort1, nil)
		if err != nil {
			t.Fatal("server1 creation error " + err.Error())
		}

		//# 2
		listener2, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(onionPort2))
		if err != nil {
			t.Fatal("listener2 creation error " + err.Error())
		}
		tor2, err = torl.Init(int16(controlPort2), torExe,
			int16(onionPort2),
			storage2,
			torrc2,
			workDir2,
			listener2)
		if err != nil {
			t.Fatal("tor2 creation error " + err.Error())
		}
		_, err = server.Init(listener2, storage2, socksPort2, nil)
		if err != nil {
			t.Fatal("server2 creation error " + err.Error())
		}
	})

	t.Run("wait tor", func(t *testing.T) {
		for !tor1.OnionConnected || !tor2.OnionConnected {
			time.Sleep(3 * time.Second)
		}

		time.Sleep(5 * time.Second)
	})

	t.Run("create contact from user1 to user2", func(t *testing.T) {
		link, err := torl.GetSelfLink(storage2)
		if err != nil {
			t.Fatal("link request error on user 2 " + err.Error())
		}
		contact1, err = contact.NewFromLink(string(link), "hi", storage1)
		if err != nil {
			t.Fatal("contact1 creation error " + err.Error())
		}
	})

	t.Run("send public key user1 to user2", func(t *testing.T) {
		for !contact1.PubKeySended {
			var err error
			contact1, err = contact.Load(storage1, 1)
			if err != nil {
				t.Fatal("contact1 reload error " + err.Error())
			}
			time.Sleep(3 * time.Second)
		}
	})

	t.Run("allow contact from user2 side", func(t *testing.T) {
		cl, err := contact.LoadList(storage2)
		if err != nil {
			t.Fatal("contact2 load list error " + err.Error())
		}
		if cl == nil || len(cl) < 1 {
			t.Fatal("contact2 not exist error " + err.Error())
		}
		//send answer message
		contact2 = cl[0]
		message2, err = tmes.Create(storage2, contact2.DbMessagesTableName, "hi hi")
		if err != nil {
			t.Fatal("contact2 message creation error " + err.Error())
		}
		contact2.FirstUnsentMessageId = message2.GetDBkey()
		contact2.Save(storage2)
	})

	t.Run("wait message on user1 side", func(t *testing.T) {
		for !message2.IsSended {
			var err error
			message2, err = tmes.Load(storage2, contact2.DbMessagesTableName, 1)
			if err != nil {
				t.Fatal("message2 reload error " + err.Error())
			}
			time.Sleep(3 * time.Second)
		}
		time.Sleep(5 * time.Second)
	})

	t.Run("check user2 send pub key", func(t *testing.T) {
		var err error
		contact2, err = contact.Load(storage2, 1)
		if err != nil {
			t.Fatal("contact2 reload error " + err.Error())
		}
		if !contact2.PubKeySended {
			t.Fatal("contact2 not sent key error ")
		}
		if contact2.FirstUnsentMessageId != nil {
			t.Fatal("contact2 FirstUnsentMessageId is not nil error ")
		}
	})

	t.Run("check user1 message recived", func(t *testing.T) {
		ml, err := tmes.LoadList(storage1, contact1.DbMessagesTableName)
		if err != nil {
			t.Fatal("load message list user 1 error " + err.Error())
		}
		if len(ml) != 1 {
			t.Fatal("did not receive a message")
		}
	})

	t.Run("user2 send second message", func(t *testing.T) {
		var err error
		contact2, err = contact.Load(storage2, 1)
		if err != nil {
			t.Fatal("contact2 reload error " + err.Error())
		}
		message2, err = tmes.Create(storage2, contact2.DbMessagesTableName, "hi hi 2")
		if err != nil {
			t.Fatal("contact2 message creation error " + err.Error())
		}
		contact2.FirstUnsentMessageId = message2.GetDBkey()
		contact2.Save(storage2)
	})

	t.Run("user2 wait message sended", func(t *testing.T) {
		for !message2.IsSended {
			var err error
			message2, err = tmes.Load(storage2, contact2.DbMessagesTableName, 2)
			if err != nil {
				t.Fatal("message2 reload error " + err.Error())
			}
			time.Sleep(3 * time.Second)
		}
		time.Sleep(5 * time.Second)
	})

	t.Run("user2 check message list", func(t *testing.T) {
		var err error
		contact2, err = contact.Load(storage2, 1)
		if err != nil {
			t.Fatal("contact2 reload error " + err.Error())
		}
		ml, err := tmes.LoadList(storage2, contact2.DbMessagesTableName)
		if err != nil {
			t.Fatal("load message list user2 error " + err.Error())
		}
		if len(ml) != 2 {
			t.Fatal("message list length is not 2")
		}
	})

	t.Run("user1 check message list", func(t *testing.T) {
		var err error
		contact1, err = contact.Load(storage1, 1)
		if err != nil {
			t.Fatal("contact1 reload error " + err.Error())
		}
		ml, err := tmes.LoadList(storage1, contact1.DbMessagesTableName)
		if err != nil {
			t.Fatal("load message list user1 error " + err.Error())
		}
		if len(ml) != 2 {
			t.Fatal("message list length is not 2")
		}
	})
}
