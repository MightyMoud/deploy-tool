package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func doStuff() {
	server := "206.189.83.57:22"

	// connect to local ssh-agent to grab all keys
	sshAgentSock := os.Getenv("SSH_AUTH_SOCK")
	if sshAgentSock == "" {
		log.Fatal("No SSH SOCK AVAIBALEB")
	}
	// make a connection to SSH agent over unix protocl
	conn, err := net.Dial("unix", sshAgentSock)
	if err != nil {
		log.Fatalf("Failed to connect to SSH agent: %s", err)
	}
	defer conn.Close()

	// make a ssh agent out of the connection
	agentClient := agent.NewClient(conn)

	// Check that we can get all the public keys added to the agent properly
	_, signersErr := agentClient.Signers()
	if signersErr != nil {
		log.Fatalf("Failed to get signers from SSH agent: %v", err)
	}

	// now that we have our key, we need to start ssh client sesssion
	// ƒirst we make some config we pass later
	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			// passing the public keys to callback to get the auth methods
			ssh.PublicKeysCallback(agentClient.Signers),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// create SSH client with the said config and connect to server
	client, sshClientErr := ssh.Dial("tcp", server, config)
	if sshClientErr != nil {
		log.Fatalf("Failed to dial: %s", err)
	}
	defer client.Close()

	// create a session of that client
	sshSession, sshSessErr := client.NewSession()
	if sshSessErr != nil {
		log.Fatalf("Failed to create session: %s", err)
	}
	defer sshSession.Close()

	// Need to hook into the pipe of output coming from that session
	// Hook a reader into the pipe and read then write out to our output here
	sshReader, err := sshSession.StdoutPipe()
	if err != nil {
		fmt.Printf("Something went wrong getting the reader: /%s", err)
	}
	// make a scanner of that reader that will read as we get new stuff
	scanner := bufio.NewScanner(sshReader)
	// start a separate go routine to read from the pipe and print out
	go func() {
		for scanner.Scan() {
			fmt.Printf("%s\n", scanner.Text())
		}
	}()

	// Start the session with a command
	sshSession.Start("apt update")
	// wait for command to exit -> block to let the go routine also send back the text
	sshSession.Wait()

}
