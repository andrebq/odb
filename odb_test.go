package odb

import "testing"

import "reflect"

func TestPutGet(t *testing.T) {
	s, err := Connect("localhost:5432", "fda_owner", "fda_owner")
	if err != nil {
		t.Fatal(err)
	}
	err = s.TruncateAll()
	if err != nil {
		t.Fatal(err)
	}
	type hello struct {
		Msg string
	}
	col, err := s.Collection("some-user", "some-db", "some-col")
	_, err = col.PutObject("key", hello{Msg: "hi"})
	if err != nil {
		t.Fatal(err)
	}
	var out hello
	err = col.GetObject(&out, "key")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(out, hello{Msg: "hi"}) {
		t.Error("Objects don't match")
	}
}
