package common

import (
    "fmt"
    "net"
    "strings"
	"encoding/binary"
)

const (
    FIELD_COUNT_RESPONSE = 2
	NID = 0
	NUMBER = 1
	MAX_LENGTH = 1024
	SIZE = 2
)

type ClientProtocol struct {
	conn   net.Conn
	maxLength int
}

func NewClientProtocol(conn net.Conn) *ClientProtocol {
	client_protocol := &ClientProtocol{
		conn: conn,
		maxLength: MAX_LENGTH,
	}
	return client_protocol
}

func (cp *ClientProtocol) serializeBet(bet Bet) string {
    return fmt.Sprintf("%s|%s|%s|%s|%s|%s\n",
        bet.Agency, bet.FirstName, bet.LastName, bet.Document, bet.Birthdate, bet.Number)
}

func (cp *ClientProtocol) sendBetSize(size int) error {
    buf := make([]byte, 2)
    binary.BigEndian.PutUint16(buf, uint16(size))
    return cp.sendAllBytes(buf)
}


func (cp *ClientProtocol) sendAllBytes(buf []byte) error {
    total := 0
    for total < len(buf) {
        n, err := cp.conn.Write(buf[total:])
        if err != nil {
            return err
        }
        total += n
    }
    return nil
}

func (cp *ClientProtocol) sendAllMessage(message string) error {
    return cp.sendAllBytes([]byte(message))
}

func (cp *ClientProtocol) sendBet(bet Bet) error {

	var msg strings.Builder
	msg.WriteString(cp.serializeBet(bet))

	message := msg.String()
	sizeBet := len(message)

	if sizeBet > cp.maxLength {
		return fmt.Errorf("message too long")
	}

    if err := cp.sendBetSize(sizeBet); err != nil { return err }
    return cp.sendAllMessage(message)
}

func (cp *ClientProtocol) recvSizeMsg () (int, error) {
	buf := make([]byte, SIZE)
	total := 0
	for total < SIZE {
		n, err := cp.conn.Read(buf[total:])
		if err != nil {
			return 0, err
		}
		total += n
	}
	return int(binary.BigEndian.Uint16(buf)), nil
}



func (cp *ClientProtocol) recvResponseBet() (string, string, error) {
	size, err := cp.recvSizeMsg()
	if err != nil {
		return "", "", err
	}
	
	buf := make([]byte, size)
    total := 0
    for total < size {
        n, err := cp.conn.Read(buf[total:])
        if err != nil { return "", "", err }
        total += n
    }

    payload := string(buf)
    parts := strings.Split(payload, "|")
	
	if len(parts) != FIELD_COUNT_RESPONSE {
		return "", "", fmt.Errorf("invalid response format")
	}

	return parts[NID], parts[NUMBER], nil
}
