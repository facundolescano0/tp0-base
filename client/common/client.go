package common

import (
	"net"
	"time"
	"os"
    "os/signal"
    "syscall"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

type Bet struct {
    Agency     string
    FirstName  string
    LastName   string
    Document   string
    Birthdate  string
    Number     string
}

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
	Name          string
	Surname       string
	NID           string
	Birth         string
	Number        string
}


// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

func (c *Client) send_bet(client_protocol *ClientProtocol){
	// self, agency: str, first_name: str, last_name: str, document: str, birthdate: str, number: str
	bet := Bet{
		Agency:     c.config.ID,
		FirstName:  c.config.Name,
		LastName:   c.config.Surname,
		Document:   c.config.NID,
		Birthdate:  c.config.Birth,
		Number:     c.config.Number,
	}
	client_protocol.send_bet(bet)
}

func (c *Client) recv_response_bet(client_protocol *ClientProtocol) {
	nid, number, err := client_protocol.recv_response_bet()
	if err!= nil {
		log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
	}	
	
	log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
		nid,
		number,
	)
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
    done := make(chan bool, 1)

	go func() {
		<-sigs
		log.Infof("action: shutdown | result: success | client_id: %v", c.config.ID)
		done <- true
	}()

	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
		select {
		case <-done:
			c.conn.Close()
			return
		default:
			// Create the connection the server in every loop iteration. Send an
			c.createClientSocket()
			client_protocol := NewClientProtocol(c.conn)

			// TODO: Modify the send to avoid short-write
			c.send_bet(client_protocol)

			c.recv_response_bet(client_protocol)

			// Wait a time between sending one message and the next one
			time.Sleep(c.config.LoopPeriod)
	
		}
	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}
