// echo-client는 echo 서버에 접속하는 클라이언트입니다. M0 DoD의 클라이언트 쪽입니다.
//
// 노트북에서:
//
//	# 인자로 준 메시지들을 보내고 끝낸다
//	go run ./08_tcp_echo/cmd/echo-client -addr pi.tailnet-xxxx.ts.net:9000 hello world
//
//	# 인자가 없으면 대화형 모드 (한 줄 입력할 때마다 보낸다. Ctrl-D로 종료)
//	go run ./08_tcp_echo/cmd/echo-client -addr 100.x.y.z:9000
//
// 이 파일은 이미 완성되어 있습니다. 8장의 문제를 풀면 바로 동작합니다.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	echo "cldpi/gotour/08_tcp_echo"
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

	// 대화형 모드. 연결 하나를 계속 재사용합니다.
	conn, err := net.Dial("tcp", *addr)
	if err != nil {
		log.Fatalf("dial %s: %v", *addr, err)
	}
	defer conn.Close()
	log.Printf("connected to %s (Ctrl-D로 종료)", conn.RemoteAddr())

	// ★ Reader는 연결당 하나만 만듭니다. 루프 안에서 매번 만들면
	// 미리 읽어둔 데이터가 버려집니다.
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
