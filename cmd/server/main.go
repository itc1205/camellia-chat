package main

import (
	"context"
	"itc1205/tcp-chat/internal/cipher"
	"itc1205/tcp-chat/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const SECRET_KEY = "abcdefghjklmnopq"

func main() {

	cip, err := cipher.NewCamelliaCipher([]byte(SECRET_KEY))

	if err != nil {
		log.Fatal(err)
	}

	// cip := cipher.NoEncryptionCipher{}

	cfg := server.Config{
		Port:   9859,
		Cipher: cip,
	}

	srv, err := server.New(cfg)

	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	go srv.Run(ctx)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(ch)

	<-ch
	cancel()
}
