package chezmoi

import (
	"os"
	"path/filepath"

	vfs "github.com/twpayne/go-vfs"
	bolt "go.etcd.io/bbolt"
)

// A BoltPersistentState is a state persisted with bolt.
type BoltPersistentState struct {
	fs   vfs.FS
	path string
	perm os.FileMode
	db   *bolt.DB
}

// NewBoltPersistentState returns a new, unopened BoltPersistentState.
func NewBoltPersistentState(fs vfs.FS, path string) (*BoltPersistentState, error) {
	b := &BoltPersistentState{
		fs:   fs,
		path: path,
		perm: 0600,
	}
	_, err := fs.Stat(b.path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return b, nil
}

// Open the connection to b.
func (b *BoltPersistentState) Open(write bool) error {
	return b.openDB(write)
}

// Close the connection to b.
func (b *BoltPersistentState) Close() error {
	if b.db == nil {
		return nil
	}
	if err := b.db.Close(); err != nil {
		return err
	}
	b.db = nil
	return nil
}

// Delete deletes the value associate with key in bucket. If bucket or key does
// not exist then Delete does nothing.
func (b *BoltPersistentState) Delete(bucket, key []byte) error {
	if b.db == nil {
		return nil
	}
	return b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return nil
		}
		return b.Delete(key)
	})
}

// Get returns the value associated with key in bucket.
func (b *BoltPersistentState) Get(bucket, key []byte) ([]byte, error) {
	value := []byte(nil)
	if b.db == nil {
		return value, nil
	}
	return value, b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return nil
		}
		v := b.Get(key)
		if v != nil {
			value = make([]byte, len(v))
			copy(value, v)
		}
		return nil
	})
}

// Set sets the value associated with key in bucket. bucket will be created if
// it does not already exist.
func (b *BoltPersistentState) Set(bucket, key, value []byte) error {
	if b.db == nil {
		if err := b.openDB(true); err != nil {
			return err
		}
		defer b.Close()
	}
	return b.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return err
		}
		return b.Put(key, value)
	})
}

func (b *BoltPersistentState) openDB(write bool) error {
	parentDir := filepath.Dir(b.path)
	if _, err := b.fs.Stat(parentDir); os.IsNotExist(err) {
		if err := vfs.MkdirAll(b.fs, parentDir, 0755); err != nil {
			return err
		}
	}
	options := &bolt.Options{
		OpenFile: b.fs.OpenFile,
		ReadOnly: !write,
	}
	db, err := bolt.Open(b.path, b.perm, options)
	if err != nil {
		return err
	}
	b.db = db
	return err
}
