package main

import (
	"bufio"
	"crypto/rsa"
	"flag"
	"fmt"
	"github.com/RoboCup-SSL/ssl-game-controller/internal/app/client"
	"github.com/RoboCup-SSL/ssl-game-controller/pkg/refproto"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslconn"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

var udpAddress = flag.String("udpAddress", "224.5.23.1:10003", "The multicast address of ssl-game-controller")
var autoDetectAddress = flag.Bool("autoDetectHost", true, "Automatically detect the game-controller host and replace it with the host given in address")
var refBoxAddr = flag.String("address", "localhost:10007", "Address to connect to")
var privateKeyLocation = flag.String("privateKey", "", "A private key to be used to sign messages")
var clientIdentifier = flag.String("identifier", "test", "The identifier of the client")

var privateKey *rsa.PrivateKey

type Client struct {
	conn  net.Conn
	token string
}

func main() {
	flag.Parse()

	privateKey = client.LoadPrivateKey(*privateKeyLocation)

	if *autoDetectAddress {
		log.Print("Trying to detect host based on incoming referee messages...")
		host := client.DetectHost(*udpAddress)
		if host != "" {
			log.Print("Detected game-controller host: ", host)
			*refBoxAddr = client.GetConnectionString(*refBoxAddr, host)
		}
	}

	conn, err := net.Dial("tcp", *refBoxAddr)
	if err != nil {
		log.Fatal("could not connect to game-controller at ", *refBoxAddr)
	}
	defer conn.Close()
	log.Printf("Connected to game-controller at %v", *refBoxAddr)
	c := Client{}
	c.conn = conn

	c.register()

	go func() {
		for {
			time.Sleep(1 * time.Second)
			c.sendEmptyMessage()
		}
	}()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("-> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Print("Can not read from stdin: ", err)
			for {
				time.Sleep(1 * time.Second)
			}
		}
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)

		if strings.Compare("ballLeftField", text) == 0 {
			c.sendBallLeftField()
		} else if strings.Compare("botCrashUnique", text) == 0 {
			c.sendBotCrashUnique()
		} else if strings.Compare("doubleTouch", text) == 0 {
			c.sendDoubleTouch()
		} else {
			fmt.Println("Available commands: ")
			fmt.Printf("  %-20s: %s\n", "help", "Show this help text")
			fmt.Printf("  %-20s: %s\n", "ballLeftField", "Send game event")
			fmt.Printf("  %-20s: %s\n", "botCrashUnique", "Send game event")
			fmt.Printf("  %-20s: %s\n", "doubleTouch", "Send game event")
		}
	}
}

func (c *Client) register() {
	reply := refproto.ControllerToAutoRef{}
	if err := sslconn.ReceiveMessage(c.conn, &reply); err != nil {
		log.Fatal("Failed receiving controller reply: ", err)
	}
	if reply.GetControllerReply() == nil || reply.GetControllerReply().NextToken == nil {
		log.Fatal("Missing next token")
	}

	registration := refproto.AutoRefRegistration{}
	registration.Identifier = clientIdentifier
	if privateKey != nil {
		registration.Signature = &refproto.Signature{Token: reply.GetControllerReply().NextToken, Pkcs1V15: []byte{}}
		registration.Signature.Pkcs1V15 = client.Sign(privateKey, &registration)
	}
	log.Print("Sending registration")
	if err := sslconn.SendMessage(c.conn, &registration); err != nil {
		log.Fatal("Failed sending registration: ", err)
	}
	log.Print("Sent registration, waiting for reply")
	reply = refproto.ControllerToAutoRef{}
	if err := sslconn.ReceiveMessage(c.conn, &reply); err != nil {
		log.Fatal("Failed receiving controller reply: ", err)
	}
	if reply.GetControllerReply().StatusCode == nil || *reply.GetControllerReply().StatusCode != refproto.ControllerReply_OK {
		reason := ""
		if reply.GetControllerReply().Reason != nil {
			reason = *reply.GetControllerReply().Reason
		}
		log.Fatal("Registration rejected: ", reason)
	}
	log.Printf("Successfully registered as %v", *clientIdentifier)
	if reply.GetControllerReply().NextToken != nil {
		c.token = *reply.GetControllerReply().NextToken
	} else {
		c.token = ""
	}
}

func (c *Client) sendBallLeftField() {
	event := refproto.GameEvent_BallLeftFieldTouchLine{}
	event.BallLeftFieldTouchLine = new(refproto.GameEvent_BallLeftField)
	event.BallLeftFieldTouchLine.ByBot = new(uint32)
	*event.BallLeftFieldTouchLine.ByBot = 2
	event.BallLeftFieldTouchLine.ByTeam = new(refproto.Team)
	*event.BallLeftFieldTouchLine.ByTeam = refproto.Team_BLUE
	event.BallLeftFieldTouchLine.Location = &refproto.Location{X: new(float32), Y: new(float32)}
	*event.BallLeftFieldTouchLine.Location.X = 1
	*event.BallLeftFieldTouchLine.Location.Y = 4.5
	gameEvent := refproto.GameEvent{Event: &event, Type: new(refproto.GameEventType)}
	*gameEvent.Type = refproto.GameEventType_BALL_LEFT_FIELD_TOUCH_LINE
	request := refproto.AutoRefToController{GameEvent: &gameEvent}
	c.sendRequest(&request, true)
}

func (c *Client) sendDoubleTouch() {
	event := refproto.GameEvent_AttackerDoubleTouchedBall_{}
	event.AttackerDoubleTouchedBall = new(refproto.GameEvent_AttackerDoubleTouchedBall)
	event.AttackerDoubleTouchedBall.ByBot = new(uint32)
	*event.AttackerDoubleTouchedBall.ByBot = 2
	event.AttackerDoubleTouchedBall.ByTeam = new(refproto.Team)
	*event.AttackerDoubleTouchedBall.ByTeam = refproto.Team_BLUE
	event.AttackerDoubleTouchedBall.Location = &refproto.Location{X: new(float32), Y: new(float32)}
	*event.AttackerDoubleTouchedBall.Location.X = 1
	*event.AttackerDoubleTouchedBall.Location.Y = 4.5
	gameEvent := refproto.GameEvent{Event: &event, Type: new(refproto.GameEventType)}
	*gameEvent.Type = refproto.GameEventType_ATTACKER_DOUBLE_TOUCHED_BALL
	request := refproto.AutoRefToController{GameEvent: &gameEvent}
	c.sendRequest(&request, true)
}

func (c *Client) sendBotCrashUnique() {
	event := refproto.GameEvent_BotCrashUnique_{}
	event.BotCrashUnique = new(refproto.GameEvent_BotCrashUnique)
	event.BotCrashUnique.Violator = new(uint32)
	*event.BotCrashUnique.Violator = 2
	event.BotCrashUnique.Victim = new(uint32)
	*event.BotCrashUnique.Victim = 5
	event.BotCrashUnique.ByTeam = new(refproto.Team)
	*event.BotCrashUnique.ByTeam = refproto.Team_BLUE
	event.BotCrashUnique.Location = &refproto.Location{X: new(float32), Y: new(float32)}
	*event.BotCrashUnique.Location.X = 1
	*event.BotCrashUnique.Location.Y = 4.5
	gameEvent := refproto.GameEvent{Event: &event, Type: new(refproto.GameEventType)}
	*gameEvent.Type = refproto.GameEventType_BOT_CRASH_UNIQUE
	request := refproto.AutoRefToController{GameEvent: &gameEvent}
	c.sendRequest(&request, true)
}

func (c *Client) sendEmptyMessage() {
	request := refproto.AutoRefToController{}
	c.sendRequest(&request, false)
}

func (c *Client) sendRequest(request *refproto.AutoRefToController, doLog bool) {
	if privateKey != nil {
		request.Signature = &refproto.Signature{Token: &c.token, Pkcs1V15: []byte{}}
		request.Signature.Pkcs1V15 = client.Sign(privateKey, request)
	}

	logIf(doLog, "Sending ", request)

	if err := sslconn.SendMessage(c.conn, request); err != nil {
		log.Fatalf("Failed sending request: %v (%v)", request, err)
	}

	logIf(doLog, "Waiting for reply...")
	reply := refproto.ControllerToAutoRef{}
	if err := sslconn.ReceiveMessage(c.conn, &reply); err != nil {
		log.Fatal("Failed receiving controller reply: ", err)
	}
	logIf(doLog, "Received reply: ", reply)
	if reply.GetControllerReply() == nil || reply.GetControllerReply().StatusCode == nil || *reply.GetControllerReply().StatusCode != refproto.ControllerReply_OK {
		log.Fatal("Message rejected: ", *reply.GetControllerReply().Reason)
	}
	if reply.GetControllerReply().NextToken != nil {
		c.token = *reply.GetControllerReply().NextToken
	} else {
		c.token = ""
	}
}

func logIf(doLog bool, v ...interface{}) {
	if doLog {
		log.Print(v...)
	}
}
