package boltdb

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	c "github.com/lazy-am/zvart/pkg/cipher"
	bolt "go.etcd.io/bbolt"
)

const (
	aesteststring = "ZWART is a program for people"
	userBucket    = "user"
	aestestKey    = "aesteststring"
	nameKey       = "name"
	onionKey      = "onion"
)

type db struct {
	Path string
	Bolt *bolt.DB
	hash []byte
	Err  error
}

func InitBoltFile(fn string) (*db, error) {
	b := db{Path: fn}
	db, err := bolt.Open(fn, 0600, &bolt.Options{Timeout: 10 * time.Second})
	if err != nil {
		b.Err = err
		return nil, err
	}
	b.Bolt = db
	b.initTable()
	return &b, nil
}

func (db *db) initTable() {
	db.CreateTable([]byte("user"), nil)
	db.CreateTable([]byte("contacts"), nil)
	db.CreateTable([]byte("messages"), nil)
}

func (db *db) GetError() error {
	return db.Err
}

func (db *db) Close() {
	db.Bolt.Close()
}

func (db *db) SetPass(pass string) {
	db.hash = c.GetSHA256([]byte(pass))
	db.saveTestString()
}

func (db *db) saveTestString() {
	if db.Err != nil {
		return
	}

	encoded, err := c.AESEncript(db.hash, []byte(aesteststring))
	if err != nil {
		db.Err = err
		return
	}

	db.update([]byte(userBucket), nil, []byte(aestestKey), encoded)

}

func (db *db) TestPass(pass string) bool {

	if db.Err != nil {
		return false
	}

	dat, err := db.view([]byte(userBucket), nil, []byte(aestestKey))
	if err != nil {
		db.Err = err
		return false
	} else if dat == nil {
		db.Err = errors.New("empty aesteststring string")
		return false
	}

	hash := c.GetSHA256([]byte(pass))
	encoded, err := c.AESDecript(hash, []byte(dat))
	if err != nil {
		return false
	}

	if string(encoded) == aesteststring {
		db.hash = hash
		return true
	}

	return false
}

func (db *db) Set(table, subtable, key, value []byte) error {
	if db.Err != nil {
		return db.Err
	}
	if db.hash == nil {
		return errors.New("password is not set")
	}
	enc, err := c.AESEncript(db.hash, []byte(value))
	if err != nil {
		db.Err = err
		return err
	}
	return db.update(table, subtable, key, enc)
}

func (db *db) Get(table, subtable, key []byte) ([]byte, error) {
	if db.Err != nil {
		return nil, db.Err
	}
	if db.hash == nil {
		return nil, errors.New("password is not set")
	}
	buf, err := db.view(table, subtable, key)
	if err != nil {
		return nil, err
	}

	res, err := c.AESDecript(db.hash, buf)
	if err != nil {
		db.Err = err
		return nil, err
	}
	return res, nil
}

func (db *db) update(bucket, subbucket, key, val []byte) error {

	return db.Bolt.Update(func(tx *bolt.Tx) error {

		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			db.Err = err
			return err
		}
		if subbucket != nil {
			b, err = b.CreateBucketIfNotExists(subbucket)
			if err != nil {
				db.Err = err
				return err
			}
		}
		b.Put([]byte(key), val)
		if err != nil {
			db.Err = err
			return err
		}
		return nil

	})

}

func (db *db) view(bucket, subbucket, key []byte) ([]byte, error) {
	var buf []byte
	err := db.Bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return fmt.Errorf("bucket %s return nil", string(bucket))
		}
		if subbucket != nil {
			b = b.Bucket(subbucket)
			if b == nil {
				return fmt.Errorf("subbucket %s return nil", string(bucket))
			}
		}
		v := b.Get(key)
		if v != nil {
			buf = make([]byte, len(v))
			copy(buf, v)
		}
		return nil
	})
	if (buf == nil) && (err == nil) {
		err = errors.New("key does not exist")
	}
	return buf, err
}

func (db *db) GetNextId(table []byte, subtable []byte) ([]byte, error) {

	if db.Err != nil {
		return nil, db.Err
	}

	var buf []byte = make([]byte, 8)
	err := db.Bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(table)
		if b == nil {
			return fmt.Errorf("bucket %s return nil", string(table))
		}
		if subtable != nil {
			b = b.Bucket(subtable)
			if b == nil {
				return fmt.Errorf("subbucket %s return nil", string(subtable))
			}
		}
		id, err := b.NextSequence()
		if err != nil {
			return err
		}
		binary.LittleEndian.PutUint64(buf, id)
		return nil
	})
	return buf, err
}

func (db *db) ClearTable(table, subtable []byte) error {
	if db.Err != nil {
		return db.Err
	}
	err := db.Bolt.Update(func(tx *bolt.Tx) error {

		if subtable == nil {
			err := tx.DeleteBucket(table)
			if err != nil {
				return err
			}
		} else {
			b := tx.Bucket(table)
			if b == nil {
				return fmt.Errorf("bucket %s return nil", string(table))
			}
			err := b.DeleteBucket(subtable)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return db.CreateTable(table, subtable)
}

func (db *db) CreateTable(table, subtable []byte) error {
	if db.Err != nil {
		return db.Err
	}

	return db.Bolt.Update(func(tx *bolt.Tx) error {

		b, err := tx.CreateBucketIfNotExists([]byte(table))
		if err != nil {
			db.Err = err
			return err
		}
		if subtable != nil {
			_, err = b.CreateBucketIfNotExists(subtable)
			if err != nil {
				db.Err = err
				return err
			}
		}
		return nil
	})
}

func (db *db) LoadList(table []byte, subtable []byte) (map[uint64][]byte, error) {

	if db.Err != nil {
		return nil, db.Err
	}

	if db.hash == nil {
		return nil, errors.New("password is not set")
	}

	list := make(map[uint64][]byte)
	err := db.Bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(table)
		if b == nil {
			return fmt.Errorf("bucket %s return nil", string(table))
		}
		if subtable != nil {
			b = b.Bucket(subtable)
			if b == nil {
				return fmt.Errorf("subbucket %s return nil", string(subtable))
			}
		}
		cur := b.Cursor()
		for k, v := cur.First(); k != nil; k, v = cur.Next() {
			elem, err := c.AESDecript(db.hash, v)
			if err != nil {
				db.Err = err
				return err
			}
			list[binary.LittleEndian.Uint64(k)] = elem
		}
		return nil
	})
	return list, err
}

func (db *db) LoadListFromId(table []byte, subtable []byte, id []byte) (map[uint64][]byte, error) {

	if db.Err != nil {
		return nil, db.Err
	}

	if db.hash == nil {
		return nil, errors.New("password is not set")
	}

	list := make(map[uint64][]byte)
	err := db.Bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(table)
		if b == nil {
			return fmt.Errorf("bucket %s return nil", string(table))
		}
		if subtable != nil {
			b = b.Bucket(subtable)
			if b == nil {
				return fmt.Errorf("subbucket %s return nil", string(subtable))
			}
		}
		cur := b.Cursor()
		for k, v := cur.Seek(id); k != nil; k, v = cur.Next() {
			elem, err := c.AESDecript(db.hash, v)
			if err != nil {
				db.Err = err
				return err
			}
			list[binary.LittleEndian.Uint64(k)] = elem
		}
		return nil
	})
	return list, err
}

func (db *db) LoadLast(table []byte, subtable []byte, count uint64) (lst map[uint64][]byte,
	startIndex uint64,
	er error) {
	if db.Err != nil {
		return nil, 0, db.Err
	}

	if db.hash == nil {
		return nil, 0, errors.New("password is not set")
	}

	list := make(map[uint64][]byte)
	err := db.Bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(table)
		if b == nil {
			return fmt.Errorf("bucket %s return nil", string(table))
		}
		if subtable != nil {
			b = b.Bucket(subtable)
			if b == nil {
				return fmt.Errorf("subbucket %s return nil", string(subtable))
			}
		}
		cur := b.Cursor()
		k, _ := cur.Last()
		if k == nil {
			return errors.New("empty table")
		}
		startIndex = binary.LittleEndian.Uint64(k)
		if startIndex > count {
			startIndex -= count
		} else {
			startIndex = 1
		}
		startKey := make([]byte, 8)
		binary.LittleEndian.PutUint64(startKey, startIndex)
		for k, v := cur.Seek(startKey); k != nil; k, v = cur.Next() {
			elem, err := c.AESDecript(db.hash, v)
			if err != nil {
				db.Err = err
				return err
			}
			list[binary.LittleEndian.Uint64(k)] = elem
		}
		return nil
	})
	return list, startIndex, err
}

func (db *db) GetMaxId(table []byte, subtable []byte) (uint64, error) {
	if db.Err != nil {
		return 0, db.Err
	}

	var rez uint64 = 0
	err := db.Bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(table)
		if b == nil {
			return fmt.Errorf("bucket %s return nil", string(table))
		}
		if subtable != nil {
			b = b.Bucket(subtable)
			if b == nil {
				return fmt.Errorf("subbucket %s return nil", string(subtable))
			}
		}
		cur := b.Cursor()
		k, _ := cur.Last()
		if k == nil {
			return errors.New("empty table")
		}
		rez = binary.LittleEndian.Uint64(k)
		return nil
	})
	return rez, err
}
