package server

import (
	"fmt"
	"github.com/dcbCIn/jankenpo/impl/rabbitMQ"
	"github.com/dcbCIn/jankenpo/shared"
	"os"
	"strconv"
	"sync"
)

const NAME = "jankenpo/rabbitMQ/server"

func waitForConection(rMQ rabbitMQ.RabbitMQ, idx int) {
	shared.PrintlnInfo(NAME, "Connection", strconv.Itoa(idx), "started")
	var wg sync.WaitGroup
	wg.Add(1)

	rMQ.CreateQueue("moves")
	rMQ.CreateQueue("result")

	var msgFromClient shared.Request

	shared.PrintlnInfo(NAME, "Servidor pronto para receber solicitações (rabbitMQ)")

	messages := rMQ.ReadChannel("moves")
	go func() {
		i := 0
		for d := range messages {
			message := string(d.Body)

			shared.PrintlnInfo(NAME, "Message received: ", message)
			_, err := fmt.Sscanf(message, "%s %s", &msgFromClient.Player1, &msgFromClient.Player2)
			if err != nil {
				shared.PrintlnError(NAME, err)
				os.Exit(1)
			}

			// processa a solicitação
			r := shared.ProcessaSolicitacao(msgFromClient)

			// envia resposta ao cliente
			rMQ.Write("result", strconv.Itoa(r))

			i++
			if i >= shared.SAMPLE_SIZE {
				shared.PrintlnInfo(NAME, "Atingida quantidade de Sample Size, finalizando servidor!")
				break
			}
		}
		wg.Done()
	}()

	wg.Wait()
	shared.PrintlnInfo(NAME, "Servidor finalizado (rabbitMQ)")
	shared.PrintlnInfo(NAME, "Connection", strconv.Itoa(idx), "ended")
}

func StartJankenpoServer() {
	var wg sync.WaitGroup
	shared.PrintlnInfo(NAME, "Initializing server rabbitMQ")

	// escuta na porta rabbitMQ configurada
	var rMQ rabbitMQ.RabbitMQ
	rMQ.ConnectToServer("localhost", strconv.Itoa(shared.RABBITMQ_PORT))
	defer rMQ.CloseConnection()

	for idx := 0; idx < shared.CONECTIONS; idx++ {
		wg.Add(1)
		go func(i int) {
			waitForConection(rMQ, i)

			wg.Done()
		}(idx)
	}
	wg.Wait()
	shared.PrintlnInfo(NAME, "Fim do Servidor rabbitMQ")
}
