package main

import (
	"encoding/json"
	"fmt"
	"jankenpo/shared"
	"net"
	"os"
	"strconv"
	"sync"
)

func inArray(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func ProcessaSolicitacao(request shared.Request) int {
	possibilities := []string{"P", "A", "T"}

	if !inArray(request.Player1, possibilities) {
		return -1
	}
	if !inArray(request.Player2, possibilities) {
		return -1
	}
	if request.Player1 == request.Player2 {
		return 0
	}
	switch request.Player1 {
	case "P":
		if request.Player2 == "A" {
			return 2
		} else {
			return 1
		}
	case "A":
		if request.Player2 == "P" {
			return 1
		} else {
			return 2
		}
	case "T":
		if request.Player2 == "P" {
			return 2
		} else {
			return 1
		}
	default:
		return -1
	}
}

func StartJankenpoServer() {
	fmt.Println("Initializing server")
	// escuta na porta tcp configurada
	ln, _ := net.Listen("tcp", ":"+strconv.Itoa(shared.SERVER_PORT))

	// aceita conexões na porta
	conn, err := ln.Accept()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// fecha o socket
	defer func() {
		err := conn.Close()
		if err != nil {
			panic(err)
		}
	}()

	// cria um cofificador/decodificador Json
	jsonDecoder := json.NewDecoder(conn)
	jsonEncoder := json.NewEncoder(conn)
	var msgFromClient shared.Request

	fmt.Println("Servidor pronto para receber solicitações (TCP)...")
	for idx := 0; idx < shared.SAMPLE_SIZE; idx++ {

		// recebe solicitações do cliente e decodifica-as
		err = jsonDecoder.Decode(&msgFromClient)
		if err != nil {
			fmt.Println(err)
			return
		}

		// processa a solicitação
		r := ProcessaSolicitacao(msgFromClient)

		// envia resposta ao cliente
		msgToClient := shared.Reply{r}
		err = jsonEncoder.Encode(msgToClient)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func main() {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		StartJankenpoServer()
		wg.Done()
	}()

	wg.Wait()
}