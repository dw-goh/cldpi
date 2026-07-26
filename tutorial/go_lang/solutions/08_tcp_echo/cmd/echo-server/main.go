// echo-server (정답판) — 문제를 아직 안 풀었어도 바로 동작하는 서버입니다.
// 파이에 올려서 M0 DoD를 먼저 확인하고 싶을 때 쓰세요.
//
//	go run ./solutions/08_tcp_echo/cmd/echo-server -addr :9000
package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	echo "cldpi/gotour/solutions/08_tcp_echo"
)

func main() {
	addr := flag.String("addr", ":9000", "listen address")
	flag.Parse()

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen %s: %v", *addr, err)
	}
	log.Printf("listening on %s", ln.Addr())

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
