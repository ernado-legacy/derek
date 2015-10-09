package derek

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"log"
	"time"

	"github.com/nats-io/nats"
)

const (
	defaultTimeout = time.Second * 60
)

// Client is remote http client
type Client interface {
	Do(Request) (Response, error)
}

// HTTPClient is absrtaction under http client, that can Do requests
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// NATSAgent implements Agent via NATS subscription
type NATSAgent struct {
	conn       *nats.Conn
	httpClient HTTPClient
}

// NATSClient
type NATSClient struct {
	Prefix string
	Conn *nats.Conn
	Timeout time.Duration
}

func (c NATSClient) Do(req Request) (res Response, err error) {
	data, err := json.Marshal(req)
	if err != nil {
		return res, err
	}
	log.Println("request", c.subj())
	m, err := c.Conn.Request(c.subj(), data, c.Timeout)
	if err != nil {
		return res, err
	}
	return res, json.Unmarshal(m.Data, &res)
}

func (c NATSClient) subj() string {
	return c.Prefix
}

func NewNATSClient(subjPrefix string, conn *nats.Conn) Client {
	client := NATSClient{}
	client.Conn = conn
	client.Timeout = defaultTimeout
	client.Prefix = subjPrefix
	return client
}

func (n NATSAgent) reply(subj string, resp Response) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return n.conn.Publish(subj, data)
}

func (n NATSAgent) replyError(subj string, resp Response, err error) error {
	resp.Error = err
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return n.conn.Publish(subj, data)
}

func (n NATSAgent) handler(msg *nats.Msg) {
	var (
		req  = new(Request)
		res  = new(Response)
		err  error
		hReq *http.Request
		hRes *http.Response
	)
	defer func() {
		if err := n.replyError(msg.Reply, *res, err); err != nil {
			log.Println("nats reply:", err)
		}
	}()
	if err = json.Unmarshal(msg.Data, req); err != nil {
		return
	}
	hReq, err = req.Get()
	if err != nil {
		return
	}
	hRes, err = n.httpClient.Do(hReq)
	if err != nil {
		return
	}
	res, err = NewResponse(hRes)
}

// NewNATSAgent subscribes for subject in queue and returns agent
func NewNATSAgent(subj, queue string, conn *nats.Conn, client HTTPClient) (NATSAgent, error) {
	h := NATSAgent{}
	h.httpClient = client
	h.conn = conn
	_, err := conn.QueueSubscribe(subj, queue, h.handler)
	return h, err
}

// Request is serializable http request
type Request struct {
	Method string
	URL    string
	Body   []byte
}

func (r Request) Get() (req *http.Request, err error) {
	return http.NewRequest(r.Method, r.URL, bytes.NewBuffer(r.Body))
}

// NewRequest serializes http.Request
func NewRequest(req *http.Request) (r Request, err error) {
	r.Method = req.Method
	r.URL = req.URL.String()
	if req.Body == nil {
		return r, nil
	}
	r.Body, err = ioutil.ReadAll(req.Body)
	if err != nil {
		return r, err
	}
	return r, nil
}

// Response is serializable http response
type Response struct {
	Status     string // e.g. "200 OK"
	StatusCode int    // e.g. 200
	Proto      string // e.g. "HTTP/1.0"
	ProtoMajor int    // e.g. 1
	ProtoMinor int    // e.g. 0

	Header        http.Header
	Body          []byte
	ContentLength int64

	TransferEncoding []string
	Close            bool
	Trailer          http.Header
	Error            error
}

func (r Response) Get() (res *http.Response, err error) {
	res = new(http.Response)
	res.Status = r.Status
	res.StatusCode = r.StatusCode
	res.Proto = r.Proto
	res.ProtoMajor = r.ProtoMajor
	res.ProtoMinor = r.ProtoMinor
	res.Header = r.Header
	res.Body = ioutil.NopCloser(bytes.NewBuffer(r.Body))
	res.ContentLength = r.ContentLength
	res.TransferEncoding = r.TransferEncoding
	res.Trailer = r.Trailer
	res.Request = nil
	return res, r.Error
}

// NewResponse serializes http.Response
func NewResponse(r *http.Response) (res *Response, err error) {
	res = new(Response)
	res.Status = r.Status
	res.StatusCode = r.StatusCode
	res.Proto = r.Proto
	res.ProtoMajor = r.ProtoMajor
	res.ProtoMinor = r.ProtoMinor
	res.Header = r.Header
	res.ContentLength = r.ContentLength
	res.TransferEncoding = r.TransferEncoding
	res.Trailer = r.Trailer
	res.Body, err = ioutil.ReadAll(r.Body)
	return res, err
}
