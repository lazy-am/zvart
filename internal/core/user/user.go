package user

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
)

const (
	NameMaxLen = 32
	MinPassLen = 1
	tableDb    = "user" //database table name
	nameInDB   = "base" //the key where saved the name and ed25519 keys
)

type userStorage interface {
	Set(table []byte, subtable []byte, key []byte, value []byte) error
	Get(table []byte, subtable []byte, key []byte) ([]byte, error)
	LoadList(table []byte, subtable []byte) (map[uint64][]byte, error)
	LoadListFromId(table []byte, subtable []byte, id []byte) (map[uint64][]byte, error)
}

type User struct {
	Name    string
	PrivKey *rsa.PrivateKey
}

func Create(name string, db userStorage) (*User, error) {
	u := User{Name: name}
	var err error
	u.PrivKey, err = rsa.GenerateKey(rand.Reader, 3072)
	if err != nil {
		return nil, err
	}
	u.Save(db)
	return &u, nil
}

// first load user name, and ed25519 keys from storage
func Load(db userStorage) (*User, error) {
	u := User{}
	b, err := db.Get([]byte(tableDb), nil, []byte(nameInDB))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

// save user name, private and public key
func (u *User) Save(db userStorage) error {
	j, err := json.Marshal(u)
	if err != nil {
		return err
	}
	if err = db.Set([]byte(tableDb), nil, []byte(nameInDB), j); err != nil {
		return err
	}
	return nil
}
