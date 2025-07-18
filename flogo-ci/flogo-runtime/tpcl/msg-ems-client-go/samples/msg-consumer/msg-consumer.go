/*
 * Copyright (c) 2001-2023 TIBCO Software Inc.
 * All Rights Reserved. Confidential & Proprietary.
 * For more information, please contact:
 * TIBCO Software Inc., Palo Alto, California, USA
 */

/*
 * This is a simple sample of a basic msg-consumer.
 *
 * This sample subscribes to specified destination and
 * receives and prints all received messages.
 *
 * Notice that the specified destination should exist in your configuration
 * or your topics/queues configuration file should allow
 * creation of the specified destination.
 *
 * If this sample is used to receive messages published by the
 * msg-producer sample, it must be started prior
 * to running the msg-producer sample.
 *
 * Usage:  msg-consumer [options]
 *
 *    where options are:
 *
 *      -server     Server URL.
 *                  If not specified this sample assumes a
 *                  serverUrl of "tcp://localhost:7222"
 *
 *      -user       User name. Default is null.
 *      -password   User password. Default is null.
 *      -topic      Topic name. Default is "topic.sample"
 *      -queue      Queue name. No default
 *      -ackmode    Session acknowledge mode. Default is AUTO.
 *                  Other values: DUPS_OK, CLIENT, EXPLICIT_CLIENT,
 *                  EXPLICIT_CLIENT_DUPS_OK and NO.
 *
 */

package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/tibco/msg-ems-client-go/samples"
	"github.com/tibco/msg-ems-client-go/tibems"
	"os"
	"strings"
)

func main() {
	var err error

	flag.Usage = func() {
		fmt.Println(
			`
Usage: msg-consumer [options] [tls options]

   where options are:

   -server   <server URL>  - EMS server URL, default is local server
   -user     <user name>   - user name, default is null
   -password <password>    - password, default is null
   -topic    <topic-name>  - topic name, default is "topic.sample"
   -queue    <queue-name>  - queue name, no default
   -ackmode  <ack-mode>    - acknowledge mode, default is AUTO
                                               other modes: CLIENT, DUPS_OK, NO,
                                               EXPLICIT_CLIENT and EXPLICIT_CLIENT_DUPS_OK
   -help-ssl               - help on tls parameters`)
	}

	server := flag.String("server", "tcp://localhost:7222", "EMS server URL, default is local server")
	user := flag.String("user", "", "user name, default is null")
	password := flag.String("password", "", "password, default is null")
	topic := flag.String("topic", "topic.sample", "topic name, default is \"topic.sample\"")
	queue := flag.String("queue", "", "queue name, no default")
	ackmodeStr := flag.String("ackmode", "AUTO", "acknowledge mode")
	helpSsl := flag.Bool("help-ssl", false, "help on tls parameters")

	flag.Parse()

	if *helpSsl {
		fmt.Printf("%s", samples.SslUsage)
		os.Exit(0)
	}

	destinationName := topic
	if *queue != "" {
		destinationName = queue
	}

	fmt.Printf("------------------------------------------------------------------------\n"+
		"msg-consumer SAMPLE\n"+
		"------------------------------------------------------------------------\n"+
		"Server....................... %s\n"+
		"User......................... %s\n"+
		"Destination.................. %s\n",
		*server, *user, *destinationName)
	fmt.Printf("------------------------------------------------------------------------\n")

	var connection *tibems.Connection

	if strings.HasPrefix(*server, "ssl") {
		sslParams, pkeyPassword, err := samples.CreateSSLParams()
		if err != nil {
			fmt.Printf("Error creating SSL parameters: %v\n", err)
			os.Exit(1)
		}
		connection, err = tibems.CreateConnection(*server, "", *user, *password,
			&tibems.ConnectionOptions{
				SSLParams:          sslParams,
				PrivateKeyPassword: pkeyPassword,
			})
		if err != nil {
			sslParams.Close()
			fmt.Printf("Error creating EMS server connection: %v\n", err)
			os.Exit(1)
		}
		sslParams.Close()
	} else {
		connection, err = tibems.CreateConnection(*server, "", *user, *password, nil)
		if err != nil {
			fmt.Printf("Error creating EMS server connection: %v\n", err)
			os.Exit(1)
		}
	}

	err = connection.SetExceptionListener(func(connection *tibems.Connection, err error) {
		if errors.Is(err, tibems.ErrNotConnected) {
			fmt.Println("CONNECTION EXCEPTION: Server Disconnected")
		}
	})
	if err != nil {
		fmt.Printf("Error setting exception handler: %v\n", err)
		os.Exit(1)
	}

	ackMode := tibems.AcknowledgeMode(tibems.AckModeAutoAcknowledge)
	switch *ackmodeStr {
	case "AUTO":
		ackMode = tibems.AckModeAutoAcknowledge
	case "CLIENT":
		ackMode = tibems.AckModeClientAcknowledge
	case "DUPS_OK":
		ackMode = tibems.AckModeDupsOkAcknowledge
	case "NO":
		ackMode = tibems.AckModeNoAcknowledge
	case "EXPLICIT_CLIENT":
		ackMode = tibems.AckModeExplicitClientAcknowledge
	case "EXPLICIT_CLIENT_DUPS_OK":
		ackMode = tibems.AckModeExplicitClientDupsOkAcknowledge
	default:
		fmt.Printf("Unknown or unsupported ACK mode '%s'\n", *ackmodeStr)
		os.Exit(1)
	}

	session, err := connection.CreateSession(false, ackMode)
	if err != nil {
		fmt.Printf("Error creating session: %v\n", err)
		os.Exit(1)
	}

	var destination *tibems.Destination
	if *queue != "" {
		destination, err = tibems.CreateDestination(tibems.DestTypeQueue, *queue)
	} else {
		destination, err = tibems.CreateDestination(tibems.DestTypeTopic, *topic)
	}
	if err != nil {
		fmt.Printf("Error creating destination: %v\n", err)
		os.Exit(1)
	}

	consumer, err := session.CreateConsumer(destination, "", false)
	if err != nil {
		fmt.Printf("Error creating consumer: %v\n", err)
		os.Exit(1)
	}

	err = connection.Start()
	if err != nil {
		fmt.Printf("Error starting connection: %v\n", err)
		os.Exit(1)
	}

	for {
		msg, err := consumer.Receive()
		if err != nil {
			fmt.Printf("Error on receive: %v\n", err)
		}
		if msg == nil {
			break
		}
		if ackMode == tibems.AckModeClientAcknowledge ||
			ackMode == tibems.AckModeExplicitClientAcknowledge ||
			ackMode == tibems.AckModeExplicitClientDupsOkAcknowledge {
			err = msg.Acknowledge()
			if err != nil {
				fmt.Printf("Error acknowledging message: %v\n", err)
			}
		}
		messageTypeName := "UNKNOWN"
		bodyType, err := msg.GetBodyType()
		if err != nil {
			fmt.Printf("Error getting message body type: %v\n", err)
		}
		switch bodyType {
		case tibems.MsgTypeUnknown:
			messageTypeName = "UNKNOWN"
		case tibems.MsgTypeMessage:
			messageTypeName = "MESSAGE"
		case tibems.MsgTypeBytesMessage:
			messageTypeName = "BYTES"
		case tibems.MsgTypeObjectMessage:
			messageTypeName = "OBJECT"
		case tibems.MsgTypeStreamMessage:
			messageTypeName = "STREAM"
		case tibems.MsgTypeMapMessage:
			messageTypeName = "MAP"
		case tibems.MsgTypeTextMessage:
			messageTypeName = "TEXT"
		}

		if bodyType != tibems.MsgTypeTextMessage {
			fmt.Printf("Received %s message:\n", messageTypeName)
			msg.Print()
		} else {
			text, err := msg.(*tibems.TextMsg).GetText()
			if err != nil {
				fmt.Printf("Error getting message text: %v\n", err)
			}
			if text == "" {
				text = "<text is set to NULL>"
			}
			fmt.Printf("Received TEXT message: %s\n", text)
		}

		switch specificMsg := msg.(type) {
		case *tibems.TextMsg:
			specificMsg.Close()
		case *tibems.MapMsg:
			specificMsg.Close()
		case *tibems.Msg:
			specificMsg.Close()
		}
	}

	connection.Stop()
	consumer.Close()
	destination.Close()
	session.Close()
	connection.Close()
}
