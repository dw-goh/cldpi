// Package echo — 08_tcp_echo 정답.
package echo

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// 문제 1 — 클라이언트: 한 번 보내고 한 번 받기
func Echo(addr, msg string) (string, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		// %w로 감싸면 호출한 쪽이 errors.As로 *net.OpError를 꺼내
		// "연결 거부인지 타임아웃인지"를 구분할 수 있다.
		return "", fmt.Errorf("dial %s: %w", addr, err)
	}
	defer conn.Close()

	// net.Conn은 io.Writer다(4장). 그래서 fmt.Fprintf가 그대로 먹는다.
	// 파일에 쓰든 소켓에 쓰든 코드가 같다는 게 인터페이스의 힘이다.
	if _, err := fmt.Fprintf(conn, "%s\n", msg); err != nil {
		return "", fmt.Errorf("write: %w", err)
	}

	// net.Conn은 io.Reader이기도 하다.
	// ReadString('\n')은 개행을 만날 때까지 계속 읽는다.
	// 서버가 몇 번에 나눠 보냈든 상관없다 — bufio가 알아서 모아준다.
	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read: %w", err)
	}

	// ReadString은 구분자(\n)를 포함해서 준다. 떼어내야 한다.
	// 윈도우에서 온 데이터를 대비해 \r도 같이 뗀다.
	return strings.TrimRight(line, "\r\n"), nil
}

// 문제 2 — 연결 재사용
func EchoMany(addr string, msgs []string) ([]string, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", addr, err)
	}
	defer conn.Close()

	// ★ 여기가 핵심. Reader를 루프 **바깥에서** 한 번만 만든다.
	//
	// bufio.Reader는 소켓에서 한 번에 큰 덩어리를 읽어 자기 버퍼에 쌓아둔다.
	// 서버가 "a\nb\nc\n"를 한 번에 보냈다면 첫 ReadString이 세 줄을 다 읽어와
	// "a\n"만 돌려주고 "b\nc\n"는 버퍼에 남긴다.
	// 루프 안에서 Reader를 새로 만들면 그 버퍼가 통째로 버려지고,
	// 두 번째 응답을 영원히 기다리게 된다.
	r := bufio.NewReader(conn)

	out := make([]string, 0, len(msgs))
	for _, msg := range msgs {
		if _, err := fmt.Fprintf(conn, "%s\n", msg); err != nil {
			return nil, fmt.Errorf("write %q: %w", msg, err)
		}
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("read reply for %q: %w", msg, err)
		}
		out = append(out, strings.TrimRight(line, "\r\n"))
	}
	return out, nil
}

// 문제 3 — 서버: 연결 하나 처리하기
func handleConn(c net.Conn) {
	// 어떤 경로로 끝나든 소켓을 닫는다.
	// 안 닫으면 접속이 쌓여 "too many open files"로 서버가 죽는다.
	defer c.Close()

	// Scanner는 기본적으로 줄 단위로 쪼개준다(bufio.ScanLines).
	// 커널이 데이터를 어떻게 쪼개서 주든, \n을 만날 때까지 모아서 한 줄을 준다.
	// 이게 "바이트 스트림 위에 메시지 경계를 얹는" 가장 간단한 방법이다.
	sc := bufio.NewScanner(c)

	for sc.Scan() {
		// sc.Text()에는 개행이 없다. 되돌려 보낼 때 다시 붙인다.
		if _, err := fmt.Fprintf(c, "%s\n", sc.Text()); err != nil {
			// 클라이언트가 이미 끊었을 수 있다. 조용히 끝낸다.
			return
		}
	}
	// 여기 도달하는 경우:
	//   - 클라이언트가 연결을 끊었다 (EOF) — 정상
	//   - 읽기 에러 (sc.Err() != nil)
	//   - 한 줄이 64KB를 넘었다 ← 조용히 끝나서 디버깅하기 괴로운 경우
	//
	// 진짜 서버라면 sc.Err()를 로그로 남긴다. M1에서 길이 접두사 프레이밍을
	// 직접 만드는 이유 중 하나가 이 64KB 한계다.
}

// 문제 4 — 서버: 접속 받기 ★
func Serve(ln net.Listener) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			// 리스너가 닫히면 여기로 온다. 그게 정상 종료 경로다.
			// (진짜 서버라면 일시적 에러와 영구적 에러를 구분하지만,
			//  여기서는 단순하게 전부 종료로 취급한다)
			return fmt.Errorf("accept: %w", err)
		}

		// 접속마다 goroutine 하나. goroutine이 2KB로 싸기 때문에 가능한 사치다.
		// C였다면 epoll 이벤트 루프나 스레드 풀을 짰을 자리다.
		go func() {
			// ★ goroutine 안의 panic은 다른 goroutine이 잡을 수 없다.
			// 여기서 막지 않으면 접속 하나가 서버 프로세스 전체를 죽인다.
			defer func() {
				if r := recover(); r != nil {
					// 진짜 서버라면 log.Printf로 남긴다.
					_ = r
				}
			}()
			handleConn(conn)
		}()
	}
}
