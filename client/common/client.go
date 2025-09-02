package common

import (
	"net"
	"time"

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
	config        ClientConfig
	conn          net.Conn
	keepRunning   bool
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
		keepRunning: true,
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

func (c *Client) sendBet(client_protocol *ClientProtocol){
	// self, agency: str, first_name: str, last_name: str, document: str, birthdate: str, number: str
	bet := Bet{
		Agency:     c.config.ID,
		FirstName:  c.config.Name,
		LastName:   c.config.Surname,
		Document:   c.config.NID,
		Birthdate:  c.config.Birth,
		Number:     c.config.Number,
	}
	client_protocol.sendBet(bet)
}

func (c *Client) recvResponseBet(client_protocol *ClientProtocol) {
	nid, number, err := client_protocol.recvResponseBet()
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
func (c *Client) StartClientLoop(done <-chan bool) {
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
		if !c.keepRunning {
			break
		}
		select {
		case <-done:
			log.Infof("action: shutdown | result: success | client_id: %v", c.config.ID)
			c.Shutdown()
			return
		default:
			// Create the connection the server in every loop iteration. Send an
			c.createClientSocket()
			client_protocol := NewClientProtocol(c.conn)

			// TODO: Modify the send to avoid short-write
			c.sendBet(client_protocol)

			c.recvResponseBet(client_protocol)

			// Wait a time between sending one message and the next one
			time.Sleep(c.config.LoopPeriod)
	
		}
	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

func (c *Client) Shutdown() {
	c.keepRunning = false
	if c.conn != nil {
		c.conn.Close()
	}
}