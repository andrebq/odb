package main

import (
	"testing"
	"reflect"
	"os"
	"path/filepath"
	"io/ioutil"
)

func TestPutGet(t *testing.T) {
	dir, err := ioutil.TempDir("", "odbd-test")
	noError(err)
	defer os.RemoveAll(dir)

	s := NewServer(filepath.Join(dir, "server"))
	if version, err := s.Put("/abc/1-2.3", []byte("hello")); err != nil {
		t.Fatalf("%+v", err)
	} else if version != 1 {
		t.Fatal("invalid version")
	}
	if version, err := s.Put("/abc/1-2.3", []byte("hello2")); err != nil {
		t.Fatalf("%+v", err)
	} else if version != 2 {
		t.Fatal("invalid version")
	}

	if value, version, err := s.Get("/abc/1-2.3", 0); err != nil {
		t.Fatalf("%+v", err)
	} else if !reflect.DeepEqual(value, []byte("hello2")) {
		t.Error("invalid value read")
	} else if version != 2 {
		t.Error("invalid version")
	}

	if versions, err := s.Versions("/abc/1-2.3"); err != nil {
		t.Fatalf("%+v", err)
	} else if !reflect.DeepEqual(versions, []uint64{1,2}) {
		t.Error("invalid version set")
	}

	if value, _, err := s.Get("/abc/1-2.3", 1); err != nil {
		t.Fatalf("%+v", err)
	} else if !reflect.DeepEqual(value, []byte("hello")) {
		t.Error("invalid value read")
	}

	if value, _, err := s.Get("/abc/1-2.3", 2); err != nil {
		t.Fatalf("%+v", err)
	} else if !reflect.DeepEqual(value, []byte("hello2")) {
		t.Error("invalid value read")
	}
}