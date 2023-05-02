package rwho

import (
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
)

func bytesToString(b []byte) string {
	end := len(b)
	for i := 0; i < len(b); i++ {
		if b[i] == 0 {
			end = i
			break
		}
	}
	return string(b[:end])
}

func ParseWhoEntry(r io.Reader) (*WhoEntry, error) {
	e := new(WhoEntry)
	err := binary.Read(r, binary.LittleEndian, e)
	return e, err
}

func ParseWhodHeader(r io.Reader) (*WhodHeader, error) {
	h := new(WhodHeader)
	err := binary.Read(r, binary.LittleEndian, h)
	return h, err
}

func ReadWhod(path string) (*WhodHeader, []*WhoEntry, error) {
	// Open the rwho packet file.
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	// Read the rwho packet from the file.
	header, err := ParseWhodHeader(f)
	if err != nil {
		return nil, nil, err
	}

	entries := make([]*WhoEntry, 0)
	for {
		e, err := ParseWhoEntry(f)
		if err != nil {
			break
		}
		entries = append(entries, e)
	}

	return header, entries, nil
}

func ParseWhod(r io.Reader) (*Whod, error) {
	w := new(Whod)
	var err error

	w.Header, err = ParseWhodHeader(r)
	if err != nil {
		return nil, err
	}

	for {
		e, err := ParseWhoEntry(r)
		if err != nil {
			break
		}
		w.WhoEntries = append(w.WhoEntries, e)
	}

	return w, nil
}

func ReadWhodFile(path string) (*Whod, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseWhod(f)
}

func ScanHosts() ([]*Whod, error) {
	paths, err := filepath.Glob("/var/spool/rwho/whod.*")
	if err != nil {
		return nil, err
	}
	l := make([]*Whod, 0)
	for _, p := range paths {
		w, err := ReadWhodFile(p)
		if err != nil {
			return nil, err
		}
		l = append(l, w)
	}
	return l, nil
}
