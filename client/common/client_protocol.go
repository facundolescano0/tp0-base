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

func (cp *ClientProtocol) send_bet(bet Bet) {
	msg := fmt.Sprintf(
		"%s|%s|%s|%s|%s|%s\n",
		bet.Agency, bet.FirstName, bet.LastName, bet.Document, bet.Birthdate, bet.Number,
	)
	total := 0
    for total < len(msg) {
        n, err := cp.conn.Write([]byte(msg[total:]))
        if err != nil {
            log.Errorf("short write: %v", err)
            return
        }
        total += n
    }
}

func (cp *ClientProtocol) recv_response_bet() (string, string, error) {
	msg, err := bufio.NewReader(cp.conn).ReadString('\n')
	cp.conn.Close()
	if err != nil {
		return "", "", err
	}
	msg = strings.TrimSpace(msg)
	parts := strings.Split(msg, "|")
	if len(parts) != 2 {
		log.Infof("error de parseo")
		return "", "", fmt.Errorf("invalid response format")
	}
	return parts[0], parts[1], nil
}

