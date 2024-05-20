package op

import (
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/synctv-org/synctv/internal/model"
	pb "github.com/synctv-org/synctv/proto/message"
)

type Client struct {
	u       *User
	r       *Room
	c       chan Message
	wg      sync.WaitGroup
	conn    *websocket.Conn
	timeOut time.Duration
	closed  uint32
}

func newClient(user *User, room *Room, conn *websocket.Conn) *Client {
	return &Client{
		r:       room,
		u:       user,
		c:       make(chan Message, 128),
		conn:    conn,
		timeOut: 10 * time.Second,
	}
}

func (c *Client) User() *User {
	return c.u
}

func (c *Client) Room() *Room {
	return c.r
}

func (c *Client) Broadcast(msg Message, conf ...BroadcastConf) error {
	return c.r.hub.Broadcast(msg, conf...)
}

func (c *Client) SendChatMessage(message string) error {
	if !c.u.HasRoomPermission(c.r, model.PermissionSendChatMessage) {
		return model.ErrNoPermission
	}
	return c.Broadcast(&pb.ElementMessage{
		Type: pb.ElementMessageType_CHAT_MESSAGE,
		Time: time.Now().UnixMilli(),
		ChatResp: &pb.ChatResp{
			Message: message,
			Sender: &pb.Sender{
				Userid:   c.u.ID,
				Username: c.u.Username,
			},
		},
	})
}

func (c *Client) Send(msg Message) error {
	c.wg.Add(1)
	defer c.wg.Done()
	if c.Closed() {
		return ErrAlreadyClosed
	}
	c.c <- msg
	return nil
}

func (c *Client) Close() error {
	if !atomic.CompareAndSwapUint32(&c.closed, 0, 1) {
		return ErrAlreadyClosed
	}
	c.wg.Wait()
	close(c.c)
	return nil
}

func (c *Client) Closed() bool {
	return atomic.LoadUint32(&c.closed) == 1
}

func (c *Client) GetReadChan() <-chan Message {
	return c.c
}

func (c *Client) NextWriter(messageType int) (io.WriteCloser, error) {
	return c.conn.NextWriter(messageType)
}

func (c *Client) NextReader() (int, io.Reader, error) {
	return c.conn.NextReader()
}

func (c *Client) SetSeekRate(seek float64, rate float64, timeDiff float64) (*Status, error) {
	return c.u.SetRoomCurrentSeekRate(c.r, seek, rate, timeDiff)
}

func (c *Client) SetStatus(playing bool, seek float64, rate float64, timeDiff float64) (*Status, error) {
	return c.u.SetRoomCurrentStatus(c.r, playing, seek, rate, timeDiff)
}
