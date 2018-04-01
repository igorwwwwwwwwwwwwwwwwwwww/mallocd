package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"reflect"
	"syscall"
	"unsafe"
)

var refs = make(map[unsafe.Pointer]interface{})

const (
	MethodMalloc = iota
	MethodFree
	MethodRead
	MethodWrite
)

func handleClient(conn *net.UDPConn, req, rep []byte, fd uintptr, addr syscall.RawSockaddrInet4, addrSize uintptr) error {
	r1, _, e := syscall.Syscall6(syscall.SYS_RECVFROM, fd, uintptr(unsafe.Pointer(&req[0])), uintptr(len(req)), 0, uintptr(unsafe.Pointer(&addr)), uintptr(unsafe.Pointer(&addrSize)))

	if e > 0 {
		return fmt.Errorf("recvfrom syscall returned errno %v %v", e, r1)
	}

	method := binary.BigEndian.Uint64(req[:8])
	switch method {
	case MethodMalloc:
		len := binary.BigEndian.Uint64(req[8:16])

		t := reflect.ArrayOf(int(len), reflect.TypeOf(byte(0)))
		ptr := reflect.New(t).Pointer()
		p := unsafe.Pointer(uintptr(ptr))

		refs[p] = nil

		binary.BigEndian.PutUint64(rep, uint64(ptr))

		r1, _, e := syscall.Syscall6(syscall.SYS_SENDTO, fd,
			uintptr(unsafe.Pointer(&rep[0])), uintptr(8),
			0, uintptr(unsafe.Pointer(&addr)), unsafe.Sizeof(addr))
		if e > 0 {
			return fmt.Errorf("recvfrom syscall returned errno %v %v", e, r1)
		}
	case MethodFree:
		ptr := binary.BigEndian.Uint64(req[8:16])
		p := unsafe.Pointer(uintptr(ptr))

		delete(refs, p)
	case MethodRead:
		ptr := binary.BigEndian.Uint64(req[8:16])
		len := binary.BigEndian.Uint64(req[16:24])

		for i := uint64(0); i < len; i++ {
			rep[i] = *((*byte)(unsafe.Pointer(uintptr(ptr) + uintptr(i))))
		}

		r1, _, e := syscall.Syscall6(syscall.SYS_SENDTO, fd,
			uintptr(unsafe.Pointer(&rep[0])), uintptr(len),
			0, uintptr(unsafe.Pointer(&addr)), unsafe.Sizeof(addr))
		if e > 0 {
			return fmt.Errorf("recvfrom syscall returned errno %v %v", e, r1)
		}
	case MethodWrite:
		ptr := binary.BigEndian.Uint64(req[8:16])
		len := binary.BigEndian.Uint64(req[16:24])
		offset := uint64(24)

		for i := uint64(0); i < len; i++ {
			charp := ((*byte)(unsafe.Pointer(uintptr(ptr) + uintptr(i))))
			*charp = req[offset+i]
		}
	}

	return nil
}

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:8081")
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fd, err := conn.File()
	if err != nil {
		panic(err)
	}

	req := make([]byte, 256)
	rep := make([]byte, 256)

	var addr syscall.RawSockaddrInet4
	addrSize := unsafe.Sizeof(addr)

	for {
		err = handleClient(conn, req, rep, fd.Fd(), addr, addrSize)
		if err != nil {
			panic(err)
		}
	}
}
