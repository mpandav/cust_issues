/*
 * Copyright (c) 2001-2023 TIBCO Software Inc.
 * All Rights Reserved. Confidential & Proprietary.
 * For more information, please contact:
 * TIBCO Software Inc., Palo Alto, California, USA
 */

/*
 * This is a simple sample of a basic EMS message producer.
 *
 * This samples publishes specified message(s) on a specified
 * destination and quits.
 *
 * Notice that the specified destination should exist in your configuration
 * or your topics/queues configuration file should allow
 * creation of the specified topic. Sample configuration supplied with
 * the TIBCO Enterprise for EMS distribution allows creation of any
 * destination.
 *
 * If this sample is used to publish messages into the
 * msg-consumer sample, the msg-consumer
 * sample must be started first.
 *
 * If -topic is not specified this sample will use a topic named
 * "topic.sample".
 *
 * Usage:  msg-producer  [options]
 *                               <message-text1>
 *                               ...
 *                               <message-textN>
 *
 *  where options are:
 *
 *   -server    <server-url>  Server URL.
 *                            If not specified this sample assumes a
 *                            serverUrl of "tcp://localhost:7222"
 *   -user      <user-name>   User name. Default is null.
 *   -password  <password>    User password. Default is null.
 *   -topic     <topic-name>  Topic name. Default value is "topic.sample"
 *   -queue     <queue-name>  Queue name. No default
 *   -async                   Send asynchronously, default is false.
 *
 */

package main

import (
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
Usage: msg-producer [options] [tls options]
                            <message-text-1>
                           [<message-text-2>] ...

   where options are:

   -server   <server URL>  - EMS server URL, default is local server
   -user     <user name>   - user name, default is null
   -password <password>    - password, default is null
   -topic    <topic-name>  - topic name, default is "topic.sample"
   -queue    <queue-name>  - queue name, no default
   -async                  - send asynchronously, default is false
   -help-ssl               - help on tls parameters`)
	}

	server := flag.String("server", "tcp://localhost:7222", "EMS server URL, default is local server")
	user := flag.String("user", "", "user name, default is null")
	password := flag.String("password", "", "password, default is null")
	topic := flag.String("topic", "topic.sample", "topic name, default is \"topic.sample\"")
	queue := flag.String("queue", "", "queue name, no default")
	async := flag.Bool("async", false, "send asynchronously, default is false")
	helpSsl := flag.Bool("help-ssl", false, "help on tls parameters")

	flag.Parse()

	if *helpSsl {
		fmt.Printf("%s", samples.SslUsage)
		os.Exit(0)
	}

	messages := make([]string, 0)
	flagsDone := false
	for i := 1; i < len(os.Args); i++ {
		if !flagsDone {
			if os.Args[i] == "--" {
				flagsDone = true
				continue
			} else if !strings.HasPrefix(os.Args[i], "-") {
				flagsDone = true
			}
		}

		if flagsDone {
			messages = append(messages, os.Args[i])
		}
	}
	if len(messages) == 0 {
		fmt.Println("Must specify at least one message to send")
		os.Exit(1)
	}

	destinationName := topic
	if *queue != "" {
		destinationName = queue
	}

	fmt.Printf("------------------------------------------------------------------------\n"+
		"msg-producer SAMPLE\n"+
		"------------------------------------------------------------------------\n"+
		"Server....................... %s\n"+
		"User......................... %s\n"+
		"Destination.................. %s\n"+
		"Send Asynchronously.......... %t\n"+
		"Message Text................. \n",
		*server, *user, *destinationName, *async)

	for i := 0; i < len(messages); i++ {
		fmt.Printf("\t%s \n", messages[i])
	}

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

	session, err := connection.CreateSession(false, tibems.AckModeAutoAcknowledge)
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

	producer, err := session.CreateProducer(destination)
	if err != nil {
		fmt.Printf("Error creating producer: %v\n", err)
		os.Exit(1)
	}

	for i := 0; i < len(messages); i++ {
		msg, err := tibems.CreateTextMsg()
		if err != nil {
			fmt.Printf("Error creating message: %v\n", err)
			os.Exit(1)
		}
		err = msg.SetText(messages[i])
		if err != nil {
			fmt.Printf("Error setting message text: %v\n", err)
			os.Exit(1)
		}
		if *async {
			err = producer.AsyncSend(msg, func(msg tibems.Message, err error) {
				textMsg := msg.(*tibems.TextMsg)
				text, err := textMsg.GetText()
				if err != nil {
					fmt.Println("Error retrieving message!")
					return
				}

				if err == nil {
					fmt.Printf("Sucessfully sent message '%s'\n", text)
				} else {
					fmt.Printf("Error sending message '%s'.\n", text)
					fmt.Printf("Error: %v.\n", err)
				}
				textMsg.Close()
			}, nil)
			if err != nil {
				fmt.Printf("Error sending message: %v\n", err)
			}
		} else {
			err = producer.Send(msg, nil)
			if err != nil {
				fmt.Printf("Error sending message: %v\n", err)
			}
			msg.Close()
		}

		if err == nil {
			fmt.Printf("Published message: %s\n", messages[i])
		}
	}

	producer.Close()
	destination.Close()
	session.Close()
	connection.Close()
}
