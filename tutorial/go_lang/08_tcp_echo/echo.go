// Package echo — TCP echo 서버와 클라이언트.
//
//	go test ./08_tcp_echo/
//	go test -race ./08_tcp_echo/
//
// 이 장이 M0의 DoD입니다:
// "노트북에서 라즈베리파이의 Go 프로그램에 TCP로 접속해 문자열을 주고받는다."
//
// 프로토콜은 아주 단순합니다. 줄 단위입니다.
//
//	클라이언트 → 서버:  "hello\n"
//	서버 → 클라이언트:  "hello\n"
//
// ★ 이 장의 핵심 개념: TCP는 바이트 스트림입니다. 메시지 경계가 없습니다.
// 내가 Write를 몇 번 했는지는 상대방에게 전달되지 않습니다.
// 그래서 "\n까지가 한 메시지"라는 약속을 우리가 직접 만들어야 합니다.
package echo

import (
	"net"
)

// 문제 1 — 클라이언트: 한 번 보내고 한 번 받기
//
// addr에 접속해서 msg를 한 줄로 보내고, 돌아온 한 줄을 반환하세요.
// 반환값에는 끝의 개행문자가 없어야 합니다.
//
//	Echo("127.0.0.1:9000", "hello") → ("hello", nil)
//
// 순서:
//  1. conn, err := net.Dial("tcp", addr)
//  2. err != nil이면 그대로 반환
//  3. defer conn.Close()               ← 잊으면 소켓이 샙니다
//  4. fmt.Fprintf(conn, "%s\n", msg)   ← net.Conn은 io.Writer입니다 (4장!)
//  5. bufio.NewReader(conn).ReadString('\n')로 한 줄 읽기
//  6. strings.TrimRight(line, "\r\n")으로 개행 제거 후 반환
//
// 에러는 감싸서 반환하세요 (5장): fmt.Errorf("dial %s: %w", addr, err)
func Echo(addr, msg string) (string, error) {
	// TODO
	return "", nil
}

// 문제 2 — 연결 재사용
//
// 한 번만 접속해서 msgs를 순서대로 보내고 받은 응답들을 반환하세요.
//
//	EchoMany(addr, []string{"a", "b"}) → ([]string{"a", "b"}, nil)
//
// 매번 새로 접속하는 것(문제 1)과 달리, 연결 하나로 여러 요청을 처리합니다.
// TCP 핸드셰이크가 한 번만 일어나므로 훨씬 빠릅니다.
//
// ★ bufio.Reader는 **연결마다 하나만** 만들어야 합니다.
// 루프 안에서 매번 bufio.NewReader(conn)을 새로 만들면,
// 이전 Reader가 미리 읽어둔(buffered) 데이터가 통째로 버려집니다.
// 서버가 "a\nb\n"을 한 번에 보냈다면 두 번째 응답을 영원히 못 받습니다.
// 이게 바이트 스트림을 다룰 때 가장 자주 나오는 실수입니다.
func EchoMany(addr string, msgs []string) ([]string, error) {
	// TODO
	return nil, nil
}

// 문제 3 — 서버: 연결 하나 처리하기
//
// c에서 한 줄씩 읽어서 그대로 되돌려 보내세요.
// 상대가 연결을 끊으면(EOF) 조용히 끝냅니다.
//
// 순서:
//  1. defer c.Close()
//  2. sc := bufio.NewScanner(c)
//  3. for sc.Scan() { ... }              ← Scan은 개행을 만날 때까지 모읍니다
//  4. 루프 안에서 sc.Text()에 "\n"을 붙여 c에 Write
//
// sc.Text()는 개행이 **제거된** 한 줄을 줍니다. 되돌려 보낼 때 다시 붙여야 합니다.
//
// ★ bufio.Scanner의 기본 한 줄 최대 길이는 64KB입니다.
// 더 긴 줄이 오면 Scan()이 false를 반환하고 조용히 끝납니다.
// "왜 큰 파일을 보내면 잘리지?"의 범인이 대개 이것입니다.
// M1에서 길이 접두사 프레이밍을 만드는 이유 중 하나이기도 합니다.
func handleConn(c net.Conn) {
	// TODO
}

// 문제 4 — 서버: 접속 받기 ★
//
// ln에서 접속을 계속 받아 각각을 goroutine으로 처리하세요.
// Accept가 실패하면 그 에러를 반환하고 끝냅니다
// (리스너가 닫히면 Accept가 에러를 내므로, 이게 곧 종료 조건입니다).
//
//	for {
//	    conn, err := ln.Accept()
//	    if err != nil {
//	        return err
//	    }
//	    go handleConn(conn)
//	}
//
// 이 네 줄이 Go 서버의 전부입니다.
// C였다면 epoll/kqueue 이벤트 루프를 짜거나 스레드 풀을 관리해야 했을 자리인데,
// Go는 goroutine이 싸서(2KB) 접속마다 하나씩 띄워도 됩니다.
// 런타임이 내부적으로 epoll을 써주고, 우리는 동기 코드처럼 짜면 됩니다.
//
// ★ 실전에서는 handleConn을 부르는 goroutine에 defer recover()를 답니다.
// goroutine 안의 panic은 다른 goroutine이 잡을 수 없어서
// 접속 하나 때문에 서버 프로세스 전체가 죽기 때문입니다 (5장 참고).
func Serve(ln net.Listener) error {
	// TODO
	return nil
}
