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
	Vars    []Variable
}

type Dimension struct {
	Name   string
	Length uint32
}

type Variable struct {
	Name     string
	DimIDs   []uint32
	DataType uint32
	VSize    uint32
	Offset   uint32
	Attrs    []Attribute
}

type Attribute struct {
	Name   string
	Type   uint32
	Values []byte
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
		Vars:    []Variable{},
	}

	if err := nc.readDimList(); err != nil {
		f.Close()
		return nil, err
	}

	if err := nc.readVarList(); err != nil {
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
	n, err := readU32(f)
	if err != nil {
		return "", err
	}

	buf := make([]byte, n)
	_, err = f.Read(buf)
	if err != nil {
		return "", err
	}

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

func (nc *File) readAttrList() ([]Attribute, error) {
	tag, err := readU32(nc.f)
	if err != nil {
		return nil, err
	}

	if tag == 0 {
		return []Attribute{}, nil
	}

	if tag != 0x0C {
		return nil, fmt.Errorf("invalid attr_list tag: %d", tag)
	}

	nelems, err := readU32(nc.f)
	if err != nil {
		return nil, err
	}

	attrs := make([]Attribute, 0, nelems)

	for i := 0; i < int(nelems); i++ {
		name, err := readString(nc.f)
		if err != nil {
			return nil, err
		}

		atype, err := readU32(nc.f)
		if err != nil {
			return nil, err
		}

		nvals, err := readU32(nc.f)
		if err != nil {
			return nil, err
		}

		buf := make([]byte, nvals)
		_, err = nc.f.Read(buf)
		if err != nil {
			return nil, err
		}

		pad := (4 - (nvals % 4)) % 4
		if pad > 0 {
			nc.f.Seek(int64(pad), 1)
		}

		attrs = append(attrs, Attribute{
			Name:   name,
			Type:   atype,
			Values: buf,
		})
	}

	return attrs, nil
}

func (nc *File) readVarList() error {
	tag, err := readU32(nc.f)
	if err != nil {
		return err
	}

	if tag == 0 {
		nc.Vars = []Variable{}
		return nil
	}

	if tag != 0x0B {
		return fmt.Errorf("invalid var_list tag: %d", tag)
	}

	nelems, err := readU32(nc.f)
	if err != nil {
		return err
	}

	vars := make([]Variable, 0, nelems)

	for i := 0; i < int(nelems); i++ {

		name, err := readString(nc.f)
		if err != nil {
			return err
		}

		nDims, err := readU32(nc.f)
		if err != nil {
			return err
		}

		dimIDs := make([]uint32, nDims)
		for j := uint32(0); j < nDims; j++ {
			dimIDs[j], err = readU32(nc.f)
			if err != nil {
				return err
			}
		}

		attrs, err := nc.readAttrList()
		if err != nil {
			return err
		}

		dtype, err := readU32(nc.f)
		if err != nil {
			return err
		}

		vsize, err := readU32(nc.f)
		if err != nil {
			return err
		}

		offset, err := readU32(nc.f)
		if err != nil {
			return err
		}

		vars = append(vars, Variable{
			Name:     name,
			DimIDs:   dimIDs,
			Attrs:    attrs,
			DataType: dtype,
			VSize:    vsize,
			Offset:   offset,
		})
	}

	nc.Vars = vars
	return nil
}
