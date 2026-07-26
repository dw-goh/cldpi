package echo

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

// startServer는 임의의 빈 포트에 서버를 띄우고 주소를 반환합니다.
// "127.0.0.1:0"의 0은 "아무 포트나 알아서 골라줘"라는 뜻입니다.
// 테스트에서 포트 충돌을 피하는 표준 방법입니다.
func startServer(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { ln.Close() })

	go Serve(ln) // 리스너가 닫히면 알아서 끝난다

	return ln.Addr().String()
}

// runWithTimeout은 f가 제한 시간 안에 끝나지 않으면 테스트를 실패시킵니다.
// 서버가 아직 구현되지 않았거나 응답을 안 보내면 클라이언트가 영원히 멈추는데,
// 그때 10분을 기다리지 않게 해줍니다.
func runWithTimeout(t *testing.T, msg string, f func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		f()
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal(msg)
	}
}

func TestEcho(t *testing.T) {
	addr := startServer(t)

	tests := []string{"hello", "", "안녕하세요", "with spaces and 123", strings.Repeat("x", 4096)}
	for _, msg := range tests {
		name := msg
		if len(name) > 20 {
			name = name[:20] + "..."
		}
		t.Run(name, func(t *testing.T) {
			var (
				got string
				err error
			)
			runWithTimeout(t, "응답을 5초 안에 받지 못했습니다. 서버(문제 3, 4)와 "+
				"클라이언트(문제 1)가 모두 구현됐는지 확인하세요", func() {
				got, err = Echo(addr, msg)
			})
			if err != nil {
				t.Fatalf("Echo(%q) 에러: %v", msg, err)
			}
			if got != msg {
				t.Errorf("Echo(%q) = %q; want %q", msg, got, msg)
			}
		})
	}
}

func TestEchoDialError(t *testing.T) {
	// 아무도 안 듣고 있는 포트. 에러가 나야 하고 panic이 나면 안 된다.
	_, err := Echo("127.0.0.1:1", "hello")
	if err == nil {
		t.Error("닫힌 포트에 Echo했는데 err = nil; want 에러")
	}
}

func TestEchoMany(t *testing.T) {
	addr := startServer(t)

	msgs := []string{"one", "two", "three", "네", "다섯"}
	var (
		got []string
		err error
	)
	runWithTimeout(t, "EchoMany가 5초 안에 끝나지 않았습니다. bufio.Reader를 "+
		"루프 안에서 매번 새로 만들면 버퍼에 남은 데이터가 버려져서 여기서 멈춥니다", func() {
		got, err = EchoMany(addr, msgs)
	})
	if err != nil {
		t.Fatalf("EchoMany 에러: %v", err)
	}
	if len(got) != len(msgs) {
		t.Fatalf("응답 %d개를 받았습니다; want %d개 (%v)", len(got), len(msgs), got)
	}
	for i := range msgs {
		if got[i] != msgs[i] {
			t.Errorf("응답 %d = %q; want %q", i, got[i], msgs[i])
		}
	}

	runWithTimeout(t, "EchoMany(addr, nil)이 끝나지 않습니다", func() {
		if got, err := EchoMany(addr, nil); err != nil || len(got) != 0 {
			t.Errorf("EchoMany(addr, nil) = %v, %v; want 빈 슬라이스, nil", got, err)
		}
	})
}

// 서버가 여러 접속을 진짜로 동시에 처리하는지, 그리고 응답이 섞이지 않는지 봅니다.
// M0 DoD의 실질적인 내용입니다.
func TestConcurrentClients(t *testing.T) {
	addr := startServer(t)

	const n = 100
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			msg := fmt.Sprintf("client-%03d", i)
			got, err := Echo(addr, msg)
			if err != nil {
				t.Errorf("Echo(%q) 에러: %v", msg, err)
				return
			}
			if got != msg {
				t.Errorf("Echo(%q) = %q — 다른 클라이언트의 응답을 받았습니다", msg, got)
			}
		}()
	}

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("동시 접속 100개가 끝나지 않습니다. Serve가 접속마다 goroutine을 띄우는지 확인하세요 " +
			"(go 없이 handleConn을 직접 부르면 한 번에 하나씩만 처리됩니다)")
	}
}

// ★ 이 장의 핵심 개념 테스트.
//
// TCP는 바이트 스트림입니다. 내가 Write를 몇 번 나눠서 했는지,
// 얼마나 큰 덩어리로 보냈는지는 상대에게 전혀 전달되지 않습니다.
// 커널이 마음대로 합치기도(Nagle) 쪼개기도(MTU) 합니다.
//
// 그래서 "메시지"라는 개념은 우리가 직접 만들어야 합니다.
// 여기서는 "\n까지가 한 메시지"라는 약속을 씁니다.
func TestStreamHasNoMessageBoundaries(t *testing.T) {
	addr := startServer(t)

	t.Run("Write 한 번에 여러 메시지", func(t *testing.T) {
		c, err := net.Dial("tcp", addr)
		if err != nil {
			t.Fatalf("dial: %v", err)
		}
		defer c.Close()
		c.SetDeadline(time.Now().Add(5 * time.Second))

		// Write는 한 번인데 메시지는 세 개다.
		if _, err := fmt.Fprint(c, "one\ntwo\nthree\n"); err != nil {
			t.Fatalf("write: %v", err)
		}

		r := bufio.NewReader(c)
		for _, want := range []string{"one", "two", "three"} {
			line, err := r.ReadString('\n')
			if err != nil {
				t.Fatalf("%q를 기다리다 에러: %v (서버가 줄 단위로 나누지 않고 "+
					"받은 덩어리를 통째로 되돌려 보내고 있을 수 있습니다)", want, err)
			}
			if got := strings.TrimRight(line, "\r\n"); got != want {
				t.Fatalf("응답 = %q; want %q", got, want)
			}
		}
	})

	t.Run("메시지 하나를 Write 여러 번에 나눠서", func(t *testing.T) {
		c, err := net.Dial("tcp", addr)
		if err != nil {
			t.Fatalf("dial: %v", err)
		}
		defer c.Close()
		c.SetDeadline(time.Now().Add(5 * time.Second))

		// "hello\n" 한 메시지를 세 조각으로 나눠 보낸다.
		// 서버는 \n을 볼 때까지 기다렸다가 하나의 메시지로 처리해야 한다.
		for _, part := range []string{"he", "ll", "o\n"} {
			if _, err := fmt.Fprint(c, part); err != nil {
				t.Fatalf("write %q: %v", part, err)
			}
			time.Sleep(30 * time.Millisecond) // 확실히 따로 도착하게
		}

		c.SetReadDeadline(time.Now().Add(5 * time.Second))
		line, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			t.Fatalf("read: %v (조각을 각각 하나의 메시지로 처리했을 수 있습니다)", err)
		}
		if got := strings.TrimRight(line, "\r\n"); got != "hello" {
			t.Errorf("응답 = %q; want %q — 서버가 조각을 모으지 않았습니다", got, "hello")
		}
	})
}

// handleConn만 따로 시험합니다.
// net.Pipe는 실제 네트워크 없이 메모리 안에서 연결된 net.Conn 한 쌍을 만들어줍니다.
func TestHandleConn(t *testing.T) {
	client, server := net.Pipe()
	go handleConn(server)

	client.SetDeadline(time.Now().Add(5 * time.Second))
	if _, err := fmt.Fprint(client, "ping\n"); err != nil {
		t.Fatalf("write: %v", err)
	}
	line, err := bufio.NewReader(client).ReadString('\n')
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got := strings.TrimRight(line, "\r\n"); got != "ping" {
		t.Errorf("= %q; want %q", got, "ping")
	}

	// 클라이언트가 끊으면 handleConn도 끝나야 한다.
	client.Close()
}

// handleConn이 연결을 닫는지 확인합니다.
// 안 닫으면 접속이 쌓여서 결국 "too many open files"로 서버가 죽습니다.
func TestHandleConnClosesConnection(t *testing.T) {
	addr := startServer(t)

	c, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Close()
	c.SetDeadline(time.Now().Add(5 * time.Second))

	if _, err := fmt.Fprint(c, "bye\n"); err != nil {
		t.Fatalf("write: %v", err)
	}
	r := bufio.NewReader(c)
	if _, err := r.ReadString('\n'); err != nil {
		t.Fatalf("read: %v", err)
	}

	// 우리 쪽에서 쓰기를 끝냈다고 알린다 → 서버의 Scan()이 false가 되어야 한다.
	c.(*net.TCPConn).CloseWrite()

	c.SetReadDeadline(time.Now().Add(5 * time.Second))
	if _, err := r.ReadString('\n'); err == nil {
		t.Error("클라이언트가 끊었는데 서버가 연결을 닫지 않았습니다. handleConn에 defer c.Close()가 있나요?")
	}
}

func TestServeStopsWhenListenerCloses(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	errCh := make(chan error, 1)
	go func() { errCh <- Serve(ln) }()

	time.Sleep(50 * time.Millisecond)
	ln.Close()

	select {
	case err := <-errCh:
		if err == nil {
			t.Error("Serve가 nil을 반환했습니다; Accept의 에러를 그대로 반환해야 합니다")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("리스너를 닫았는데 Serve가 반환하지 않습니다. Accept 에러 시 return하는지 확인하세요")
	}
}
