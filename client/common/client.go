package common

import (
	"net"
	"time"
	"encoding/csv"
	"io"
	"os"

	"github.com/op/go-logging"
)

const (
    FirstNameIdx = iota
    LastNameIdx
    DocumentIdx
    BirthdateIdx
    NumberIdx
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
	BatchAmount   int
	Name          string
	Surname       string
	NID           string
	Birth         string
	Number        string
	AgencyPath    string
}


// Client Entity that encapsulates how
type Client struct {
	config        ClientConfig
	conn          net.Conn
	keepRunning   bool
	clientProtocol *ClientProtocol
	MaxBytesPerBatch int
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
		keepRunning: true,
		MaxBytesPerBatch: 8192,
	}
	return client
}


func (c *Client) LoadBetsFromFile(path string) ([]Bet, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var bets []Bet
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(record) != 5 {
			log.Errorf("action: load_bets | result: fail | client_id: %v | error: carga de archivo",
				c.config.ID,
			)
			continue
		}
		bet := Bet{
			Agency:    c.config.ID,
			FirstName: record[FirstNameIdx],
			LastName:  record[LastNameIdx],
			Document:  record[DocumentIdx],
			Birthdate: record[BirthdateIdx],
			Number:    record[NumberIdx],
		}
		bets = append(bets, bet)
	}
	return bets, nil
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

func (c *Client) sendBet(){
	bet := Bet{
		Agency:     c.config.ID,
		FirstName:  c.config.Name,
		LastName:   c.config.Surname,
		Document:   c.config.NID,
		Birthdate:  c.config.Birth,
		Number:     c.config.Number,
	}
	c.clientProtocol.sendBet(bet)
}

func (c *Client)SplitBetsInBatches(bets []Bet, maxAmount int, maxBytes int) [][]Bet {
	var batches [][]Bet
	var currentBatch []Bet
	currentBatchBytes := 0

	for _, bet := range bets {
		betStr := c.clientProtocol.serializeBet(bet)
		betBytes := len(betStr)

		headerBytes := len(c.clientProtocol.serializeHeader(len(currentBatch) + 1))
		totalBytes := headerBytes + currentBatchBytes + betBytes

		if len(currentBatch) >= maxAmount || totalBytes > maxBytes {
			if len(currentBatch) > 0 {
				batches = append(batches, currentBatch)
			}
			currentBatch = []Bet{}
			currentBatchBytes = 0
		}

		currentBatch = append(currentBatch, bet)
		currentBatchBytes += betBytes
	}

	if len(currentBatch) > 0 {
		batches = append(batches, currentBatch)
	}
	return batches
}

func (c *Client) sendBatch(batch []Bet) error {
	return c.clientProtocol.sendBatch(batch)
}

func (c *Client) recvResponseBatch() (string, error) {
	return c.clientProtocol.recvResponseBatch()
}

func (c *Client) recvResponseBet() {
	nid, number, err := c.clientProtocol.recvResponseBet()
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

func (c *Client) try_connect(max_retries int) {
	for i := 0; i < max_retries; i++ {
		c.createClientSocket()
		if c.conn != nil {
			c.clientProtocol = NewClientProtocol(c.conn, c.MaxBytesPerBatch)
			break
		}
	}
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop(done <-chan bool) {

	bets, err := c.LoadBetsFromFile(c.config.AgencyPath)
	if err != nil {
		log.Errorf("action: load_bets | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	c.try_connect(max_retries=5)

	if c.conn == nil {
		log.Errorf("action: connect | result: fail | client_id: %v | error: no se pudo conectar",
			c.config.ID,
		)
		return
	}

	batches := c.SplitBetsInBatches(bets, c.config.BatchAmount, c.MaxBytesPerBatch)

	for _, batch := range batches {
		if !c.keepRunning {
			break
		}
		select {
		case <-done:
			log.Infof("action: shutdown | result: success | client_id: %v", c.config.ID)
			c.Shutdown()
			return
		default:

			err = c.sendBatch(batch)
			if err != nil {
				log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
					c.config.ID,
					err,
				)
				c.keepRunning = false
				break
			}
			 response, err := c.recvResponseBatch()
			//_, err := c.recvResponseBatch()
			if err != nil {
				log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
					c.config.ID,
					err,
				)
				c.keepRunning = false
				break
			}
			log.Infof("action: receive_message | result: %v | client_id: %v",
			response,
			c.config.ID,
			)
			// Wait a time between sending one message and the next one
			time.Sleep(c.config.LoopPeriod)
		}
	}

	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
	c.Shutdown()
}

func (c *Client) Shutdown() {
	c.keepRunning = false
	if c.conn != nil {
		c.conn.Close()
	}
}