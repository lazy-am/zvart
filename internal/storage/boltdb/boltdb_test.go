package boltdb

import (
	"os"
	"testing"

	"github.com/lazy-am/zvart/pkg/random"
)

func TestPasswordMatching(t *testing.T) {
	dbname := random.RandStringBytes(10) + ".db"
	db, err := InitBoltFile(dbname)
	if err != nil {
		t.Fatal(dbname + " creation error")
	}
	defer func() {
		db.Close()
		os.Remove(dbname)
	}()

	db.SetPass("pass")
	db.Close()
	db, err = InitBoltFile(dbname)
	if err != nil {
		t.Fatal(dbname + " reopen error")
	}
	if db.TestPass("222") {
		t.Fatal("The wrong password was verified")
	}
	if !db.TestPass("pass") {
		t.Fatal("Password did not pass the test")
	}
}

func TestNonExistKeys(t *testing.T) {
	dbname := random.RandStringBytes(10) + ".db"
	db, err := InitBoltFile(dbname)
	if err != nil {
		t.Fatal(dbname + " creation error")
	}
	defer func() {
		db.Close()
		os.Remove(dbname)
	}()

	db.SetPass("pass")
	_, err2 := db.Get([]byte("table1"), nil, []byte("keyname"))
	if err2 == nil {
		t.Fatal("No error even though the key does not exist")
	}
}

func TestShortValues(t *testing.T) {
	dbname := random.RandStringBytes(10) + ".db"
	db, err := InitBoltFile(dbname)
	defer func() {
		db.Close()
		os.Remove(dbname)
	}()
	if err != nil {
		t.Fatal(dbname + " creation error")
	}
	db.SetPass("pass")
	s := "v"
	db.Set([]byte("table1"), nil, []byte("keyname2"), []byte(s))
	res, err2 := db.Get([]byte("table1"), nil, []byte("keyname2"))
	if err2 != nil {
		t.Fatal("Error of reading keyname2 in table1")
	}
	if string(res) != s {
		t.Fatal("strings are not equal")
	}
}
