package tarfs

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
)

type tarfs func() (io.ReadCloser, error)

func New(opener func() (io.ReadCloser, error)) http.FileSystem {
	return tarfs(opener)
}

func TarIterate(archive *tar.Reader, cb func(*tar.Header, io.Reader) error) error {
	for {
		hdr, err := archive.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err := cb(hdr, archive); err != nil {
			return err
		}
	}
	return nil
}

func (t tarfs) open(name string) (file *tar.Header, dirEntries []*tar.Header, contents []byte, err error) {
	name = path.Clean(name)

	var tf io.ReadCloser
	tf, err = t()
	if err != nil {
		return
	}
	defer tf.Close()

	archive := tar.NewReader(tf)

	err = TarIterate(archive, func(hdr *tar.Header, r io.Reader) error {
		hdrName := path.Clean("/" + hdr.Name)
		if hdrName == name || hdrName == name+"/" {
			file = hdr
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, archive); err != nil {
				return err
			}
			contents = buf.Bytes()
		}
		if path.Dir(hdrName) == name {
			dirEntries = append(dirEntries, hdr)
		}
		return nil
	})
	return
}

type dir struct {
	fi         os.FileInfo
	dirEntries []os.FileInfo
}

func (d *dir) Close() error {
	return nil
}

func (d *dir) Read(buf []byte) (int, error) {
	return 0, io.EOF
}

func (d *dir) Readdir(count int) ([]os.FileInfo, error) {
	if count > len(d.dirEntries) {
		count = len(d.dirEntries)
	}
	var r []os.FileInfo
	r, d.dirEntries = d.dirEntries[:count], d.dirEntries[count:]
	return r, nil
}

func (d *dir) Seek(offset int64, whence int) (int64, error) {
	return 0, fmt.Errorf("a directory is not seekable")
}

func (d *dir) Stat() (os.FileInfo, error) {
	return d.fi, nil
}

type file struct {
	fi os.FileInfo
	io.ReadSeeker
}

func (f *file) Close() error {
	return nil
}

func (f *file) Readdir(count int) ([]os.FileInfo, error) {
	return nil, fmt.Errorf("not a directory")
}

func (f *file) Stat() (os.FileInfo, error) {
	return f.fi, nil
}

var errFileType = fmt.Errorf("only regular files and directories are supported")

func (t tarfs) Open(name string) (http.File, error) {
	fileHdr, dirEntries, contents, err := t.open(name)
	if err != nil {
		return nil, err
	}
	if fileHdr == nil {
		return nil, os.ErrNotExist
	}
	fi := fileHdr.FileInfo()
	if fi.Mode().IsDir() {
		fis := make([]os.FileInfo, len(dirEntries))
		for i, h := range dirEntries {
			fis[i] = h.FileInfo()
		}
		return &dir{fi, fis}, nil
	}
	if fi.Mode().IsRegular() {
		return &file{fi, bytes.NewReader(contents)}, nil
	}
	// TODO: Support sym- and hardlinks (figure out how to do that -- just serve the file or issue a redirect?)
	// If we redirect, we must check that this doesn't point into another repository.
	return nil, errFileType
}
