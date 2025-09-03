package common

import (
    "bufio"
    "fmt"
    "net"
    "strings"   
)

const (
	WinnersRequest = "1"
	BatchFinished   = "0"
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

func (cp *ClientProtocol) serializeHeader(betAmount int) string {
    return fmt.Sprintf("%d!", betAmount)
}

func (cp *ClientProtocol) sendAllMessage(message string) error {
	total := 0
    for total < len(message) {
        n, err := cp.conn.Write([]byte(message[total:]))
        if err != nil {
            return err
        }
        total += n
    }
	return nil
}

func (cp *ClientProtocol) sendBet(bet Bet) {
	msg := cp.serializeBet(bet)
	cp.sendAllMessage(msg)
}

func (cp *ClientProtocol) sendBatch(batch []Bet) error {
	header := cp.serializeHeader(len(batch))
	log.Infof("action: send_batch | result: success | batch_size: %d | header: %s", len(batch), header)
	message := header
	for _, bet := range batch {
		message += cp.serializeBet(bet)
	}
	if len(message) > cp.maxLength {
		return fmt.Errorf("message too long")
	}
	message = message + "\n"
	log.Infof("action: send_batch | result: success | batch_size: %d | message: (%s)", len(batch), message)
	return cp.sendAllMessage(message)
}

func (cp *ClientProtocol) recvResponseBet() (string, string, error) {
	msg, err := bufio.NewReader(cp.conn).ReadString('\n')
	if err != nil {
		return "", "", err
	}
	parts := strings.Split(msg, "|")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid response format")
	}
	return parts[0], parts[1], nil
}

func (cp *ClientProtocol) recvResponseBatch() (string, error) {
	msg, err := bufio.NewReader(cp.conn).ReadString('\n')
	if err != nil {
		return "", err
	}
	msg = strings.TrimSpace(msg)
	return msg, nil
}

func (cp *ClientProtocol) sendWinnersRequest() {
	cp.sendAllMessage(fmt.Sprintf("%s\n", WinnersRequest))
}


func (cp *ClientProtocol) recvResponseWinners() ([]string, error) {
	   msg, err := bufio.NewReader(cp.conn).ReadString('\n')
	   if err != nil {
		   return nil, err
	   }
	   response := strings.Split(msg, "|")
	   return response, nil
}

func (cp *ClientProtocol) sendAgencyID(agencyID string) {
	cp.sendAllMessage(fmt.Sprintf("%s\n", agencyID))
}