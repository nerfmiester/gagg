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
	"time"
	"strconv"

	"math/rand"
"errors"
	"net/url"
)
var tomlConfig Config.TomlConfig

var name = flag.String("name", "World", "A name to say hello to")

var meta bool
var health bool
var address string

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
var arrivalDate string
var departureDate string
var from string
var to string
var date time.Time
var port string
var Url *url.URL

var channel string

var channelMap = make(map[string]string)

func init() {
	date = time.Now()
	flag.BoolVar(&health, "health", false, "Get meta endpoitnt.")
	flag.BoolVar(&meta, "meta", false, "Get meta endpoitnt.")
	flag.BoolVar(&meta, "m", false, "Geta meta endpoint.")
	flag.BoolVar(&health, "h", false, "Geta meta endpoint.")
	flag.BoolVar(&getFlights, "f", false, "Use json language.")
	flag.BoolVar(&getFlights, "flights", false, "Use TOML language.")
	flag.BoolVar(&debug, "d", false, "Debug mode.")
	flag.BoolVar(&debug, "debug", false, "debug mode.")
	flag.BoolVar(&shootme, "shoot", false, "shoot me.")
	flag.StringVar(&address,"url", "http://public-api.agg.pre1.gb.tsm.internal","host url")
	flag.StringVar(&port,"p", "8080","Host port to use, default is non prod envs.")
	flag.StringVar(&port,"port", "8080","Host port to use, default is non prod envs.")
	flag.StringVar(&departureDate,"dd", "2006-01-02","Departure date")
	flag.StringVar(&arrivalDate,"ad", "2006-01-02","Arival date")
	flag.StringVar(&from,"from", "MAN","From IATA code")
	flag.StringVar(&to,"to", "CPH","To IATA code")
	flag.StringVar(&channel, "c" , "f", "channel f=flights, c=carhire, h=hotels, l=holidays")
	if (departureDate=="2006-01-02") {
		departureDate = date.Format("2006-01-02")
	}
	rand.Seed(time.Now().UnixNano())

	channelMap["f"]="flights"
	channelMap["c"]="carhire"
	channelMap["h"]="hotels"
	channelMap["l"]="holidays"



}

func main() {


	flag.Parse()
	if (shootme) {



		fmt.Println("Bang..........")
		fmt.Printf("Date as string........%s-%d-%s\n", strconv.Itoa(date.Year())  ,date.Month(),strconv.Itoa(date.Day()))
		fmt.Printf("From is .........%s\n", departureDate)
		if (arrivalDate=="2006-01-02") {
			fourteenDays := time.Hour * 24 * 14
			parsedDate, err := time.Parse("2006-01-02", departureDate)
			if err != nil {
				fmt.Println(err)
			}
			arrivalDate = parsedDate.Add(fourteenDays).Format("2006-01-02")
		}
		fmt.Printf("To is .........%s\n", arrivalDate)

		fmt.Printf("A random user is ......... %s\n", RandStringBytes(18))
        err, channelString := channelLookup(channel)
		if (err != nil) {
			fmt.Printf("Your channel value  .........%s , is bullshit use f=flights, c=carhire, h=hotels, l=holidays\n", channel)
			os.Exit(1)
		}


		Url, err = url.Parse(address + ":" + port + "/")
		if err != nil {
			panic("boom")
		}
		//curl http://localhost:8080/gb/flights/v1/search/MAN/2015-11-22/DUB/2015-11-29/2/0/0/e/0/false\?return=true\&userId=hgjfkdshgdfs\&source=TIV
		Url.Path +=  "gb/" + channelString + "v1/search/" + from + "/" + departureDate + "/" + to +"/" + arrivalDate + "/2/0/0/e/0/false"
		parameters := url.Values{}
		parameters.Add("return", "true")
		parameters.Add("userId", RandStringBytes(18))
		parameters.Add("source", "TIV")
		Url.RawQuery = parameters.Encode()

		fmt.Printf("Encoded URL is %q\n", Url.String())

		if resp, err := http.Get(Url.String()); err != nil {
			failOnError(err, "couldn't get url")
		} else {
			robots, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%s", robots)
		}


		os.Exit(0)

	}

	fmt.Println("Hello")

	getToml()



	if resp, err := http.Get(Url.String()); err != nil {
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

func RandStringBytes(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func channelLookup(channel string) (err error, channelstring string) {


	if _, ok := channelMap[channel] ; ok == true {
		return nil, channelMap[channel]
	} else {
		return errors.New("No strings supplied"),""
	}




}
