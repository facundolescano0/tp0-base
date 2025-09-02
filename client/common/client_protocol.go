package common

import (
    "bufio"
    "fmt"
    "net"
    "strings"   
)

type ClientProtocol struct {
	conn   net.Conn
}

func NewClientProtocol(conn net.Conn) *ClientProtocol {
	client_protocol := &ClientProtocol{
		conn: conn,
	}
	return client_protocol
}

func (cp *ClientProtocol) serializeBet(bet Bet) string {
    return fmt.Sprintf("%s|%s|%s|%s|%s|%s\n",
        bet.Agency, bet.FirstName, bet.LastName, bet.Document, bet.Birthdate, bet.Number)
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


func (cp *ClientProtocol) recvResponseBet() (string, string, error) {
	msg, err := bufio.NewReader(cp.conn).ReadString('\n')
	if err != nil {
		return "", "", err
	}
	msg = strings.TrimSpace(msg)
	parts := strings.Split(msg, "|")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid response format")
	}
	return parts[0], parts[1], nil
}
