package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func run() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("destination-port argument required")
	}

	destPort := os.Args[1]

	host := "host.docker.internal"
	addrs, err := net.LookupHost(host)
	if err != nil {
		if err, ok := err.(*net.DNSError); ok && err.IsNotFound {
			addrs = nil
		} else {
			return err
		}
	}

	if len(addrs) == 0 {
		host = "172.17.0.1"
	}

	port := "80"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	destination := fmt.Sprintf("%s:%s", host, destPort)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				log.Println(err)
				continue
			}
			return err
		}

		go (func() {
			err = copyConn(conn, destination)
			if err != nil {
				log.Println(err)
			}
			conn.Close()
		})()
	}
}

func copyConn(src net.Conn, destination string) error {
	dst, err := net.Dial("tcp", destination)
	if err != nil {
		return err
	}

	done := make(chan struct{})

	go func() {
		io.Copy(dst, src)
		done <- struct{}{}
	}()

	go func() {
		io.Copy(src, dst)
		done <- struct{}{}
	}()

	<-done
	dst.Close()
	src.Close()
	<-done

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
