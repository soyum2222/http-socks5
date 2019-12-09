package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"http-socks5/socks5-client"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func main() {

	socks_server := flag.String("s", "127.0.0.1:1080", "socks server addr")
	local_port := flag.String("p", "10800", "local listen port")
	help := flag.Bool("help", false, "")

	flag.Parse()

	if *help {
		fmt.Println("-s   socks server addr default 127.0.0.1:1080")
		fmt.Println("-p   local listen port default 10800")
		os.Exit(0)
	}

	l, err := net.Listen("tcp", ":"+*local_port)
	if err != nil {
		panic(err)
	}

	for {

		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		go func() {

			defer conn.Close()
			reader := bufio.NewReader(conn)
			http_request_buf := make([]byte, 0)
			for {

				line, _, err := reader.ReadLine()
				if err != nil {
					fmt.Println(err)
					return
				}
				//fmt.Println(string(line))
				// Each line of the http message is divided using 0x0a 0x0d
				line = append(line, 0x0d, 0x0a)
				http_request_buf = append(http_request_buf, line...)
				if len(line) == 2 && line[0] == 0x0d && line[1] == 0x0a {
					break
				}
			}

			request, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(http_request_buf)))
			if err != nil {
				fmt.Println(err)
				return
			}

			bl := request.ContentLength

			if bl != 0 {
				//get http request body
				body := make([]byte, bl)
				_, err = conn.Read(body)
				if err != nil {
					fmt.Println(err)
					return
				}

				http_request_buf = append(http_request_buf, body...)
			}

			addr := strings.Split(request.Host, ":")
			r, err := regexp.Compile(`(\.((1(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})){3}`)
			if err != nil {
				fmt.Println(err)
				return
			}

			var atyp byte
			if r.Match([]byte(addr[0])) {
				atyp = socks5.ATYPIpv4
			} else {
				atyp = socks5.ATYPHost
			}

			var port uint16
			if len(addr) > 1 {

				p, err := strconv.Atoi(addr[1])
				if err != nil {
					fmt.Println(err)
					return
				}
				port = uint16(p)

			} else {
				port = 80
			}
			var proxy_conn net.Conn

			proxy_conn, err = socks5.GetSocks5Conn(*socks_server, atyp, addr[0], port)
			if err != nil {
				fmt.Println(err)
				return
			}

			defer proxy_conn.Close()

			// method is connect ues https
			if request.Method == http.MethodConnect {

				_, err := conn.Write([]byte("HTTP/1.0 200\r\n\r\n"))
				if err != nil {
					fmt.Println(err)
					return
				}

			} else {

				// request
				_, err = proxy_conn.Write(http_request_buf)

			}

			if err != nil {
				fmt.Println(err)
				return
			}

			go func() {

				conn_stdout := io.MultiWriter(conn, os.Stdout)
				_, err := io.Copy(conn_stdout, proxy_conn)

				if err != nil {
					fmt.Println(err)
					proxy_conn.Close()
					conn.Close()
				}
			}()

			_, err = io.Copy(io.MultiWriter(proxy_conn, os.Stdout), conn)
			if err != nil {
				fmt.Println(err)
				proxy_conn.Close()
				conn.Close()
			}

		}()
	}

}
