package common

import (
	"net"
	"time"
	"os"
	"strings"
	"fmt"
	"bufio"

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

func (c *Client) sendBatch(batch []Bet) error {
	return c.clientProtocol.sendBatch(batch)
}

func (c *Client) recvResponseBatch() (int, error) {
	return c.clientProtocol.recvResponseBatch()
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
func (c *Client) splitBet(line string) (Bet, error) {
	record := strings.Split(line, ",")
	
	if len(record) != 5 {
		log.Errorf("action: load_bets | result: fail | client_id: %v | error: carga de archivo",
			c.config.ID,
		)
		return Bet{}, fmt.Errorf("invalid record length")
	}
	bet := Bet{
		Agency:    c.config.ID,
		FirstName: record[FirstNameIdx],
		LastName:  record[LastNameIdx],
		Document:  record[DocumentIdx],
		Birthdate: record[BirthdateIdx],
		Number:    record[NumberIdx],
	}
	return bet, nil
}

func (c *Client) LoadBatchfromfile(scanner *bufio.Scanner, lastBet Bet) ([]Bet, Bet, error) {
	var batch []Bet
	batchBytes := 0
	if lastBet != (Bet{}) {
		batch = append(batch, lastBet)
		batchBytes += len(c.clientProtocol.serializeBet(lastBet))
	}

	for scanner.Scan() {
		line := scanner.Text()
		bet, err := c.splitBet(line)
		if err != nil {
			log.Errorf("action: load_bets | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return nil, Bet{}, fmt.Errorf("failed to split bet: %v", err)
		}
		betStr := c.clientProtocol.serializeBet(bet)
		betBytes := len(betStr)

		totalBytes := batchBytes + betBytes

		if len(batch) >= c.config.BatchAmount || totalBytes > c.MaxBytesPerBatch {
			return batch, bet, nil
		}

		batch = append(batch, bet)
		batchBytes += betBytes
	}
	if err := scanner.Err(); err != nil {
        return nil, Bet{}, err
    }

    if len(batch) > 0 {
        return batch, Bet{}, nil
    }

    return nil, Bet{}, nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop(done <-chan bool) {

	max_retries := 5
	c.try_connect(max_retries)

	if c.conn == nil {
		log.Errorf("action: connect | result: fail | client_id: %v | error: no se pudo conectar",
			c.config.ID,
		)
		return
	}

	file, err := os.Open(c.config.AgencyPath)
    if err != nil {
        fmt.Printf("Error opening file: %v\n", err)
        return
    }
    defer file.Close() 

    scanner := bufio.NewScanner(file)
	last_bet := Bet{}

	// batches := c.SplitBetsInBatches(bets, c.config.BatchAmount, c.MaxBytesPerBatch)
	for c.keepRunning {

		batch, next_last_bet, err := c.LoadBatchfromfile(scanner, last_bet)
		if err != nil {
			log.Errorf("action: load_bets | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			break
		}
		if batch == nil || len(batch) == 0 {
			if next_last_bet != (Bet{}) {
            	batch = []Bet{next_last_bet}
				next_last_bet = Bet{}
				c.keepRunning = false
        	} else {
				break
			}
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
			last_bet = next_last_bet

			responseStr := ""
			if response == BATCH_OK {
				responseStr = "success"
			} else {
				responseStr = "fail"
			}
			log.Infof("action: receive_message | result: %v | client_id: %v",
				responseStr,
				c.config.ID,
			)
			// Wait a time between sending one message and the next one
			// time.Sleep(c.config.LoopPeriod)
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