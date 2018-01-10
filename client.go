package logdna

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type payload struct {
	Lines []Line `json:"lines"`

	lock       sync.RWMutex
	bufferSize uint32
}

// Flush payload
func (p *payload) Flush() {
	p.lock.Lock()
	p.Lines = []Line{}
	p.lock.Unlock()
}

func (p *payload) Write(l Line) {
	p.lock.Lock()
	p.Lines = append(p.Lines, l)
	p.lock.Unlock()
}

func (p *payload) Size() uint32 {
	p.lock.Lock()
	size := len(p.Lines)
	p.lock.Unlock()
	return uint32(size)
}

type Line struct {
	Timestamp int64       `json:"timestamp"`
	Line      string      `json:"line"`
	App       string      `json:"app"`
	Level     string      `json:"level,omitempty"`
	Env       string      `json:"env,omitempty"`
	Meta      interface{} `json:"meta,omitempty"`
}

type Client struct {
	config  Config
	baseURL string
	payload *payload
}

// MustInit return a sdk client initialized with given config
func MustInit(conf Config) *Client {
	var err error
	if err = conf.Validate(); err != nil {
		panic(err)
	}
	if conf.BufferSize == 0 {
		conf.BufferSize = DefaultBufferSize
	}

	c := new(Client)
	c.config = conf
	err = c.constructURL()
	if err != nil {
		panic(err)
	}

	c.payload = new(payload)
	return c
}

// Construct LogDNA ingestion API url
func (c *Client) constructURL() error {
	api, err := url.Parse(LogDNAIngestionAPI)
	if err != nil {
		return err
	}
	api.User = url.User(c.config.APIKey)
	params := url.Values{}
	params.Add("hostname", c.config.Hostname)
	params.Add("mac", c.config.Mac)
	params.Add("ip", c.config.IP)
	params.Add("tags", strings.Join(c.config.Tags, ","))
	api.RawQuery = params.Encode()
	c.baseURL = fmt.Sprintf("%s&now=", api)
	return nil
}

func (c *Client) Write(log string) {
	c.WriteLine(Line{
		Line: log,
	})
}

func (c *Client) WriteLine(line Line) {
	if c.payload.Size() >= c.config.BufferSize {
		err := c.Emit()
		if err != nil {
			log.Printf("Emit log error: %v", err)
		}
	}
	// set required fields with config if is empty
	if line.App == "" {
		line.App = c.config.App
	}
	if line.Timestamp == 0 {
		line.Timestamp = time.Now().UnixNano() / 1000000
	}
	c.payload.Write(line)
}

// Emit all lines in payload
func (c *Client) Emit() error {
	if c.payload.Size() == 0 {
		return nil
	}

	// read payload into body
	c.payload.lock.Lock()
	body, err := json.Marshal(c.payload)
	if err != nil {
		c.payload.lock.Unlock()
		return err
	}
	c.payload.lock.Unlock()
	c.payload.Flush() // flush payload

	var buf bytes.Buffer
	buf.WriteString(c.baseURL)
	buf.WriteString(strconv.FormatInt(time.Now().UnixNano()/1000000, 10))

	resp, err := http.Post(buf.String(), "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// read error once get unexpected HTTP status code
	if resp.StatusCode >= 400 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Read response body error: %v", err)
		} else {
			return fmt.Errorf("API bad request: %s", b)
		}
	}

	return err
}

func (c *Client) Close() error {
	// clean up all logs
	return c.Emit()
}
