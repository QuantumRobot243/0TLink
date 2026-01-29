package main

import (
	"github.com/hashicorp/yamux" //yamux is aa streaam multiplexing over singel TCP connection
	"io"
	"log"
	"net"        //TCP networking
)

fnc main() {
	clintLn, _ := net.Listen("tcp", ":7000") //Opens TCP listeneer on port 7000 For local clint
	log.Println("Tunnel Control Listining on :7000")
	
	conn, _ :=clintLn.Accept() //Initialize the Yamux connection
	session, _ := yamux.Server(conn, nil)
// yamux.server will open or accept streams && The session can now create many logical connections on single tcp socket
	publicLn, _ := net.Listen("tcp", ":8000")
	log.Println("Public traffic on :8000") //Anyone  who will connect here will be forward through the tunnel
	
	for {
		userconn, _ := publicLn.Accept() //each user conn is a new external clint
	
		stream, _ := session.Open()    //A new virtual connection inside of the yamux sessionn the stream goes to the agent amnd no new tcp connection is created 
		//Meaning Every user = one yamux stream
		go func() {      //Bidirectional Piping
			defer userConn.Close()
			defer stream.Close()
			go io.Copy(userConn, stream) //Agent -> User (Read data from yamux stream and write to public user)
			io.Copy(stream,userConn)     //user -> agent (Reads from public user sends to yanux stream)
		}()
	}

}
