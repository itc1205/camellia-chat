package client

import (
	"bufio"
	"context"
	"fmt"
	"itc1205/tcp-chat/internal/cipher"
	"log"
	"net"
	"os"
	"strings"
)

type Config struct {
	Name   string
	Addr   string
	Cipher cipher.Cipher
}

type Client struct {
	conn     net.Conn
	username string
	cipher   cipher.Cipher
}

func New(c Config) (*Client, error) {
	conn, err := net.Dial("tcp", c.Addr)
	if err != nil {
		log.Println("Could not estabilish connection with the tcp server:", err)
		return nil, err
	}
	return &Client{conn, c.Name, c.Cipher}, nil
}

func (c *Client) Run(ctx context.Context) error {
	defer func() {
		err := c.conn.Close()
		if err != nil {
			log.Println("Could not close connection with the server:", err)
		}
	}()
	// Write username to the server
	err := c.Write(c.username)
	if err != nil {
		log.Println("Could not send username to the server:", err)
		return err
	}

	go c.RunTX(ctx)
	go c.RunRX(ctx)

	for range ctx.Done() {
		return nil
	}
	return nil
}

func (c *Client) Write(str string) error {
	data := ([]byte)(str)

	alloc_size := len(data)

	if alloc_size < c.cipher.BlockSize() {
		alloc_size = c.cipher.BlockSize()
	}

	if alloc_size%c.cipher.BlockSize() != 0 {
		alloc_size += c.cipher.BlockSize() - alloc_size%c.cipher.BlockSize()
	}

	for len(data) <= alloc_size {
		data = append(data, byte(0))
	}

	// Allocate data for encrypting it
	encrypted_data := make([]byte, alloc_size)
	// Encrypt the data before sending it
	for i := 0; i < len(encrypted_data)/c.cipher.BlockSize(); i++ {
		lower := i * c.cipher.BlockSize()
		upper := (i + 1) * c.cipher.BlockSize()
		c.cipher.Encrypt(encrypted_data[lower:upper], data[lower:upper])
	}

	// Append with null point
	data = append(encrypted_data, byte('\r'))
	if _, err := c.conn.Write(data); err != nil {
		return err
	}
	return nil
}

func (c *Client) Read() (string, error) {
	reader := bufio.NewReader(c.conn)

	data, err := reader.ReadBytes('\r')

	if err != nil {
		return "", err
	}

	// if len(data) < c.cipher.BlockSize() {
	// 	log.Println("What the fucking fuck", data, len(data))
	// }

	for len(data)%c.cipher.BlockSize() != 1 {
		buffer, err := reader.ReadBytes('\r')
		if err != nil {
			return "", err
		}
		data = append(data, buffer...)
	}

	// Remove last '\r' byte
	data = data[:len(data)-1]

	alloc_size := len(data)

	if alloc_size < c.cipher.BlockSize() {
		alloc_size = c.cipher.BlockSize()
	}

	// if alloc_size%c.cipher.BlockSize() != 0 {
	// 	alloc_size += c.cipher.BlockSize() - alloc_size%c.cipher.BlockSize()
	// }

	decrypted_data := make([]byte, alloc_size)
	// Decrypt data in blocks
	for i := 0; i < len(decrypted_data)/c.cipher.BlockSize(); i++ {
		lower := i * c.cipher.BlockSize()
		upper := (i + 1) * c.cipher.BlockSize()
		c.cipher.Decrypt(decrypted_data[lower:upper], data[lower:upper])
	}

	// Trim bytes
	for i := len(decrypted_data) - 1; i > 0 && (decrypted_data[i] == byte(0) || decrypted_data[i] == byte('\r')); i-- {
		decrypted_data = decrypted_data[:i]
	}

	return (string)(decrypted_data), nil

}

// RunTX is a function runner for transmitting data to the server
func (c *Client) RunTX(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			reader := bufio.NewReader(os.Stdin)
			text, err := reader.ReadString('\n')
			text = strings.TrimSuffix(text, "\n")
			if err != nil {
				log.Fatal("Error happened while reading from stdin, err:", err)
			}
			if err := c.Write(text); err != nil {
				log.Fatal("Could not write to tcp connection, err:", err)
			}

		}
	}
}

// RunRX is a function runner for recieving data from the server
func (c *Client) RunRX(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			text, err := c.Read()
			if err != nil {
				log.Fatal("Couldn't read from tcp connection", err)
			}
			fmt.Println(text)
		}

	}
}
