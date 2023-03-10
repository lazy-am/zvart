package contact

import (
	"bytes"
	"crypto/rsa"
	"encoding/binary"
	"encoding/json"
	"errors"
	"time"

	"github.com/lazy-am/zvart/pkg/cipher"
)

type contactStorage interface {
	Set(table []byte, subtable []byte, key []byte, value []byte) error
	Get(table []byte, subtable []byte, key []byte) ([]byte, error)
	GetNextId(table []byte, subtable []byte) ([]byte, error)
	CreateTable(table, subtable []byte) error
	ClearTable(table, subtable []byte) error
	LoadList(table []byte, subtable []byte) (map[uint64][]byte, error)
	LoadListFromId(table []byte, subtable []byte, id []byte) (map[uint64][]byte, error)
}

const (
	contactsTableName = "contacts"
	messTableName     = "messages"
	secSessionKeyLife = 3600
)

type Contact struct {
	//the key in the table where this contact is stored
	dbKey []byte
	//is created when a contact is created from the link,
	//it is needed for verification at the first reply message
	SecretPass []byte
	SessionKey []byte
	SesKeyTime time.Time
	//the app already works with this contact
	ServerWork bool
	//
	RemoteId []byte
	//
	OnionID []byte
	// the number of the table storing all messages of this contact
	DbMessagesTableName []byte
	//Name reported by the contact himself
	ReportedName string
	//
	PubKey *rsa.PublicKey
	//For outgoing contacts - a hello message
	HelloMessage string
	//If true, the connection has already been established and our public key has been sent
	PubKeySended bool
	//
	CreatedByMe bool
	//Time of last established connection
	LastCallTime time.Time
	//Last connection attempt time
	LastTryTime time.Time
	//Contact creation time
	CreationTime time.Time
	//If any messages are unsent, the variable stores the index of the first one
	FirstUnsentMessageId []byte
	LastViewedMessageId  []byte
	NeedUpdateGuiInfo    bool
}

func new(db contactStorage) (*Contact, error) {

	//request a new id for the contact table
	contactTableID, err := db.GetNextId([]byte(contactsTableName), nil)
	if err != nil {
		return nil, err
	}
	//request a new id for the table with messages
	messageTableName, err := db.GetNextId([]byte(messTableName), nil)
	if err != nil {
		return nil, err
	}
	//create a table
	if err := db.CreateTable([]byte(messTableName), messageTableName); err != nil {
		return nil, err
	}
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], 0)
	c := Contact{dbKey: contactTableID,
		OnionID:             []byte{},
		ReportedName:        "",
		HelloMessage:        "",
		DbMessagesTableName: messageTableName,
		CreationTime:        time.Now(),
		LastViewedMessageId: buf[:],
	}
	return &c, nil
}

func SetSessionKey(db contactStorage, remoteId []byte, pass []byte, sesKey []byte) error {

	//
	c, err := Load(db, binary.LittleEndian.Uint64(remoteId))
	if err != nil {
		return err
	}

	if !bytes.Equal(c.SecretPass, pass) {
		return errors.New("the passwords didn't match")
	}

	c.SessionKey = sesKey
	c.SesKeyTime = time.Now()
	return c.Save(db)
}

func AddPubKey(db contactStorage,
	name string,
	remoteId []byte,
	pass []byte,
	key *rsa.PublicKey) error {

	c, err := Load(db, binary.LittleEndian.Uint64(remoteId))
	if err != nil {
		return err
	}

	if c.PubKey != nil {
		return errors.New("key already exists")
	}

	if !bytes.Equal(c.SecretPass, pass) {
		return errors.New("the passwords didn't match")
	}

	c.PubKey = key
	c.ReportedName = name
	return c.Save(db)

}

func NewAccepted(db contactStorage,
	onionID []byte,
	pubKey *rsa.PublicKey,
	hello string,
	name string,
	pass []byte,
	rid []byte) (*Contact, error) {

	// the system bottleneck is easy enough to flood the program with requests
	// and its table of contacts will take up all the memory, and the decryption / encryption - CPU resource
	// is necessary to protect this place from the possibility of a simple attack
	cl, err := LoadList(db)
	if err != nil {
		return nil, err
	}

	for _, c := range cl {
		if bytes.Equal(c.OnionID, onionID) || ((c.PubKey != nil) && c.PubKey.Equal(pubKey)) {
			return nil, errors.New("contact already exists")
		}
	}

	c, err := new(db)
	if err != nil {
		return nil, err
	}
	c.HelloMessage = hello
	c.OnionID = onionID
	c.PubKey = pubKey
	c.ReportedName = name
	c.SecretPass = pass
	c.RemoteId = rid
	c.LastCallTime = time.Now()

	return c, c.Save(db)

}

func NewFromLink(id, himessage string, db contactStorage) (*Contact, error) {

	//check if such a contact exists in the database
	cl, err := LoadList(db)
	if err != nil {
		return nil, err
	}
	for _, c := range cl {
		if bytes.Equal(c.OnionID, []byte(id)) {
			return nil, errors.New("a contact with this ID already exists")
		}
	}

	c, err := new(db)
	if err != nil {
		return nil, err
	}
	c.OnionID = []byte(id)
	c.HelloMessage = himessage
	c.CreatedByMe = true
	b := binary.LittleEndian.AppendUint64(append([]byte(himessage), c.OnionID...),
		uint64(time.Now().Unix()))
	c.SecretPass = cipher.GetSHA256(b)

	return c, c.Save(db)
}

func Load(db contactStorage, index uint64) (*Contact, error) {
	var key []byte = make([]byte, 8)
	binary.LittleEndian.PutUint64(key, index)
	data, err := db.Get([]byte(contactsTableName), nil, key)
	if err != nil {
		return nil, err
	}
	c := Contact{dbKey: key}
	if err = json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// Called when the application starts and load contacts list from the database
func LoadList(db contactStorage) ([]*Contact, error) {
	list, err := db.LoadList([]byte(contactsTableName), nil)
	if err != nil {
		return nil, err
	}
	cl := []*Contact{}
	for i := 1; i <= len(list); i++ {
		var buf []byte = make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, uint64(i))
		c := Contact{dbKey: buf}
		if err := json.Unmarshal(list[uint64(i)], &c); err != nil {
			return nil, err
		}
		cl = append(cl, &c)
	}
	return cl, nil
}

func (c *Contact) CheckSesKey(db contactStorage) (genereted bool) {
	if c.SessionKey == nil || (int(time.Since(c.SesKeyTime).Seconds()) > secSessionKeyLife) {
		c.SessionKey = cipher.GenerateSesKey()
		c.SesKeyTime = time.Now()
		c.Save(db)
		return true
	}
	return false
}

func (c *Contact) Save(db contactStorage) error {
	j, err := json.Marshal(c)
	if err != nil {
		return err
	}
	if err = db.Set([]byte(contactsTableName), nil, []byte(c.dbKey), j); err != nil {
		return err
	}
	return nil
}

func (c *Contact) GetDBkey() []byte {
	return c.dbKey
}

func (c *Contact) Equal(cont *Contact) bool {
	if cont == nil {
		return false
	}
	if bytes.Equal(c.dbKey, cont.dbKey) && bytes.Equal(c.OnionID, cont.OnionID) {
		return true
	}
	return false
}

func (c *Contact) ClearMessages(db contactStorage, outerror chan error) error {
	go func() {
		for c.ServerWork {
			var err error
			time.Sleep(time.Second)
			c, err = Load(db, binary.LittleEndian.Uint64(c.dbKey))
			if err != nil {
				outerror <- err
				return
			}
		}
		if err := db.ClearTable([]byte(messTableName), c.DbMessagesTableName); err != nil {
			outerror <- err
			return
		}
		c.NeedUpdateGuiInfo = true
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, 0)
		c.FirstUnsentMessageId = b
		c.LastViewedMessageId = b
		if err := c.Save(db); err != nil {
			outerror <- err
		}
	}()
	return nil
}
