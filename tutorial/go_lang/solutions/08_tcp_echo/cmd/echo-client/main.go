// echo-client (정답판) — 문제를 아직 안 풀었어도 바로 동작하는 클라이언트입니다.
//
//	go run ./solutions/08_tcp_echo/cmd/echo-client -addr 100.x.y.z:9000 hello world
//	go run ./solutions/08_tcp_echo/cmd/echo-client -addr 100.x.y.z:9000
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	echo "cldpi/gotour/solutions/08_tcp_echo"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:9000", "server address")
	flag.Parse()

	if msgs := flag.Args(); len(msgs) > 0 {
		replies, err := echo.EchoMany(*addr, msgs)
		if err != nil {
			log.Fatal(err)
		}
		for _, r := range replies {
			fmt.Println(r)
		}
		return
	}

	conn, err := net.Dial("tcp", *addr)
	if err != nil {
		log.Fatalf("dial %s: %v", *addr, err)
	}
	defer conn.Close()
	log.Printf("connected to %s (Ctrl-D로 종료)", conn.RemoteAddr())

	server := bufio.NewReader(conn)
	stdin := bufio.NewScanner(os.Stdin)
	for stdin.Scan() {
		if _, err := fmt.Fprintf(conn, "%s\n", stdin.Text()); err != nil {
			log.Fatalf("write: %v", err)
		}
		line, err := server.ReadString('\n')
		if err != nil {
			log.Fatalf("read: %v", err)
		}
		fmt.Print(line)
	}
	if err := stdin.Err(); err != nil {
		log.Fatalf("stdin: %v", err)
	}
}
