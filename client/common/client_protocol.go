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
   
    // client to server
    BATCH = 1
    WINNERS_REQUEST = 2

    // server to client
    BATCH_OK = 3
    BATCH_FAIL = 4
    SEND_WINNERS = 5

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

func (cp *ClientProtocol) sendOpCode(opCode int) error {
    buf := []byte{byte(opCode)}
    return cp.sendAllBytes(buf)
}

func (cp *ClientProtocol) sendBatchSize(size int) error {
    buf := make([]byte, 2)
    binary.BigEndian.PutUint16(buf, uint16(size))
    return cp.sendAllBytes(buf)
}

func (cp *ClientProtocol) sendBatch(batch []Bet) error {
	if len(batch) == 0 {
        if err := cp.sendOpCode(BATCH); err != nil {
            return err
        }
        return cp.sendBatchSize(BATCH_FINISHED) 
    }
	
	var msg strings.Builder
	for _, bet := range batch {
		msg.WriteString(cp.serializeBet(bet))
	}

	message := msg.String()
	sizeBatch := len(message)

	if sizeBatch > cp.maxLength {
		return fmt.Errorf("message too long")
	}

    if err := cp.sendOpCode(BATCH); err != nil { return err }
    if err := cp.sendBatchSize(sizeBatch); err != nil { return err }
    return cp.sendAllMessage(message)
}

func (cp *ClientProtocol) sendAgencyId(id int) error {
    buf := []byte{byte(id)}
    return cp.sendAllBytes(buf)
}

func (cp *ClientProtocol) sendWinnersRequest(id int) error {
	if err := cp.sendOpCode(WINNERS_REQUEST); err != nil {
		return err
	}
	return cp.sendAgencyId(id)
}

func (cp *ClientProtocol) recvOpCode() (int, error) {
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
	batchResponse, err := cp.recvOpCode()
	if err != nil {
		return 0, err
	}
	return batchResponse, nil
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

func (cp *ClientProtocol) recvResponseWinners() ([]string, error) {
	recvOpCode, err := cp.recvOpCode()
	if err != nil {
		return nil, err
	}
	if recvOpCode != SEND_WINNERS {
		return nil, fmt.Errorf("unexpected opcode: %d", recvOpCode)
	}

	size, err := cp.recvSizeMsg()
    if err != nil {
        return nil, err
    }

	if size < 0 || size > cp.maxLength {
        return nil, fmt.Errorf("invalid winners size: %d", size)
    }

    buf := make([]byte, size)
    total := 0
    for total < size {
        n, err := cp.conn.Read(buf[total:])
        if err != nil { return nil, err }
        total += n
    }

    payload := string(buf)
    // payload = strings.TrimSuffix(payload, "|")
    parts := strings.Split(payload, "|")
	
	if len(parts) == 1 && parts[0] == "" {
        return []string{}, nil
    }

    return parts, nil
}