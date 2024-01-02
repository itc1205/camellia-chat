package server

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"itc1205/tcp-chat/internal/cipher"
	"log"
	"net"
	"strings"
)

type Config struct {
	Port   uint16
	Cipher cipher.Cipher
}

type Server struct {
	listener net.Listener
	cipher   cipher.Cipher
}

type Client struct {
	conn net.Conn
}

func New(cfg Config) (*Server, error) {
	log.Printf("Starting server on address: 0.0.0.0:%v...", cfg.Port)
	listner, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%v", cfg.Port))
	if err != nil {
		return nil, err
	}
	return &Server{listner, cfg.Cipher}, nil
}

func (s *Server) Run(ctx context.Context) {
	clients := make([]Client, 0)

	go func() {
		<-ctx.Done()
		for _, client := range clients {
			err := client.conn.Close()

			if err != nil {
				log.Println("Could not close connection with client. Err:", err)
			}
		}
		err := s.listener.Close()

		if err != nil {
			log.Println("Error closing connection")
		}
	}()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Println("Failed to accept connection, error:", err)
			continue
		}
		clients = append(clients, Client{conn})
		go s.HandleConnection(ctx, conn, clients)
	}
}

func (s *Server) broadcastMessage(text string, clients []Client) {
	for _, client := range clients {
		if client.conn == nil {
			continue
		}
		if err := s.Write(client.conn, text); err != nil {
			// log.Println("Failed to write back to the client, removing it from clients! err:", err)
			client.conn.Close()
			client.conn = nil
		}
	}
}

func (s *Server) Write(c net.Conn, str string) error {
	data := ([]byte)(str)

	alloc_size := len(data)

	// If data size is lesser than block size, then we just append it to match
	if alloc_size < s.cipher.BlockSize() {
		alloc_size = s.cipher.BlockSize()
	}
	// Find how much we need to allocate
	if alloc_size%s.cipher.BlockSize() != 0 {
		alloc_size += s.cipher.BlockSize() - alloc_size%s.cipher.BlockSize()
	}
	// Fill block gap
	for len(data) <= alloc_size {
		data = append(data, byte(0))
	}

	// Allocate data for encrypting it
	encrypted_data := make([]byte, alloc_size)

	// Encrypt the data before sending it
	for i := 0; i < len(encrypted_data)/s.cipher.BlockSize(); i++ {
		lower := i * s.cipher.BlockSize()
		upper := (i + 1) * s.cipher.BlockSize()
		s.cipher.Encrypt(encrypted_data[lower:upper], data[lower:upper])
	}

	// Append with null point
	data = append(encrypted_data, byte('\r'))
	if _, err := c.Write(data); err != nil {
		return err
	}
	return nil
}

func (s *Server) Read(c net.Conn) (string, error) {
	reader := bufio.NewReader(c)
	data, err := reader.ReadBytes('\r')

	if err != nil {
		return "", err
	}

	for len(data)%s.cipher.BlockSize() != 1 {
		buffer, err := reader.ReadBytes('\r')
		if err != nil {
			return "", err
		}
		data = append(data, buffer...)
	}

	// Remove last '\r' byte
	data = data[:len(data)-1]

	alloc_size := len(data)

	// if alloc_size < s.cipher.BlockSize() {
	// 	alloc_size = s.cipher.BlockSize()
	// }

	// if alloc_size%s.cipher.BlockSize() != 0 {
	// 	alloc_size += s.cipher.BlockSize() - alloc_size%s.cipher.BlockSize()
	// }

	// for len(data) < alloc_size {
	// 	data = append(data, byte(0))
	// }

	// Allocate data for encrypting it
	decrypted_data := make([]byte, alloc_size)
	// Decrypt data in blocks
	for i := 0; i < len(decrypted_data)/s.cipher.BlockSize(); i++ {
		lower := i * s.cipher.BlockSize()
		upper := (i + 1) * s.cipher.BlockSize()
		s.cipher.Decrypt(decrypted_data[lower:upper], data[lower:upper])
	}
	// Trim null bytes
	for i := len(decrypted_data) - 1; i > 0 && (decrypted_data[i] == byte(0) || decrypted_data[i] == byte('\r')); i-- {
		decrypted_data = decrypted_data[:i]
	}

	return (string)(decrypted_data), nil
}

func (s *Server) HandleConnection(ctx context.Context, conn net.Conn, clients []Client) {

	defer func() {
		err := conn.Close()
		if err != nil {
			log.Println("Error closing connection")
		}
	}()

	nickname, err := s.Read(conn)

	nickname = strings.TrimSuffix(nickname, "\r")

	if err != nil {
		log.Println("Failed to read nickname! err:", err)
	}
	for {

		select {
		case <-ctx.Done():
			return
		default:
			text, err := s.Read(conn)
			if err != nil {
				if err != io.EOF {
					log.Println("Failed to read data! err:", err)
				}
				return
			}

			line := fmt.Sprintf("[%s]: %s", nickname, text)
			log.Printf("%s\n", line)

			s.broadcastMessage(line, clients)
		}
	}
}
