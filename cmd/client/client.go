package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	MethodMalloc = iota
	MethodFree
	MethodRead
	MethodWrite
)

type client struct {
	conn *net.UDPConn
	req  []byte
	ptr  []byte
}

func connect() (*client, error) {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:8081")
	if err != nil {
		return nil, err
	}

	laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", laddr, addr)
	if err != nil {
		return nil, err
	}

	return &client{
		conn: conn,
		req:  make([]byte, 256),
		ptr:  make([]byte, 8),
	}, nil
}

func (c *client) close() error {
	return c.conn.Close()
}

func (c *client) malloc(len uint64) (uint64, error) {
	binary.BigEndian.PutUint64(c.req, MethodMalloc)
	binary.BigEndian.PutUint64(c.req[8:], len)
	_, err := c.conn.Write(c.req[:16])
	if err != nil {
		return 0, err
	}

	_, _, err = c.conn.ReadFromUDP(c.ptr)
	if err != nil {
		return 0, err
	}

	ptr := binary.BigEndian.Uint64(c.ptr)
	return ptr, nil
}

func (c *client) free(ptr uint64) error {
	binary.BigEndian.PutUint64(c.req, MethodFree)
	binary.BigEndian.PutUint64(c.req[8:], ptr)
	_, err := c.conn.Write(c.req[:16])
	return err
}

func (c *client) read(ptr, len uint64, rep []byte) error {
	binary.BigEndian.PutUint64(c.req, MethodRead)
	binary.BigEndian.PutUint64(c.req[8:], ptr)
	binary.BigEndian.PutUint64(c.req[16:], len)
	_, err := c.conn.Write(c.req[:24])
	if err != nil {
		return err
	}

	_, _, err = c.conn.ReadFromUDP(rep)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) write(ptr, len uint64, data []byte) error {
	req := make([]byte, 24)
	binary.BigEndian.PutUint64(req, MethodWrite)
	binary.BigEndian.PutUint64(req[8:], ptr)
	binary.BigEndian.PutUint64(req[16:], len)
	req = append(req, data...)

	_, err := c.conn.Write(req[:len+24])
	return err
}

// optimized write() with pre-allocated byte buffer
// expects data at b[:24]
func (c *client) writeRaw(ptr, len uint64, req []byte) error {
	binary.BigEndian.PutUint64(req, MethodWrite)
	binary.BigEndian.PutUint64(req[8:], ptr)
	binary.BigEndian.PutUint64(req[16:], len)

	_, err := c.conn.Write(req[:len+24])
	return err
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage:")
		fmt.Println("  client malloc <len>")
		fmt.Println("  client read <ptr> <len>")
		fmt.Println("  client write <ptr> <len> <data>")
		fmt.Println("  client free <ptr>")
		os.Exit(1)
	}

	c, _ := connect()
	defer c.close()

	switch os.Args[1] {
	case "malloc":
		ptr, _ := c.malloc(20)
		fmt.Printf("%d\n", ptr)
	case "free":
		ptr, _ := strconv.ParseUint(os.Args[2], 10, 64)
		c.free(ptr)
	case "read":
		ptr, _ := strconv.ParseUint(os.Args[2], 10, 64)
		len, _ := strconv.ParseUint(os.Args[3], 10, 64)
		rep := make([]byte, len)
		c.read(ptr, len, rep)
		fmt.Printf("%s\n", rep)
	case "write":
		ptr, _ := strconv.ParseUint(os.Args[2], 10, 64)
		len, _ := strconv.ParseUint(os.Args[3], 10, 64)
		c.write(ptr, len, []byte(strings.Join(os.Args[4:], " ")))
	}
}
