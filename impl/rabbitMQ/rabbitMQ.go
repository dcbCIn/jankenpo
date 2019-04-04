package rabbitMQ

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"jankenpo/shared"
	"net"
)

const NAME = "jankenpo/rabbitMQ"

type Client struct {
	connection net.Conn
}

type RabbitMQ struct {
	ip                 string
	port               string
	useJson            bool
	listener           net.Listener
	serverConnection   *amqp.Connection
	channel            *amqp.Channel
	messages           <-chan amqp.Delivery
	initialConnections int
	clients            []Client

	jsonEncoder *json.Encoder
	jsonDecoder *json.Decoder
}

func (rMQ *RabbitMQ) StartServer(ip, port string, useJson bool, initialConnections int) {
	ln, err := net.Listen("tcp", ip+":"+port)
	if err != nil {
		shared.PrintlnError(NAME, "Error while starting rabbitMQ server. Details: ", err)
	}
	rMQ.listener = ln
	rMQ.useJson = useJson
	rMQ.initialConnections = initialConnections
	rMQ.clients = make([]Client, rMQ.initialConnections)
}

func (rMQ *RabbitMQ) StopServer() {
	err := rMQ.listener.Close()
	if err != nil {
		shared.PrintlnError(NAME, "Error while stoping server. Details:", err)
	}
}

func (rMQ *RabbitMQ) ConnectToServer(ip, port string) {
	// connect to server
	conn, err := amqp.Dial("amqp://guest:guest@" + ip + ":" + port + "/")
	shared.FailOnError(NAME, err, "Failed to connect to RabbitMQ")

	rMQ.serverConnection = conn

	ch, err := conn.Channel()
	shared.FailOnError(NAME, err, "Failed to open a channel")

	rMQ.channel = ch
}

func (rMQ *RabbitMQ) WaitForConnection(cliIdx int) (cl *Client) { // TODO if cliIdx >= inicitalConnections => need to append to the slice
	// aceita conexões na porta
	conn, err := rMQ.listener.Accept()
	if err != nil {
		shared.PrintlnError(NAME, "Error while waiting for connection", err)
	}

	cl = &rMQ.clients[cliIdx]

	cl.connection = conn

	if rMQ.useJson {
		// cria um cofificador/decodificador Json
		rMQ.jsonDecoder = json.NewDecoder(conn)
		rMQ.jsonEncoder = json.NewEncoder(conn)
	}

	return cl
}

func (rMQ *RabbitMQ) CloseConnection() {
	err := rMQ.channel.Close()
	if err != nil {
		shared.PrintlnError(NAME, err)
	}
	err = rMQ.serverConnection.Close()
	if err != nil {
		shared.PrintlnError(NAME, err)
	}
}

func (cl *Client) CloseConnection() {
	err := cl.connection.Close()
	if err != nil {
		shared.PrintlnError(NAME, err)
	}
}

func (rMQ *RabbitMQ) CreateQueue(name string) {
	_, err := rMQ.channel.QueueDeclare(
		name,  // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	shared.FailOnError(NAME, err, "Failed to declare a queue")
}

func (rMQ *RabbitMQ) ReadChannel(queueName string) (messages <-chan amqp.Delivery) {
	messages, err := rMQ.channel.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	shared.FailOnError(NAME, err, "Failed to register a consumer")

	//forever := make(chan bool)

	//go func() {
	/*	for d := range msgs {
		log.Printf("Received a message: %s", d.Body)
	}*/
	//}()

	//log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	//<-forever

	return messages
}

func (rMQ *RabbitMQ) ReadOne(queueName string) (message string) {
	if rMQ.messages == nil {
		msgs, err := rMQ.channel.Consume(
			queueName, // queue
			"",        // consumer
			true,      // auto-ack
			false,     // exclusive
			false,     // no-local
			false,     // no-wait
			nil,       // args
		)
		shared.FailOnError(NAME, err, "Failed to register a consumer")

		rMQ.messages = msgs
	}

	d := <-rMQ.messages
	message = string(d.Body)
	//log.Printf("Received a message: %s", message)
	return message
}

func (rMQ *RabbitMQ) Write(queueName, message string) {
	// envia resposta
	err := rMQ.channel.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
	shared.FailOnError(NAME, err, "Failed to publish a message")
}
