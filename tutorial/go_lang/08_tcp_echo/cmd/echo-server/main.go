// echo-server는 TCP echo 서버를 띄웁니다. M0 DoD의 서버 쪽입니다.
//
// 라즈베리파이에서:
//
//	go run ./08_tcp_echo/cmd/echo-server -addr :9000
//
// 이 파일은 이미 완성되어 있습니다. 8장의 문제를 풀면 바로 동작합니다.
package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	echo "cldpi/gotour/08_tcp_echo"
)

func main() {
	addr := flag.String("addr", ":9000", "listen address (예: :9000, 127.0.0.1:9000)")
	flag.Parse()

	// ":9000"은 모든 인터페이스에서 듣는다는 뜻입니다.
	// 파이에서 이렇게 띄워야 노트북에서 접속할 수 있습니다.
	// "127.0.0.1:9000"으로 띄우면 그 기계 안에서만 접속됩니다.
	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen %s: %v", *addr, err)
	}
	log.Printf("listening on %s", ln.Addr())

	// Ctrl-C를 누르면 리스너를 닫습니다.
	// 그러면 Serve의 Accept가 에러를 내고 함수가 정상적으로 끝납니다.
	//
	// 나중에 FUSE 마운트를 다룰 때 이 패턴이 필수가 됩니다 —
	// 데몬이 Unmount 없이 죽으면 좀비 마운트가 남기 때문입니다.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Printf("%v 수신, 종료합니다", sig)
		ln.Close()
	}()

	if err := echo.Serve(ln); err != nil {
		log.Printf("serve 종료: %v", err)
	}
}
