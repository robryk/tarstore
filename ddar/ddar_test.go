package ddar

import (
	"testing"
	"io/ioutil"
	"strings"
	"os"
)

func TestSmoke(t *testing.T) {
	path, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("TempDir: %s", err)
	}
	defer os.RemoveAll(path)

	a := Archive(path + "/arc")
	err = a.Add("a", strings.NewReader("A"))
	if err != nil {
		t.Fatalf("Add(a): %s", err)
	}
	r, err := a.Get("a")
	if err != nil {
		t.Fatalf("Get(a): %s", err)
	}
	contents, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll(a): %s", err)
	}
	if string(contents) != "A" {
		t.Fatalf("Get(a): got %+v, want %+v", string(contents), "A")
	}
}
