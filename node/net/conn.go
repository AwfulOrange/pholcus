// 并发安全的连接结构体。
package net

import (
	"net"
	"sync"
)

type Conn struct {
	conn  net.Conn
	mutex *sync.Mutex
	// 用于阻塞通道
	block chan bool
}

func NewConn(conn net.Conn) *Conn {
	return &Conn{
		conn:  conn,
		mutex: &sync.Mutex{},
		block: make(chan bool),
	}
}

// 抢占使用conn
func (self *Conn) Lock() {
	self.mutex.Lock()
}

// 释放conn
func (self *Conn) Unlock() {
	self.mutex.Unlock()
}

func (self *Conn) RemoteAddr() string {
	return self.conn.RemoteAddr().String()
}

func (self *Conn) Close() error {
	// self.Lock()
	err := self.conn.Close()
	// self.Unlock()
	return err
}

func (self *Conn) Write(b []byte) (int, error) {
	// self.Lock()
	write_len, err := self.conn.Write(b)
	// self.Unlock()
	return write_len, err
}

func (self *Conn) Read(b []byte) (int, error) {
	// self.Lock()
	read_len, err := self.conn.Read(b)
	// self.Unlock()
	return read_len, err
}

func (self *Conn) Block() {
	<-self.block
}

func (self *Conn) Unblock() {
	self.block <- true
}

func (self *Conn) Get() net.Conn {
	return self.conn
}
