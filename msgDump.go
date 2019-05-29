package goStrongswanVici

import (
	"bytes"
	"fmt"
)

var dumpLevel uint32

func init() {
	dumpLevel = 3
}

func SetDumpLevel(lvl uint32) {
	dumpLevel = lvl
}

func GetDumpLevel() uint32 {
	return dumpLevel
}

func printBuffer(hdr string, b *bytes.Buffer) {
	if dumpLevel > 0 {
		fmt.Printf("\n*** %s-SSWAN (%d bytes):\n", hdr, b.Len())
	}

	switch dumpLevel {
	case 1:
		dumpBytes(b)

	case 2:
		dumpFields(b)

	case 3:
		dumpFields(b)
		dumpBytes(b)

	default:
	}
}

func dumpBytes(b *bytes.Buffer) {
	for _, c := range b.Bytes() {
		if c > ' ' && c <= '~' {
			fmt.Printf("'%c',", c)
		} else {
			fmt.Printf("%d, ", int8(c))
		}
	}
	fmt.Printf("\n")
}

// Message types
const (
	cmdRequest = iota
	cmdResponse
	cndUnknown
	evtRegister
	evtUnregister
	evtConfirm
	evtUnknown
	event
)

var msgTypes = map[byte]string{
	cmdRequest:    "CMD_REQUEST",
	cmdResponse:   "CMD_RESPONSE",
	cndUnknown:    "CMD_UNKNOWN",
	evtRegister:   "EVENT_REGISTER",
	evtUnregister: "EVENT_UNREGISTER",
	evtConfirm:    "EVENT_CONFIRM",
	evtUnknown:    "EVENT_UNKNOWN",
	event:         "EVENT",
}

// Message element types
const (
	sectionStart = iota + 1
	sectionEnd
	keyValue
	listStart
	listItem
	listEnd
)

func dumpFields(b *bytes.Buffer) {
	sLvl := 0

	msg := b.Bytes()
	mType := msg[0]
	fmt.Printf("MsgType: %s\n", msgTypes[mType])

	if mType == cmdRequest || mType == evtRegister || mType == evtUnregister || mType == event {
		msgNameLen := msg[1]
		msgName := msg[2 : 2+msgNameLen]
		fmt.Printf("MsgName: '%s'\n", msgName)
		msg = msg[2+msgNameLen:]
	} else {
		msg = msg[1:]
	}

	if mType != cmdRequest && mType != cmdResponse && mType != event {
		return
	}

	inList := false
	for len(msg) > 0 {
		var typeName []byte
		switch msg[0] {
		case sectionStart:
			typeName, msg = getTypeName(msg[1:])
			fmt.Printf("%ssection %s {\n", preamble(sLvl), typeName)
			sLvl++

		case sectionEnd:
			sLvl--
			fmt.Printf("%s}\n", preamble(sLvl))
			msg = msg[1:]

		case keyValue:
			typeName, msg = getTypeName(msg[1:])
			fmt.Printf("%s%s = ", preamble(sLvl), typeName)
			msg = printValue(msg)
			fmt.Println()

		case listStart:
			typeName, msg = getTypeName(msg[1:])
			fmt.Printf("%s%s = [", preamble(sLvl), typeName)

		case listItem:
			if inList {
				fmt.Printf(", ")
			}
			msg = printValue(msg[1:])
			inList = true

		case listEnd:
			fmt.Printf("]\n")
			msg = msg[1:]
			inList = false
		}
	}
}

func getTypeName(msg []byte) ([]byte, []byte) {
	nameLen := msg[0]
	return msg[1 : nameLen+1], msg[nameLen+1:]
}

func printValue(msg []byte) []byte {
	pLen := (uint16(msg[0]) << 8) + uint16(msg[1])
	pMsg := msg[2 : 2+pLen]

	hexMode := false
	for _, c := range pMsg {
		if c < ' ' || c > '~' {
			hexMode = true
			break
		}
	}

	if hexMode {
		for _, c := range pMsg {
			fmt.Printf("%02x ", c)
		}
	} else {
		for _, c := range pMsg {
			fmt.Printf("%c", c)
		}
	}

	return msg[pLen+2:]
}

func preamble(l int) string {
	s := ""
	for i := 0; i < l; i++ {
		s = s + "    "
	}
	return s
}
