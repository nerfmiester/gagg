package main

import (

	"fmt"
	"github.com/nerfmiester/gagg/Config"
	//"github.com/nerfmiester/gagg/Structs"
	"log"
	"github.com/streadway/amqp"
	"github.com/ivpusic/toml"
	"flag"
	"os"

	"net/http"
	"io/ioutil"
)
var tomlConfig Config.TomlConfig

var name = flag.String("name", "World", "A name to say hello to")

var meta bool
var health bool
var url string

var debug bool

var shootme bool

var (
	uri = flag.String("uri", "amqp://guest:guest@localhost:5672/", "AMQP URI")
	exchangeName = flag.String("exchange", "test-exchange", "Durable AMQP exchange name")
	exchangeType = flag.String("exchange-type", "direct", "Exchange type - direct|fanout|topic|x-custom")
	routingKey = flag.String("key", "test-key", "AMQP routing key")
	body = flag.String("body", "foobar", "Body of message")
	reliable = flag.Bool("reliable", true, "Wait for the publisher confirmation before exiting")
	getFlights bool
)


func init() {
	flag.BoolVar(&health, "health", false, "Get meta endpoitnt.")
	flag.BoolVar(&meta, "meta", false, "Get meta endpoitnt.")
	flag.BoolVar(&meta, "m", false, "Geta meta endpoint.")
	flag.BoolVar(&health, "h", false, "Geta meta endpoint.")
	flag.BoolVar(&getFlights, "f", false, "Use json language.")
	flag.BoolVar(&getFlights, "flights", false, "Use TOML language.")
	flag.BoolVar(&debug, "d", false, "Debug mode.")
	flag.BoolVar(&debug, "debug", false, "debug mode.")
	flag.BoolVar(&shootme, "shoot", false, "shoot me.")
	flag.StringVar(&url,"url", "http://example.com/","host url")
}

func main() {

	flag.Parse()
	if (shootme) {
		fmt.Println("Bang..........")
		os.Exit(0)

	}

	fmt.Println("Hello")

	getToml()



	if resp, err := http.Get(url); err != nil {
		failOnError(err, "couldn't get url")
	} else {
		robots, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s", robots)
	}


    if tomlConfig.Agg.MessageSend {


		if err := publishMessage(*uri, *exchangeName, *exchangeType, *routingKey, *body, *reliable); err != nil {
			failOnError(err, "Failed to send message")
		} else {
			fmt.Printf("Result: success\n")
		}
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}


func publishMessage(amqpURI, exchange, exchangeType, routingKey, body string, reliable bool) error {

	// This function dials, connects, declares, publishes, and tears down,
	// all in one go. In a real service, you probably want to maintain a
	// long-lived connection as state, and publish against that.

	log.Printf("dialing %q", amqpURI)
	connection, err := amqp.Dial(amqpURI)
	if err != nil {
		return fmt.Errorf("Dial: %s", err)
	}
	defer connection.Close()

	log.Printf("got Connection, getting Channel")
	channel, err := connection.Channel()
	if err != nil {
		return fmt.Errorf("Channel: %s", err)
	}

	log.Printf("got Channel, declaring %q Exchange (%q)", exchangeType, exchange)
	if err := channel.ExchangeDeclare(
		exchange, // name
		exchangeType, // type
		true, // durable
		false, // auto-deleted
		false, // internal
		false, // noWait
		nil, // arguments
	); err != nil {
		return fmt.Errorf("Exchange Declare: %s", err)
	}

	// Reliable publisher confirms require confirm.select support from the
	// connection.
	if reliable {
		log.Printf("enabling publishing confirms.")
		if err := channel.Confirm(false); err != nil {
			return fmt.Errorf("Channel could not be put into confirm mode: %s", err)
		}

		confirms := channel.NotifyPublish(make(chan amqp.Confirmation, 1))

		defer confirmOne(confirms)
	}

	log.Printf("declared Exchange, publishing %dB body (%q)", len(body), body)
	if err = channel.Publish(
		exchange, // publish to an exchange
		routingKey, // routing to 0 or more queues
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            []byte(body),
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0, // 0-9
			// a bunch of application/implementation-specific fields
		},
	); err != nil {
		return fmt.Errorf("Exchange Publish: %s", err)
	}

	return nil
}
// One would typically keep a channel of publishings, a sequence number, and a
// set of unacknowledged sequence numbers and loop until the publishing channel
// is closed.
func confirmOne(confirms <-chan amqp.Confirmation) {
	log.Printf("waiting for confirmation of one publishing")

	if confirmed := <-confirms; confirmed.Ack {
		log.Printf("confirmed delivery with delivery tag: %d", confirmed.DeliveryTag)
	} else {
		log.Printf("failed delivery of delivery tag: %d", confirmed.DeliveryTag)
	}
}



func getToml() {
	if _, err := toml.DecodeFile("./Config/config.toml", &tomlConfig); err != nil {
		log.Fatal(err)
		// return
	}
	if (tomlConfig.Debug.Enabled) {
		fmt.Printf("Title: %s\n", tomlConfig.Title)
		fmt.Printf("Owner: %s (%s, %s), Born: %s\n",
			tomlConfig.Owner.Name, tomlConfig.Owner.Org, tomlConfig.Owner.Bio,
			tomlConfig.Owner.DOB)

		for serverName, server := range tomlConfig.Servers {
			fmt.Printf("Server: %s (%s, %s)\n", serverName, server.IP, server.DC)
		}
		fmt.Printf("Client data: %v\n", tomlConfig.Clients.Data)
		fmt.Printf("Client hosts: %v\n", tomlConfig.Clients.Hosts)
		fmt.Printf("Agg: %s, \n", tomlConfig.Agg.Server)
		for i, port := range tomlConfig.Agg.Ports {

			fmt.Printf("Port %d is %d\n", i, port)

		}
	}
}
