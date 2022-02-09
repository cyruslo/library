package comet

import (
	"git.huoys.com/middle-end/library/pkg/net/comet/internal/bufio"
	"git.huoys.com/middle-end/library/pkg/net/comet/proto"
	"github.com/bilibili/kratos/pkg/log"
	gb "github.com/gogo/protobuf/proto"
	"net"
	"time"
)

type RespMsgHandle func(data []byte, code int32)
type PushMsgHandle func(data []byte)

type ClientConfig struct {
	Addr         string
	PushHandlers map[int32]PushMsgHandle
	RespHandlers map[int32]RespMsgHandle
	Token string
}

type Client struct {
	pushChan     chan *proto.Payload
	closeChan    chan bool
	pushHandlers map[int32]PushMsgHandle
	respHandlers map[int32]RespMsgHandle
}

func NewTcpClient(conf *ClientConfig) (c *Client, err error) {
	c = &Client{
		pushChan:     make(chan *proto.Payload, 100),
		closeChan:    make(chan bool),
		pushHandlers: conf.PushHandlers,
		respHandlers: conf.RespHandlers,
	}
	conn, err := net.Dial("tcp", conf.Addr)
	if err != nil {
		log.Error("net.Dial(%s) error(%v)", conf.Addr, err)
		return
	}
	wr := bufio.NewWriter(conn)
	rd := bufio.NewReader(conn)
	go c.handles(conn, rd)
	go c.dispatch(conn, wr)
	if conf.Token != "" {
		c.auth(conf.Token)
	}
	go c.sendHeart()
	go func() {
		for {
			cc := <-c.closeChan
			log.Info("client close conn")
			if cc {
				if err := conn.Close(); err != nil {
					log.Error("close err %v", err)
				}
			}
		}
	}()
	return
}

func (c *Client)auth(token string) (err error) {
	p := &proto.Payload{
		Type:                 int32(proto.Request),
		Body:                 []byte(token),
	}
	c.pushChan <- p
	return
}

func (c *Client) Request(command int32, msg gb.Message) (err error) {
	var data []byte
	if data, err = gb.Marshal(msg); err != nil {
		return
	}
	body := &proto.Body{
		PlayerId: 0,
		Ops:      command,
		Data:     data,
	}
	var pData []byte
	if pData, err = gb.Marshal(body); err != nil {
		return
	}
	p := &proto.Payload{
		Place: 1,
		Type:  int32(proto.Request),
		Body:  pData,
	}
	c.pushChan <- p
	return
}

func (c *Client) Close() {
	c.closeChan <- true
}

func (c *Client) sendHeart() {
	for {
		p := &proto.Payload{
			Place: 1,
			Type:  int32(proto.Ping),
		}
		c.pushChan <- p
		time.Sleep(time.Second * 5)
	}
}

func (c *Client) handles(conn net.Conn, rd *bufio.Reader) {
	for {
		p := &proto.Payload{}
		if err := p.ReadTCP(rd); err != nil {
			log.Error("ReadTCP err %v", err)
			c.closeChan <- true
			break
		}
		if p.Body == nil {
			continue
		}
		body := &proto.Body{}
		if err := gb.Unmarshal(p.Body, body); err != nil {
			log.Error("proto type %d Unmarshal err %v", p.Type, err)
			continue
		}
		if p.Type == int32(proto.Response) {
			handle, ok := c.respHandlers[body.Ops]
			if !ok {
				log.Error("respHandlers ops %d func is not exist", body.Ops)
				continue
			}
			handle(body.Data, p.Code)
		} else if p.Type == int32(proto.Push) {
			handle, ok := c.pushHandlers[body.Ops]
			if !ok {
				log.Error("pushHandlers ops %d func is not exist", body.Ops)
				continue
			}
			handle(body.Data)
		}
	}
}

func (c *Client) dispatch(conn net.Conn, wr *bufio.Writer) {
	seq := int32(0)
	for {
		p := <-c.pushChan
		p.Seq = seq
		if p.Type == int32(proto.Ping) {
			if err := p.WriteTCPHeart(wr); err != nil {
				log.Error("WriteTCPHeart err %v", err)
				c.closeChan <- true
				break
			}
		} else {
			if err := p.WriteTCP(wr); err != nil {
				log.Error("WriteTCP err %v", err)
				c.closeChan <- true
				break
			}
		}
		if err := wr.Flush(); err != nil {
			log.Error("Flush error(%v)", err)
			c.closeChan <- true
			break
		}
		seq += 1
	}
}
