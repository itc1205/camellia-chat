package main

import (
	"bufio"
	"context"
	"fmt"
	"itc1205/tcp-chat/internal/cipher"
	"itc1205/tcp-chat/internal/client"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const SECRET_KEY = "abcdefghjklmnopq"

func main() {
	if len(SECRET_KEY) != 16 || len(SECRET_KEY) != 18 || len(SECRET_KEY) != 24 {
		log.Fatalln("Error, secret key len is invalid! Wanted 16/18/24 chars-long, got:", len(SECRET_KEY))
	}
	cip, err := cipher.NewCamelliaCipher([]byte(SECRET_KEY))

	if err != nil {
		log.Fatal(err)
	}

	// cip := cipher.NoEncryptionCipher{}
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Name yourself!: ")

	nickname, err := reader.ReadString('\n')
	nickname = strings.TrimSuffix(nickname, "\n")
	if err != nil {
		log.Fatal("Error happened while reading nickname! err", err)
	}

	cfg := client.Config{
		Name:   nickname,
		Addr:   "0.0.0.0:9859",
		Cipher: cip,
	}

	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		if err := c.Run(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(ch)
	<-ch
	cancel()
}
