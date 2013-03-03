/*
	Reading matlab .mat data files.
	http://www.mathworks.com/help/pdf_doc/matlab/matfile_format.pdf
*/
package mat

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
)

type header struct {
	Text               [116]byte
	SubsytemDataOffset uint64
	Version            uint16
	Endian             uint16
}

type tag struct {
	Type uint32
	Size uint32
}

const (
	miINT8       = 1
	miUINT8      = 2
	miINT16      = 3
	miUINT16     = 4
	miINT32      = 5
	miUINT32     = 6
	miSINGLE     = 7
	miDOUBLE     = 9
	miINT64      = 12
	miUINT64     = 13
	miMATRIX     = 14
	miCOMPRESSED = 15
	miUTF8       = 16
	miUTF16      = 17
	miUTF32      = 18

	mxCELL_CLASS   = 1
	mxSTRUCT_CLASS = 2
	mxOBJECT_CLASS = 3
	mxCHAR_CLASS   = 4
	mxSPARSE_CLASS = 5
	mxDOUBLE_CLASS = 6
	mxSINGLE_CLASS = 7
	mxINT8_CLASS   = 8
	mxUINT8_CLASS  = 9
	mxINT16_CLASS  = 10
	mxUINT16_CLASS = 11
	mxINT32_CLASS  = 12
	mxUINT32_CLASS = 13
	mxINT64_CLASS  = 14
	mxUINT64_CLASS = 15
)

type Array struct {
	Name       string
	Dimensions []int32
	Data       []float64
}

func pad(reader io.Reader, size uint32) error {
	var extra = (8 - size%8) % 8
	if extra != 0 {
		tmp := make([]byte, extra)
		_, err := io.ReadFull(reader, tmp)
		return err
	}

	return nil
}

func readAllElements(reader io.Reader, encoding binary.ByteOrder) (res []interface{}, err error) {
	res = make([]interface{}, 0)

	for {
		var elem interface{}
		if elem, err = readDataElement(reader, encoding); err != nil {
			break
		}
		res = append(res, elem)
	}
	if err == io.EOF {
		return res, nil
	}

	return
}

func charsToString(ca []int8) string {
	s := make([]byte, len(ca))
	var lens int
	for ; lens < len(ca); lens++ {
		if ca[lens] == 0 {
			break
		}
		s[lens] = uint8(ca[lens])
	}
	return string(s[0:lens])
}

func int8ToFloat64(data []int8) []float64 {
	result := make([]float64, len(data))

	for i := range data {
		result[i] = float64(data[i])
	}

	return result
}

func uint8ToFloat64(data []uint8) []float64 {
	result := make([]float64, len(data))

	for i := range data {
		result[i] = float64(data[i])
	}

	return result
}

func readDataElement(reader io.Reader, encoding binary.ByteOrder) (result interface{}, err error) {
	var t tag
	if err = binary.Read(reader, encoding, &t); err != nil {
		return
	}

	if t.Type&0xffff0000 != 0 {
		return nil, errors.New(fmt.Sprintf("Unsupported inline data: %s", t))
	}

	tmp := make([]byte, t.Size)
	if _, err = io.ReadFull(reader, tmp); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't read %s bytes: %s", t.Size, err))
	}
	if t.Type != miCOMPRESSED {
		if err = pad(reader, t.Size); err != nil {
			return nil, errors.New(fmt.Sprintf("Can't pad %s", err))
		}
	}

	tmpReader := bytes.NewReader(tmp)

	// todo: 64-bit padding
	switch t.Type {
	case miINT8:
		res := make([]int8, t.Size)
		err = binary.Read(tmpReader, encoding, &res)
		return res, err
	case miUINT8:
		res := make([]uint8, t.Size)
		err = binary.Read(tmpReader, encoding, &res)
		return res, err
	case miINT32:
		res := make([]int32, t.Size/4)
		err = binary.Read(tmpReader, encoding, &res)
		return res, err
	case miUINT32:
		res := make([]uint32, t.Size/4)
		err = binary.Read(tmpReader, encoding, &res)
		return res, err
	case miCOMPRESSED:
		tmpReader, err := zlib.NewReader(tmpReader)
		if err != nil {
			return nil, err
		}
		return readDataElement(tmpReader, encoding)
	case miMATRIX:
		elems, err := readAllElements(tmpReader, encoding)
		if err != nil {
			return elems, err
		}
		flagsSubelement, ok := elems[0].([]uint32)
		if !ok {
			return elems, errors.New(fmt.Sprintf("Bad flags subelement: %s", elems[0]))
		}
		flags := flagsSubelement[0] & 0xff00
		class := flagsSubelement[0] & 0xff
		if flags != 0 {
			return elems, errors.New(fmt.Sprintf("Non-zero flags not supported: %s", flags))
		}

		dims, ok := elems[1].([]int32)
		if !ok {
			return elems, errors.New(fmt.Sprintf("Bad dims subelement: %s", elems[1]))
		}

		name, ok := elems[2].([]int8)
		if !ok {
			return elems, errors.New(fmt.Sprintf("Bad name subelement: %s", elems[2]))
		}

		switch class {
		case mxDOUBLE_CLASS:
			switch data := elems[3].(type) {
			case []int8:
				return Array{Name: charsToString(name), Dimensions: dims, Data: int8ToFloat64(data)}, nil
			case []uint8:
				return Array{Name: charsToString(name), Dimensions: dims, Data: uint8ToFloat64(data)}, nil
			default:
				return elems, errors.New(fmt.Sprintf("Unsupported elems: %s", reflect.TypeOf(elems[3])))
			}
		default:
			return elems, errors.New(fmt.Sprintf("Unsupported class: %s", class))
		}

		return elems, nil
	default:
		return nil, errors.New(fmt.Sprintf("Unsupported type %s", t))
	}

	panic("unreachable")
}

func Read(reader io.Reader) (result []interface{}, err error) {
	var h header
	var encoding binary.ByteOrder = binary.LittleEndian
	if err = binary.Read(reader, encoding, &h); err != nil {
		return
	}

	if h.Version != 0x0100 {
		err = errors.New("Unsupported version")
		return
	}

	if h.Endian != 0x4d49 {
		encoding = binary.BigEndian
	}

	return readAllElements(reader, encoding)
}
