package main

import (
	"fmt"
	"bytes"
	"encoding/binary"
	"github.com/pkg/errors"

	bolt "github.com/etcd-io/bbolt"
)

type (
	// server wraps a boltdb so it provides a versioned API
	Server struct {
		db *bolt.DB
	}
)

var (
	itemsBucket = []byte("items")
	counterBucket = []byte("counters")
)

func noError(err error) {
	if err != nil {
		panic(err)
	}
}

func dontPanic(fn func()) (err error) {
	defer func() {
		if p := recover(); p != nil {
			switch p := p.(type) {
			case error:
				err = p
			default:
				err = fmt.Errorf("Error: %v", p)
			}
		}
		if err != nil {
			println("here")
			err = errors.WithStack(err)
		}
	}()
	fn()
	return
}

// NewServer creates a new instance using registry as its storage location
func NewServer(registry string) *Server {
	db, err := bolt.Open(registry, 0600, nil)
	noError(err)
	noError(dontPanic(func() {
		db.Update(func(tx *bolt.Tx) error {
			ensureBucket(tx, itemsBucket)
			ensureBucket(tx, counterBucket)
			return nil
		})
	}))
	return &Server{db}
}

// Put updates the given entry with the provide value, old versions are kept
// until manually removed or a GC happens (no GC implemented, so versions will live forever).
//
// The version counter is returned, keep in mind that versions is a monotonically increasing value
// but is not sequential.
//
// Use Server.Versions to get a view of all versions of a given item
func (s *Server) Put(entry string, value []byte) (uint64, error) {
	var version uint64
	err := s.db.Update(func(tx *bolt.Tx) error {
		return dontPanic(func() {
			count := incCounter(tx.Bucket(counterBucket), entry)
			putCounter(tx.Bucket(counterBucket), entry, count)
			putItem(tx.Bucket(itemsBucket), entry, count, value)
			version = count
		})
	})
	return version, err
}

// Get return the value under the given (entry,version), if the given version is not valid
// for an item, it will return an empty array
//
// If you want the latest version, use version==0 so the system will decide which version to use.
func (s *Server) Get(entry string, version uint64) ([]byte, uint64, error) {
	var val []byte
	err := s.db.View(func(tx *bolt.Tx)error {
		return dontPanic(func(){
			if version == 0 {
				version = getCounter(tx.Bucket(counterBucket), entry)
			}
			if version == 0 {
				val = nil
				return
			}
			val = copyItem(tx.Bucket(itemsBucket), entry, version)
		})
	})
	return val, version, err
}

// Versions returns the list of versions available for a given entry
func (s *Server) Versions(entry string) ([]uint64, error) {
	var versions []uint64
	err := dontPanic(func() {
		s.db.View(func(tx *bolt.Tx) error {
			cursor := tx.Bucket(itemsBucket).Cursor()
			prefix := toDbKeyPrefix(entry)
			for key, _ := cursor.Seek(prefix); hasPrefix(key, prefix); key, _ = cursor.Next() {
				versions = append(versions, parseCounter(key[len(prefix):]))
			}
			return nil
		})
	})
	return versions, err
}

func ensureBucket(tx *bolt.Tx, name []byte) *bolt.Bucket {
	bucket, err := tx.CreateBucketIfNotExists(name)
	noError(err)
	return bucket
}

func putCounter(b *bolt.Bucket, key string, version uint64) {
	buf := toBytes(version)
	err := b.Put([]byte(key), buf[:])
	noError(err)
}

func putItem(b *bolt.Bucket, entry string, version uint64, value []byte) {
	err := b.Put(toDbKey(entry, version), value)
	noError(err)
}

func getItem(b *bolt.Bucket, entry string, version uint64) []byte {
	if version == 0 {
		return nil
	}
	return b.Get(toDbKey(entry, version))
}

func copyItem(b *bolt.Bucket, entry string, version uint64) []byte {
	ret := getItem(b, entry, version)
	if ret == nil {
		return nil
	}
	return append([]byte(nil), ret...)
}

func getCounter(b *bolt.Bucket, key string) uint64 {
	value := getValue(b, []byte(key))
	return parseCounter(value)
}

func parseCounter(value []byte) uint64 {
	switch len(value) {
	case 0:
		return 0
	case 8:
		return binary.BigEndian.Uint64(value)
	}
	panic("corrupted database")
}

func getValue(b *bolt.Bucket, key []byte) ([]byte) {
	return b.Get(key)
}

func incCounter(b *bolt.Bucket, key string) uint64 {
	count, err := b.NextSequence()
	noError(err)
	return count
}

func toBytes(v uint64) (ret [8]byte) {
	binary.BigEndian.PutUint64(ret[:], v)
	return
}

func toDbKey(entry string, version uint64) []byte {
	ret := toDbKeyPrefix(entry)
	buf := toBytes(version)
	ret = append(ret, buf[:]...)
	return ret
}

func toDbKeyPrefix(entry string) []byte {
	var ret []byte
	ret = append(ret, entry...)
	ret = append(ret, byte('@'))
	return ret
}

func hasPrefix(value, prefix []byte) bool {
	return bytes.HasPrefix(value, prefix)
}