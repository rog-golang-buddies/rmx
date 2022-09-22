package suid

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestUUID(t *testing.T) {
	uid := NewUUID()

	b, err := json.Marshal(uid.ShortUUID())
	if err != nil {
		t.Fatal(err)
	}

	var sid SUID
	if err := json.NewDecoder(strings.NewReader(string(b))).Decode(&sid); err != nil {
		t.Fatal(err)
	}

	oid, err := sid.UUID()
	if err != nil {
		t.Fatal(err)
	}

	if oid != uid {
		t.Fatalf("expected: %s;got %s\n", uid, sid)
	}
}

func TestSIze(t *testing.T) {
	s1 := NewSUID()
	s2 := NewSUID()
	s3 := NewSUID()

	t.Log(len(s1), len(s2), len(s3))
}