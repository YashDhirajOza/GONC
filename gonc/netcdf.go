package gonc

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)
const (
	ClassicFormat     = 1
	Format64BitOffset = 2
)
type File struct {
	f       *os.File
	Format  byte
	NumRecs uint32
	Dims    []Dimension
}

type Dimension struct {
	Name   string
	Length uint32
}

func Open(path string) (*File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	header := make([]byte, 4)
	_, err = f.Read(header)
	if err != nil {
		f.Close()
		return nil, err
	}
	if string(header[:3]) != "CDF" {
		f.Close()
		return nil, errors.New("not a NetCDF file")
	}
	format := header[3]
	if format != ClassicFormat && format != Format64BitOffset {
		f.Close()
		return nil, fmt.Errorf("unsupported NetCDF format: %d", format)
	}

	buf := make([]byte, 4)
	_, err = f.Read(buf)
	if err != nil {
		f.Close()
		return nil, err
	}
	numrecs := binary.BigEndian.Uint32(buf)

	nc := &File{
		f:       f,
		Format:  format,
		NumRecs: numrecs,
		Dims:    []Dimension{},
	}
	if err := nc.readDimList(); err != nil {
		f.Close()
		return nil, err
	}
	return nc, nil
}

func readU32(f *os.File) (uint32, error) {
	buf := make([]byte, 4)
	_, err := f.Read(buf)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(buf), nil
}

func readString(f *os.File) (string, error) {
	n, err := readU32(f) // number of characters
	if err != nil {
		return "", err
	}

	buf := make([]byte, n)
	_, err = f.Read(buf)
	if err != nil {
		return "", err
	}

	// skip padding
	pad := (4 - (n % 4)) % 4
	if pad > 0 {
		_, err = f.Seek(int64(pad), 1)
		if err != nil {
			return "", err
		}
	}

	return string(buf), nil
}

func (nc *File) Close() error {
	return nc.f.Close()
}
func (nc *File) readDimList() error {
	tag, err := readU32(nc.f)
	if err != nil {
		return err
	}
	if tag == 0 { 
		nc.Dims = []Dimension{}
		return nil
	}
	if tag != 0x0A { 
		return fmt.Errorf("invalid dim_list tag: %d", tag)
	}
	nelems, err := readU32(nc.f)
	if err != nil {
		return err
	}
	dims := make([]Dimension, 0, nelems)
	for i := 0; i < int(nelems); i++ {
		name, err := readString(nc.f)
		if err != nil {
			return err
		}
		length, err := readU32(nc.f)
		if err != nil {
			return err
		}
		dims = append(dims, Dimension{Name: name, Length: length})
	}
	nc.Dims = dims
	return nil
}
