package storage

import (
	"github.com/lazy-am/zvart/internal/storage/boltdb"
	"github.com/lazy-am/zvart/pkg/service"
)

type Storage interface {
	//tries to decrypt the test string with the password
	//if successful, returns true and saves it for later use
	TestPass(string) bool
	//Set password without verification
	SetPass(string)
	GetError() error
	//Write the value in encrypted form
	Set(table []byte, subtable []byte, key []byte, value []byte) error
	//Eead and decipher the value
	Get(table []byte, subtable []byte, key []byte) ([]byte, error)
	GetNextId(table []byte, subtable []byte) ([]byte, error)
	CreateTable(table, subtable []byte) error
	LoadList(table []byte, subtable []byte) (map[uint64][]byte, error)
	LoadListFromId(table []byte, subtable []byte, id []byte) (map[uint64][]byte, error)
	LoadLast(table []byte, subtable []byte, count uint64) (map[uint64][]byte, error)
	GetMaxId(table []byte, subtable []byte) (uint64, error)
}

func InitFile(dbname string) (Storage, error) {
	db, err := boltdb.InitBoltFile(dbname)
	if err != nil {
		return nil, err
	}
	service.AddService(db)
	return db, err
}
