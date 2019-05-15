package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

type ICMP struct {
	Type uint8
	Code uint8
	CheckSum uint16
	Identifier uint16
	SequenceNum uint16
}

func usage()  {
	msg := `
Usage:
		xtcping host, like ping, require root privileges !!!
		xtcping host port, like xtcping

Example: 
		./xtcping www.ifiveplus.com
		./xtcping www.ifiveplus.com 443`
	fmt.Println(msg)
	os.Exit(0)
}

func getICMP(seq uint16) ICMP {
	icmp := ICMP{
		Type: 8,
		Code: 0,
		CheckSum: 0,
		Identifier: 0,
		SequenceNum: seq,
	}
	var buffer bytes.Buffer
	binary.Write(&buffer, binary.BigEndian, icmp)
	icmp.CheckSum = CheckSum(buffer.Bytes())

	return icmp
}

func CheckSum(data[] byte) uint16 {
	var (
		sum uint32
		length int = len(data)
		index int
	)
	for length > 1 {
		sum += uint32(data[index]) << 8 + uint32(data[index + 1])
		index += 2
		length -= 2
	}
	if length > 0 {
		sum += uint32(data[index])
	}
	sum += sum >> 16
	return uint16(^sum)
}

func sendTCPRequest(icmp ICMP, destAddr *net.TCPAddr) error {
	conn, err := net.DialTCP("tcp", nil, destAddr)
	if err != nil {
		// fmt.Printf("Fail to connec to remote host: %s\n", err)
		return err
	}
	defer conn.Close()

	var buffer bytes.Buffer
	binary.Write(&buffer, binary.BigEndian, icmp)

	if _, err := conn.Write(buffer.Bytes()); err != nil {
		return err
	}

	tStart := time.Now()

	conn.SetReadDeadline(time.Now().Add(time.Second * 1))

	recv := make([]byte, 1024)
	receiveCnt, err := conn.Read(recv)
	if err != nil {
		return err
	}

	tEnd := time.Now()
	duration := tEnd.Sub(tStart).Nanoseconds() / 1e6

	fmt.Printf("%d bytes from %s: seq=%d time=%dms\n", receiveCnt, destAddr.String(), icmp.SequenceNum, duration)

	return err
}

func sendICMPRequest(icmp ICMP, destAddr *net.IPAddr) error {
	conn, err := net.DialIP("ip4:icmp", nil, destAddr)
	if err != nil {
		// fmt.Printf("Fail to connec to remote host: %s\n", err)
		return err
	}
	defer conn.Close()

	var buffer bytes.Buffer
	binary.Write(&buffer, binary.BigEndian, icmp)

	if _, err := conn.Write(buffer.Bytes()); err != nil {
		return err
	}

	tStart := time.Now()

	conn.SetReadDeadline(time.Now().Add(time.Second * 1))

	recv := make([]byte, 1024)
	receiveCnt, err := conn.Read(recv)
	if err != nil {
		return err
	}



	tEnd := time.Now()
	duration := tEnd.Sub(tStart).Nanoseconds() / 1e6

	fmt.Printf("%d bytes from %s: seq=%d time=%dms\n", receiveCnt, destAddr.String(), icmp.SequenceNum, duration)

	return err
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	host := os.Args[1]

	ripaddr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		fmt.Printf("Fail to resolve %s, %s\n", host, err)
		return
	}

	fmt.Printf("Ping %s (%s):\n\n", ripaddr.String(), host)

	if len(os.Args) == 2 {
		// ping
		for i := 1; i < 6; i++  {
			if err = sendICMPRequest(getICMP(uint16(i)), ripaddr); err != nil {
				fmt.Printf("Error %s\n", err)
			}
			time.Sleep(1 * time.Second)
		}
	} else {
		// tcping
		port, err := strconv.Atoi(os.Args[2])
		if err != nil {
			usage()
		}
		ip := ripaddr.String()
		addr := fmt.Sprintf("%s:%d", ip, port)
		raddr, err := net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			fmt.Printf("Fail to resolve %s %s, %s\n", ip, port, err)
			return
		}
		// fmt.Printf("Ping %s (%s):\n\n", raddr.String(), host)
		for i := 1; i < 6; i++  {
			if err = sendTCPRequest(getICMP(uint16(i)), raddr); err != nil {
				fmt.Printf("Error %s\n", err)
			}
			time.Sleep(1 * time.Second)
		}
	}

}
