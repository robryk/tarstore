package tarfs

import (
	"os"
	"io/ioutil"
	"io"
	"os/exec"
	"testing"
)

func TestSmoke(t *testing.T) {
	cmd := exec.Command("tar", "-C", "testdata", "-cf", "testdata.tar", ".")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("Error running tar: %s", err)
	}
	fs := New(func() (io.ReadCloser, error) { return os.Open("testdata.tar") })
	a, err := fs.Open("/a")
	if err != nil {
		t.Fatalf("Error opening a: %s", err)
	}
	flist, err := a.Readdir(100)
	if err != nil {
		t.Fatalf("Readdir(a): %s", err)
	}
	if len(flist) != 1 {
		t.Fatalf("Readdir returned %d items: %+v", len(flist), flist)
	}
	if flist[0].Name() != "z" {
		t.Fatalf("Readdir returned wrong filename: %s", flist[0].Name())
	}
	z, err := fs.Open("/a/z")
	if err != nil {
		t.Fatalf("Open(/a/z): %s", err)
	}
	contents, err := ioutil.ReadAll(z)
	if err != nil {
		t.Fatalf("ReadAll(/a/z): %s", err)
	}
	if string(contents) != "zawartosc z\n" {
		t.Fatalf("ReadAll(/a/z): got %+v, expected %+v", string(contents), "zawartosc z\n")
	}
}
