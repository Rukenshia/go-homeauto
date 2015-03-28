package main

import (
	"encoding/json"
	"log"
	"net"
)

type Server struct {
	info       net.UDPAddr
	connection *net.UDPConn
	callbacks  []DataReceivedCallback
}
type DataReceivedCallback func([]byte, *net.UDPAddr, error) (Answer, bool)
type Answer struct {
	Status string
	Data   string
}

func CreateServer(host string, port uint) *Server {
	server := Server{}
	server.info = net.UDPAddr{
		Port: int(port),
		IP:   net.ParseIP(host),
	}
	return &server
}

func (server *Server) Start() error {
	conn, err := net.ListenUDP("udp", &server.info)
	server.connection = conn
	return err
}

func (server *Server) Process() {
	buf := make([]byte, 1024)
	var answer []Answer

	defer server.connection.Close()
	for {
		n, client, err := server.connection.ReadFromUDP(buf)
		if n == 0 {
			continue
		}
		answer = answer[:0]
		for _, callback := range server.callbacks {
			retn, handled := callback(buf[:n], client, err)
			if handled {
				answer = append(answer, retn)
			}
		}

		var data []byte
		if len(answer) == 1 {
			data, err = json.Marshal(answer[0])
		} else {
			data, err = json.Marshal(answer)
		}
		if err != nil {
			log.Printf("Could not Marshal Object: '%s'\n", err)
			data = []byte("{ \"Status\": \"error\", \"Data\": \"error json marshal\" }")
		}
		server.connection.WriteToUDP(data, client)
	}
}

func (server *Server) RegisterCallback(callback DataReceivedCallback) {
	server.callbacks = append(server.callbacks, callback)
}
