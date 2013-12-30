// package kernctl provides access to the OSX Kext Control API for sending and
// receiving messages from kernel extensions.
package kernctl

// #include <stdlib.h>
// #include <sys/socket.h>
// #include <sys/kern_control.h>
// #include <string.h>
import "C"

import (
	"fmt"
	"syscall"
	"unicode/utf8"
	"unsafe"
)

const (
	SYSPROTO_CONTROL = 2
	AF_SYSTEM        = 32
	PF_SYSTEM        = AF_SYSTEM
	AF_SYS_CONTROL   = 2
	CTLIOCGINFO      = 3227799043
)

type Conn struct {
	CtlId  uint32
	UnitId uint32
	fd     int
}

type Message interface {
	Bytes() []byte
}

func (conn *Conn) socket() (int, error) {
	if conn.fd == 0 {
		fd, err := syscall.Socket(PF_SYSTEM, syscall.SOCK_DGRAM, SYSPROTO_CONTROL)
		if err != nil {
			return 0, err
		}
		conn.fd = fd
	}

	return conn.fd, nil
}

// Connect will create a connection to the control socket for the
// kernel extension named in CtlName
func (conn *Conn) Connect() error {
	_, errno := conn.connect()
	var err error = nil
	if errno != 0 {
		err = fmt.Errorf("failed to connect to kext. errno: ", errno)
	}

	return err
}

// Close closes a connection to a kernel extension
func (conn *Conn) Close() {
	if conn.fd != 0 {
		syscall.Close(conn.fd)
	}
	conn.fd = 0
}

func (conn *Conn) SendCommand(msg Message) {
	fd, err := conn.socket()
	n, err := syscall.Write(fd, msg.Bytes()[:])
}

func (conn *Conn) Select(bufSize int) (error, []byte) {
	fd, _ := conn.socket()

	timeout := &syscall.Timeval{
		Sec:  1,
		Usec: 0,
	}

	r := &syscall.FdSet{}
	FD_ZERO(r)
	FD_SET(r, fd)

	syscall.Select(fd+1, r, nil, nil, timeout)

	bytesRead := 0
	readBuf := make([]byte, bufSize)

	if FD_ISSET(r, fd) {
		n, _, err := syscall.Recvfrom(fd, readBuf, 0)
		bytesRead = n

		if err != nil {
			return err, nil
		}
	}

	if bytesRead > 0 {
		return nil, readBuf[0:bytesRead]
	} else {
		return nil, nil
	}
}

func (conn *Conn) createSockAddr() C.struct_sockaddr_ctl {
	var sockaddr C.struct_sockaddr_ctl
	sockaddr.sc_len = C.u_char(unsafe.Sizeof(C.struct_sockaddr_ctl{}))
	sockaddr.sc_family = C.u_char(PF_SYSTEM)
	sockaddr.ss_sysaddr = C.u_int16_t(AF_SYS_CONTROL)
	sockaddr.sc_id = C.u_int32_t(conn.CtlId)
	sockaddr.sc_unit = C.u_int32_t(conn.UnitId)
	return sockaddr
}

func (conn *Conn) connect() (ret int64, err syscall.Errno) {
	sockLen := 32
	sa := conn.createSockAddr()
	fd, _ := conn.socket()
	r1, r2, e := syscall.Syscall(syscall.SYS_CONNECT, uintptr(fd),
		uintptr(unsafe.Pointer(&sa)),
		uintptr(sockLen))
	return int64(r1), e
}

// Create a new connection to a named kext's kernel control socket
func NewConnByName(CtlName string) *Conn {
	conn := new(Conn)
	fd, _ := conn.socket()
	conn.CtlId, _ = GetCtlId(fd, CtlName)
	return conn
}

func NewConnByCtlId(CtlId uint32, UnitId uint32) *Conn {
	conn := new(Conn)
	conn.CtlId = CtlId
	conn.UnitId = UnitId
	return conn
}

// GetCtlId retrieves the kext control id for the kext named in CtlName using
// the socket file descriptor fd.
func GetCtlId(fd int, CtlName string) (uint32, error) {
	var info C.struct_ctl_info
	info.ctl_id = 0
	C.memcpy(unsafe.Pointer(&info.ctl_name), unsafe.Pointer(C.CString(CtlName)),
		C.size_t(utf8.RuneCountInString(CtlName)))
	syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), CTLIOCGINFO,
		uintptr(unsafe.Pointer(&info)))
	return uint32(info.ctl_id), nil
}

// stolen from https://github.com/pebbe/zmq2
func FD_SET(p *syscall.FdSet, i int) {
	p.Bits[i/64] |= 1 << uint(i) % 64
}

func FD_ISSET(p *syscall.FdSet, i int) bool {
	return (p.Bits[i/64] & (1 << uint(i) % 64)) != 0
}

func FD_ZERO(p *syscall.FdSet) {
	for i := range p.Bits {
		p.Bits[i] = 0
	}
}
