package common

import (
    "encoding/binary"
    "fmt"
    "net"
    "strings"
)

const (
    BATCH_FINISHED = 0
    ONE_BYTE = 1
    SIZE = 2

    // server to client
    BATCH_OK = 3
    BATCH_FAIL = 4
)

type ClientProtocol struct {
	conn   net.Conn
	maxLength int
}

func NewClientProtocol(conn net.Conn, maxLength int) *ClientProtocol {
	clientProtocol := &ClientProtocol{
		conn: conn,
		maxLength: maxLength,
	}
	return clientProtocol
}

func (cp *ClientProtocol) serializeBet(bet Bet) string {
    return fmt.Sprintf("%s|%s|%s|%s|%s|%s\n",
        bet.Agency, bet.FirstName, bet.LastName, bet.Document, bet.Birthdate, bet.Number)
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

func (cp *ClientProtocol) sendBatchSize(size int) error {
    buf := make([]byte, SIZE)
    binary.BigEndian.PutUint16(buf, uint16(size))
    return cp.sendAllBytes(buf)
}

func (cp *ClientProtocol) sendBatch(batch []Bet) error {
	//if len(batch) == 0 {
      //  return cp.sendBatchSize(BATCH_FINISHED) 
    //}

	var msg strings.Builder
	for _, bet := range batch {
		msg.WriteString(cp.serializeBet(bet))
	}

	message := msg.String()
	sizeBatch := len(message)

	if sizeBatch > cp.maxLength {
		return fmt.Errorf("message too long")
	}

    if err := cp.sendBatchSize(sizeBatch); err != nil { return err }
    return cp.sendAllMessage(message)
}

func (cp *ClientProtocol) recvOneByte() (int, error) {
    buf := make([]byte, ONE_BYTE)
    total := 0
    for total < ONE_BYTE {
        n, err := cp.conn.Read(buf[total:])
        if err != nil {
            return 0, err
        }
        total += n
    }
    return int(buf[0]), nil
}

func (cp *ClientProtocol) recvResponseBatch() (int, error) {
	batchResponse, err := cp.recvOneByte()
	if err != nil {
		return 0, err
	}
	if batchResponse != BATCH_OK && batchResponse != BATCH_FAIL {
		return 0, fmt.Errorf("unexpected batch response: %d", batchResponse)
	}
	return batchResponse, nil
}
