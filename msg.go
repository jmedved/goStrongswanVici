package goStrongswanVici

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
)

type segmentType byte

const (
	stCMD_REQUEST      segmentType = 0
	stCMD_RESPONSE                 = 1
	stCMD_UNKNOWN                  = 2
	stEVENT_REGISTER               = 3
	stEVENT_UNREGISTER             = 4
	stEVENT_CONFIRM                = 5
	stEVENT_UNKNOWN                = 6
	stEVENT                        = 7
)

func (t segmentType) hasName() bool {
	switch t {
	case stCMD_REQUEST, stEVENT_REGISTER, stEVENT_UNREGISTER, stEVENT:
		return true
	}
	return false
}
func (t segmentType) isValid() bool {
	switch t {
	case stCMD_REQUEST, stCMD_RESPONSE, stCMD_UNKNOWN, stEVENT_REGISTER,
		stEVENT_UNREGISTER, stEVENT_CONFIRM, stEVENT_UNKNOWN, stEVENT:
		return true
	}
	return false
}

func (t segmentType) hasMsg() bool {
	switch t {
	case stCMD_REQUEST, stCMD_RESPONSE, stEVENT:
		return true
	}
	return false
}

type elementType byte

const (
	etSECTION_START elementType = 1
	etSECTION_END               = 2
	etKEY_VALUE                 = 3
	etLIST_START                = 4
	etLIST_ITEM                 = 5
	etLIST_END                  = 6
)

type segment struct {
	typ  segmentType
	name string
	msg  map[string]interface{}
}

// msg 在内部以下列3种类型表示(降低复杂度)
// string
// map[string]interface{}
// []string
func writeSegment(w io.Writer, msg segment) (err error) {
	if !msg.typ.isValid() {
		return fmt.Errorf("[writeSegment] msg.typ %d not defined", msg.typ)
	}
	buf := &bytes.Buffer{}
	buf.WriteByte(byte(msg.typ))

	if msg.typ.hasName() {
		err = writeString1(buf, msg.name)
		if err != nil {
			return fmt.Errorf("[writeSegment] error returned from writeString1: '%s'", err)
		}
	}

	if msg.typ.hasMsg() {
		err = writeMap(buf, msg.msg)
		if err != nil {
			return fmt.Errorf("[writeSegment] error returned from writeMap: '%s'", err)
		}
	}

	err = binary.Write(w, binary.BigEndian, uint32(buf.Len()))
	if err != nil {
		return fmt.Errorf("[writeSegment] error writing to binary: '%s'", err)
	}

	printBuffer(buf)
	_, err = buf.WriteTo(w)
	if err != nil {
		return fmt.Errorf("[writeSegment] error writing to buffer: '%s'", err)
	}

	return nil
}

func readSegment(inR io.Reader) (msg segment, err error) {
	var length uint32
	err = binary.Read(inR, binary.BigEndian, &length)
	if err != nil {
		return
	}
	r := bufio.NewReader(&io.LimitedReader{
		R: inR,
		N: int64(length),
	})

	c, err := r.ReadByte()
	if err != nil {
		return
	}
	msg.typ = segmentType(c)
	if !msg.typ.isValid() {
		return msg, fmt.Errorf("[readSegment] msg.typ %d not defined", msg.typ)
	}
	if msg.typ.hasName() {
		msg.name, err = readString1(r)
		if err != nil {
			return
		}
	}
	if msg.typ.hasMsg() {
		msg.msg, err = readMap(r, true)
		if err != nil {
			return
		}
	}
	return
}

func writeString1(w *bytes.Buffer, s string) (err error) {
	length := len(s)
	if length > 255 {
		return fmt.Errorf("[writeString1] string length>255, s='%s'", s)
	}
	if err = w.WriteByte(byte(length)); err != nil {
		return
	}
	_, err = w.WriteString(s)
	return
}

func readString1(r *bufio.Reader) (s string, err error) {
	length, err := r.ReadByte()
	if err != nil {
		return
	}
	buf := make([]byte, length)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return
	}
	return string(buf), nil
}

func writeString2(w *bytes.Buffer, s string) (err error) {
	length := len(s)
	if length > 65535 {
		return fmt.Errorf("[writeString2] string length>65535")
	}
	if err = binary.Write(w, binary.BigEndian, uint16(length)); err != nil {
		return
	}
	_, err = w.WriteString(s)
	return
}

func readString2(r io.Reader) (s string, err error) {
	var length uint16
	err = binary.Read(r, binary.BigEndian, &length)
	if err != nil {
		return
	}
	buf := make([]byte, length)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return
	}
	return string(buf), nil
}

func writeKeyMap(w *bytes.Buffer, name string, msg map[string]interface{}) (err error) {
	w.WriteByte(byte(etSECTION_START))
	if err = writeString1(w, name); err != nil {
		return
	}
	if err = writeMap(w, msg); err != nil {
		return
	}
	return w.WriteByte(byte(etSECTION_END))
}

func writeKeyList(w *bytes.Buffer, name string, msg []string) (err error) {
	w.WriteByte(byte(etLIST_START))
	if err = writeString1(w, name); err != nil {
		return
	}

	for _, s := range msg {
		w.WriteByte(byte(etLIST_ITEM))
		if err = writeString2(w, s); err != nil {
			return
		}
	}
	return w.WriteByte(byte(etLIST_END))
}

func writeKeyString(w *bytes.Buffer, name string, msg string) (err error) {
	w.WriteByte(byte(etKEY_VALUE))
	err = writeString1(w, name)
	if err != nil {
		return
	}
	err = writeString2(w, msg)
	return
}

func writeMap(w *bytes.Buffer, msg map[string]interface{}) (err error) {
	for k, v := range msg {
		switch t := v.(type) {
		case map[string]interface{}:
			err = writeKeyMap(w, k, t)
		case []string:
			err = writeKeyList(w, k, t)
		case string:
			err = writeKeyString(w, k, t)
		case []interface{}:
			str := make([]string, len(t))
			for i := range t {
				str[i] = t[i].(string)
			}
			err = writeKeyList(w, k, str)
		default:
			err = fmt.Errorf("[writeMap] unkown type for value in tuple k='%s', v='%s' has x", k, v)
		}
		if err != nil {
			return
		}
	}
	return nil
}

// SECTION_START has been read already.
func readKeyMap(r *bufio.Reader) (key string, msg map[string]interface{}, err error) {
	key, err = readString1(r)
	if err != nil {
		return
	}
	msg, err = readMap(r, false)
	return
}

// LIST_START has been read already.
func readKeyList(r *bufio.Reader) (key string, msg []string, err error) {
	key, err = readString1(r)
	if err != nil {
		return
	}
	msg = []string{}
	for {
		var c byte
		c, err = r.ReadByte()
		if err != nil {
			return
		}
		switch elementType(c) {
		case etLIST_ITEM:
			value, err := readString2(r)
			if err != nil {
				return "", nil, err
			}
			msg = append(msg, value)
		case etLIST_END: // end of outer list
			return key, msg, nil
		default:
			return "", nil, fmt.Errorf("[readKeyList] protocol error 2")
		}
	}
}

// KEY_VALUE has been read already.
func readKeyString(r *bufio.Reader) (key string, msg string, err error) {
	key, err = readString1(r)
	if err != nil {
		return
	}
	msg, err = readString2(r)
	if err != nil {
		return
	}
	return
}

// Since the original key chosen can have duplicates,
// this function is used to map the original key to a new one
// to make them unique.
func getNewKeyToHandleDuplicates(key string, msg map[string]interface{}) string {
	if _, ok := msg[key]; !ok {
		return key
	}

	for i := 0; ; i++ {
		newKey := key + "##" + strconv.Itoa(i)
		if _, ok := msg[newKey]; !ok {
			return newKey
		}
	}
}

// SECTION_START has been read already.
func readMap(r *bufio.Reader, isRoot bool) (msg map[string]interface{}, err error) {
	msg = map[string]interface{}{}
	for {
		c, err := r.ReadByte()
		if err == io.EOF && isRoot { // may be root section
			return msg, nil
		}
		if err != nil {
			return nil, err
		}
		switch elementType(c) {
		case etSECTION_START:
			key, value, err := readKeyMap(r)
			if err != nil {
				return nil, err
			}
			msg[getNewKeyToHandleDuplicates(key, msg)] = value
		case etLIST_START:
			key, value, err := readKeyList(r)
			if err != nil {
				return nil, err
			}
			msg[getNewKeyToHandleDuplicates(key, msg)] = value
		case etKEY_VALUE:
			key, value, err := readKeyString(r)
			if err != nil {
				return nil, err
			}
			msg[getNewKeyToHandleDuplicates(key, msg)] = value
		case etSECTION_END: // end of outer section
			return msg, nil
		default:
			panic(fmt.Errorf("[readMap] protocol error 1, %d %#v", c, msg))
			// return nil, fmt.Errorf("[readMap] protocol error 1, %d",c)
		}
	}
	return
}

func printBuffer(b *bytes.Buffer) {
	fmt.Printf("***** byte buffer (len %d/%d: ", b.Len(), len(b.Bytes()))
	for _, c := range b.Bytes() {
		if c > ' ' && c <= '~' {
			fmt.Printf("'%c',", c)
		} else {
			fmt.Printf("%d, ", c)
		}
	}
	fmt.Printf("\n")
}
