package socks5

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func ConnectSocks5(conn net.Conn,
	targetAddress string,
	targetPort int16) error {
	// write bytes 5, 1, 0
	io, err := conn.Write([]byte{5, 1, 0})
	if (err != nil) || (io != 3) {
		if err == nil {
			return errors.New("first write size errorW")
		}
		return err
	}
	//read bytes 5, 0
	var b [2]byte
	//conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	io, err = conn.Read(b[:])
	if (err != nil) || (io != 2) {
		if err == nil {
			return errors.New("first read size")
		}
		return err
	}
	if (b[0] != 5) || (b[1] != 0) {
		return errors.New("first read value")
	}

	//write 5, 1, 0, 3, len(host), host[], port(2 bytes)
	buffer := make([]byte, (7 + len(targetAddress)))
	//header
	buffer[0] = 5
	buffer[1] = 1
	buffer[2] = 0
	buffer[3] = 3
	buffer[4] = byte(len(targetAddress))
	//host name
	for x := 0; x < len(targetAddress); x++ {
		xx := targetAddress[x]
		buffer[x+5] = xx
	}
	//port
	binary.BigEndian.PutUint16(buffer[5+len(targetAddress):], uint16(targetPort))
	conn.Write(buffer[:])

	//read bytes 5, 0
	//conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	b[0] = 99
	b[1] = 99
	io, err = conn.Read(b[:]) // <- connect to onion start now
	if (err != nil) || (io != 2) {
		if err == nil {
			return errors.New("second read size")
		}
		return err
	}
	if (b[0] != 5) || (b[1] != 0) {
		return errors.New("second read value")
	}

	return nil
}

func SendViaTor(torPort uint16,
	targetAddress []byte,
	targetPort int16,
	targetPage string,
	jsonDataBytes []byte) ([]byte, error) {

	proxyAddress := "127.0.0.1:" + strconv.Itoa(int(torPort))
	proxyUrl, err := url.Parse("socks5://" + proxyAddress)
	if err != nil {
		return nil, err
	}
	myClient := &http.Client{Timeout: time.Duration(30) * time.Second,
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}

	postData := bytes.NewBuffer(jsonDataBytes)
	resp, err := myClient.Post("http://"+string(targetAddress)+".onion/"+targetPage,
		"application/json",
		postData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body[:], nil
	// //

	// proxyAddress := "127.0.0.1:" + strconv.Itoa(torPort)
	// conn, err := net.Dial("tcp", proxyAddress)
	// if err != nil {
	// 	return "", fmt.Errorf("error dialing: %s", err)
	// }
	// defer conn.Close()

	// err = ConnectSocks5(conn, string(targetAddress)+".onion", targetPort)
	// if err != nil {
	// 	return "", fmt.Errorf("error socks5 connect: %s", err)
	// }

	// request := fmt.Sprintf("POST /%s HTTP/1.0\r\nContent-Length: %d\r\n\r\n%s",
	// 	targetPage,
	// 	len(jsonDataBytes),
	// 	string(jsonDataBytes))
	// fmt.Fprint(conn, request)

	// response, err := bufio.NewReader(conn).ReadString('\n')
	// if err != nil {
	// 	return "", fmt.Errorf("error reading response: %s", err)
	// }

	// return strings.TrimRight(response, "\n"), nil
}
