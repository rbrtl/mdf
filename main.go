package main

import (
	"bufio"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

func main() {
	log.Println("initialising")

	log.Println("listen for system interrupt")
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	errorUnset("MD_HOST")
	errorUnset("MD_PORT")
	addr := os.ExpandEnv("ws://${MD_HOST}:${MD_PORT}")
	log.Printf("connecting to websocket at address: %s", addr)

	wsClient, res, err := websocket.DefaultDialer.Dial(addr, http.Header{})
	if err != nil {
		log.Fatal("err:", res.Status, err)
	}
	defer wsClient.Close()

	log.Printf("len:res:%d", res.ContentLength)
	if res.StatusCode != 101 {
		log.Fatalf("HTTP%s", res.Status)
	}

	broker(wsClient)

	// Wait for interrupt signal to gracefully close the connection
	select {
	case <-interrupt:
		log.Println("Interrupt received, closing connection...")
		err := wsClient.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "close"))
		if err != nil {
			log.Println("write close:", err)
			return
		}
		select {
		case <-time.After(time.Second):
		}
	}

	log.Println("Exiting.")
}

func broker(wsClient *websocket.Conn) (chan<- string, <-chan string) {
	t := make(chan<- string)
	go func(chan<- string) {
		for {
			if _, p, err := wsClient.ReadMessage(); err != nil {
				log.Fatalf("%v", p)
			} else {
				log.Println(string(p))
			}
		}
	}(t)

	q := make(<-chan string)
	go func(<-chan string) {
		stdin := bufio.NewScanner(os.Stdin)
		for stdin.Scan() {

			lineReader := strings.NewReader(stdin.Text())
			line := make([]byte, lineReader.Len())
			if _, err := lineReader.Read(line); err != nil {
				log.Fatal(err)
			}

			if err := wsClient.WriteMessage(websocket.TextMessage, line); err != nil {
				log.Fatal(err)
			}

		}
	}(q)

	return t, q
}

func errorUnset(env string) {
	if _, ok := os.LookupEnv(env); !ok {
		log.Fatalf("%s unset", env)
	}
}
