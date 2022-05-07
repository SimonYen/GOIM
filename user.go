package main

import (
	"net"
)

type User struct {
	Name       string
	Address    string
	Channel    chan string
	connection net.Conn
}

func NewUser(conn net.Conn) *User {
	user := &User{
		Name:       conn.RemoteAddr().String(),
		Address:    conn.RemoteAddr().String(),
		Channel:    make(chan string),
		connection: conn,
	}
	go user.ListenMessage()
	return user
}

//监听管道，一有回复马上发送
func (user *User) ListenMessage() {
	for {
		msg := <-user.Channel
		user.connection.Write([]byte(msg + "\n"))
	}
}
