package tmes

import (
	"encoding/binary"
	"encoding/json"
	"time"

	"github.com/lazy-am/zvart/pkg/cipher"
)

type messageStorage interface {
	Set(table []byte, subtable []byte, key []byte, value []byte) error
	Get(table []byte, subtable []byte, key []byte) ([]byte, error)
	GetNextId(table []byte, subtable []byte) ([]byte, error)
	CreateTable(table, subtable []byte) error
	LoadList(table []byte, subtable []byte) (map[uint64][]byte, error)
	LoadListFromId(table []byte, subtable []byte, id []byte) (map[uint64][]byte, error)
	LoadLast(table []byte, subtable []byte, count uint64) (lst map[uint64][]byte, startIndex uint64, er error)
	GetMaxId(table []byte, subtable []byte) (uint64, error)
}

type TextMessage struct {
	dbKey        []byte
	DbTable      []byte
	IDGlobal     []byte //sha hash of start date
	CreatedByMe  bool
	IsSended     bool
	CreationTime time.Time
	SendedTime   time.Time
	Text         string
}

const (
	messTableName = "messages"
)

func new(db messageStorage, table []byte) (*TextMessage, error) {
	//request a new id for the table with messages
	messID, err := db.GetNextId([]byte(messTableName), table)
	if err != nil {
		return nil, err
	}

	m := TextMessage{dbKey: messID,
		CreationTime: time.Now(),
		DbTable:      table}
	return &m, nil
}

func Accepted(db messageStorage,
	table []byte,
	idglobal []byte,
	creationTime time.Time,
	text string) (*TextMessage, error) {

	m, err := new(db, table)
	if err != nil {
		return nil, err
	}
	m.CreationTime = creationTime
	m.Text = text
	m.IDGlobal = idglobal
	m.SendedTime = time.Now()
	m.IsSended = true

	return m, m.Save(db)
}

// Create a user message to send to a contact
func Create(db messageStorage, table []byte, text string) (*TextMessage, error) {
	m, err := new(db, table)
	if err != nil {
		return nil, err
	}

	m.Text = text
	b := append([]byte(text), m.dbKey...)
	m.IDGlobal = cipher.GetSHA256(b)

	m.CreatedByMe = true
	return m, m.Save(db)
}

func (m *TextMessage) GetDBkey() []byte {
	return m.dbKey
}

func Load(db messageStorage, table []byte, id uint64) (*TextMessage, error) {
	var key []byte = make([]byte, 8)
	binary.LittleEndian.PutUint64(key, id)
	data, err := db.Get([]byte(messTableName), table, key)
	if err != nil {
		return nil, err
	}
	m := TextMessage{dbKey: key}
	if err = json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func LoadList(db messageStorage, table []byte) ([]*TextMessage, error) {

	list, err := db.LoadList([]byte(messTableName), table)
	if err != nil {
		return nil, err
	}
	ml := []*TextMessage{}
	for k, bin := range list {
		var buf []byte = make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, k)
		m := TextMessage{dbKey: buf}
		if err := json.Unmarshal(bin, &m); err != nil {
			return nil, err
		}
		ml = append(ml, &m)
	}
	return ml, nil

}

func LoadLast(db messageStorage, table []byte, count uint64) ([]*TextMessage, error) {
	list, startIndex, err := db.LoadLast([]byte(messTableName), table, count)
	if err != nil {
		return nil, err
	}
	ml := []*TextMessage{}
	for i := startIndex; i < (uint64(len(list)) + startIndex); i++ {
		var buf []byte = make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, i)
		m := TextMessage{dbKey: buf}
		if err := json.Unmarshal(list[i], &m); err != nil {
			return nil, err
		}
		ml = append(ml, &m)
	}
	return ml, nil
}

func LoadListFromId(db messageStorage, table []byte, id []byte) ([]*TextMessage, error) {
	list, err := db.LoadListFromId([]byte(messTableName), table, id)
	if err != nil {
		return nil, err
	}
	ml := []*TextMessage{}
	for k, bin := range list {
		var buf []byte = make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, k)
		m := TextMessage{dbKey: buf}
		if err := json.Unmarshal(bin, &m); err != nil {
			return nil, err
		}
		ml = append(ml, &m)
	}
	return ml, nil
}

func GetMaxId(db messageStorage, table []byte) (uint64, error) {
	return db.GetMaxId([]byte(messTableName), table)
}

func (m *TextMessage) Save(db messageStorage) error {
	j, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if err = db.Set([]byte(messTableName), m.DbTable, m.dbKey, j); err != nil {
		return err
	}
	return nil
}
