package jsonapi

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type messageType struct {
	Type string `json:"type"`
}

type queryMessage struct {
	Query string `json:"query"`
}

type queryHostMessage struct {
	Host string `json:"host"`
}

type getLoginMessage struct {
	Entry string `json:"entry"`
}

type loginResponse struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type createEntryMessage struct {
	Name           string `json:"entry_name"`
	Login          string `json:"login"`
	Password       string `json:"password"`
	PasswordLength int    `json:"length"`
	Generate       bool   `json:"generate"`
	UseSymbols     bool   `json:"use_symbols"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func readMessage(r io.Reader) ([]byte, error) {
	stdin := bufio.NewReader(r)
	lenBytes := make([]byte, 4)
	count, err := stdin.Read(lenBytes)
	if err != nil {
		return nil, eofReturn(err)
	}
	if count != 4 {
		return nil, fmt.Errorf("недостаточно прочитано байтов, чтобы определить размер сообщения")
	}

	length, err := getMessageLength(lenBytes)
	if err != nil {
		return nil, err
	}

	msgBytes := make([]byte, length)
	count, err = stdin.Read(msgBytes)
	if err != nil {
		return nil, eofReturn(err)
	}
	if count != length {
		return nil, fmt.Errorf("сообщение прочитано не польностью")
	}

	return msgBytes, nil
}

func getMessageLength(msg []byte) (int, error) {
	var length uint32
	buf := bytes.NewBuffer(msg)
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return 0, err
	}

	return int(length), nil
}

func eofReturn(err error) error {
	if err == io.EOF {
		return nil
	}
	return err
}
