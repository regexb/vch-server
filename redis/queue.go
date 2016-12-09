package redis

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/begizi/vch-server/tunnel"
	"github.com/cenkalti/backoff"
	"github.com/garyburd/redigo/redis"
)

const SubscriberRoomName = "VOICE"

type RedisQueue struct {
	pool     *redis.Pool
	receivec tunnel.ReceiveC
}

func NewRedisQueue(address string) (tunnel.Queue, error) {
	q := &RedisQueue{
		pool:     newPool(address),
		receivec: make(tunnel.ReceiveC),
	}

	// ensure redis connection is up
	if err := q.pingRedis(); err != nil {
		return nil, err
	}

	return q, nil
}

func newPool(address string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", address)
			if err != nil {
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func (i *RedisQueue) pingRedis() error {
	return backoff.Retry(func() error {
		con := i.pool.Get()
		defer con.Close()

		_, err := con.Do("PING")
		if err != nil {
			fmt.Printf("[redis] Could not connect to redis server. %v\n", err)
		} else {
			fmt.Println("[redis] Connected to redis server.")
		}

		return err
	}, backoff.NewExponentialBackOff())
}

func marshalMessage(m *tunnel.QueueMessage) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(m)
	return buf.Bytes(), err
}

func unmarshalMessage(b []byte) (*tunnel.QueueMessage, error) {
	m := &tunnel.QueueMessage{}
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(m)
	return m, err
}

func (i *RedisQueue) Close() error {
	return i.pool.Close()
}

func (i RedisQueue) Broadcast(m *tunnel.QueueMessage) error {
	conn := i.pool.Get()
	data, err := marshalMessage(m)
	if err != nil {
		return err
	}

	_, err = conn.Do("PUBLISH", SubscriberRoomName, data)
	return err
}

func (i RedisQueue) Listen() (tunnel.ReceiveC, error) {
	// subscribe and send messages
	c := make(tunnel.ReceiveC)

	conn := i.pool.Get()

	psc := redis.PubSubConn{conn}

	err := psc.Subscribe(SubscriberRoomName)
	if err != nil {
		close(c)
		return c, err
	}

	go func(c tunnel.ReceiveC) {
		defer conn.Close()
		defer close(c)

		for {
			switch v := psc.Receive().(type) {
			case redis.Message:
				msg, err := unmarshalMessage(v.Data)
				if err != nil {
					fmt.Printf("[redis] Failed to decode message. %v\n", err)
					continue
				}
				c <- msg
			case redis.Subscription:
				fmt.Printf("[redis] Subscribed to channel: %s\n", v.Channel)
			case error:
				fmt.Printf("[redis] Error processing messages. Closing channel. %v\n", err)
				return
			default:
				fmt.Printf("[redis] Received unknown message. Ignored: %#v\n", v)
			}
		}
	}(c)

	return c, nil
}
