# Go 튜토리얼 — cldpi M0

> C/Python은 쓸 줄 알지만 Go는 처음인 사람을 위한 문제집.
> [PLAN.md](../../docs/PLAN.md)의 **M0 — 환경 구축**, 그중 "Go Tour + Effective Go (6h)" 항목에 해당한다.

## 이 튜토리얼의 목표

M0의 DoD는 이것이다:

> 노트북에서 라즈베리파이의 Go 프로그램에 TCP로 접속해 문자열을 주고받는다.

8장을 다 풀면 그 TCP echo 서버/클라이언트를 **직접 짜서** 도달한다. 마지막 장이 곧 M0의 결과물이다.

## 푸는 법

```bash
cd tutorial/go_lang

# 1장 채점
go test ./01_basics/

# 자세히 보기 — 어떤 테스트가 왜 실패했는지
go test -v ./01_basics/

# 전체 채점 (내가 푼 것만. 셸이 ./01_basics/ ./02_.../ ... 로 펼쳐준다)
go test ./0*/

# 동시성 장(06~08)은 이것도 반드시
go test -race ./06_goroutines/ ./07_select_sync/ ./08_tcp_echo/
```

각 장 디렉터리는 이렇게 생겼다:

```
01_basics/
├── README.md       ← 문제 설명. 먼저 읽는다.
├── basics.go       ← 여기의 TODO를 채운다.
└── basics_test.go  ← 채점 기준. 수정하지 않는다.
```

**테스트 파일이 곧 명세다.** 문제 설명이 애매하면 `_test.go`를 열어서 뭘 기대하는지 직접 확인하는 게 가장 빠르다. 이건 Go에서 남의 라이브러리를 읽을 때도 똑같이 쓰는 방법이다.

`import`는 직접 추가해야 한다. 처음엔 `strconv`, `strings`, `sort`, `sync`, `net` 같은 걸 언제 쓰는지 감이 안 오는 게 정상이다. 에디터가 자동으로 넣어주기도 하고, 안 되면 `goimports`를 쓰면 된다.

## 정답

```
solutions/
└── 01_basics/
    ├── basics.go       ← 주석 달린 정답
    └── basics_test.go  ← 같은 테스트 (정답이 진짜 통과하는지 확인용)
```

```bash
# 정답이 전부 통과하는지 확인
go test ./solutions/...
```

먼저 스스로 풀고, 막히면 열어보자. 정답 코드에는 "왜 이렇게 쓰는지"가 주석으로 달려 있으니 다 푼 뒤에도 한 번씩 비교해보면 좋다.

## 목차

| 장 | 주제 | 왜 필요한가 |
|---|---|---|
| [01](01_basics/) | 기초 문법 | 변수·반복·switch·포인터·클로저. C/Python과 뭐가 다른지 |
| [02](02_slices_maps/) | 슬라이스와 맵 | **Go에서 제일 많이 헷갈리는 곳.** 버그의 절반이 여기서 나온다 |
| [03](03_structs/) | 구조체와 메서드 | 값 리시버 vs 포인터 리시버, 임베딩 |
| [04](04_interfaces/) | 인터페이스 | 암묵적 만족(덕 타이핑), 타입 스위치 |
| [05](05_errors_defer/) | 에러 처리와 defer | Go에 예외는 없다. `if err != nil`을 제대로 쓰는 법 |
| [06](06_goroutines/) | goroutine과 channel | 동시성. M0의 "goroutine per connection" |
| [07](07_select_sync/) | select와 sync | 타임아웃, 뮤텍스, 요청 다중화 |
| [08](08_tcp_echo/) | TCP echo 서버 | **M0 DoD.** 파이에 올려서 노트북에서 접속한다 |

각 장은 30~60분 정도를 잡으면 된다. 8장은 조금 더 걸린다.

## 이 튜토리얼이 이후 마일스톤과 이어지는 지점

| 여기서 배운 것 | 나중에 쓰이는 곳 |
|---|---|
| 02 — 슬라이스 앨리어싱, `copy` | M1 프레임 버퍼 재사용 |
| 04 — `io.Writer` 인터페이스 | M1~M2 전 구간 (`net.Conn`이 곧 Reader/Writer다) |
| 05 — `%w` 래핑, `errors.Is` | 프로젝트 전체의 에러 전파 |
| 05 — `defer`로 반드시 닫기 | M2 파일 핸들, M5 FUSE 핸들 누수 방지 |
| 06 — goroutine per conn | M0 8장, M3 서버 |
| 07 — `map[int]chan T` + 뮤텍스 | **M1 요청 다중화** (`request_id` → `chan Response`) |
| 07 — `select` + 타임아웃 | M1 read/write deadline |
| 08 — TCP는 바이트 스트림이다 | **M1 길이 접두사 프레이밍의 존재 이유** |

## 시작하기 전에

```bash
go version   # go1.21 이상이면 된다. 이 문서는 1.26 기준
```

Go가 없다면 https://go.dev/dl/ 또는 `brew install go`.

문법이 막힐 때 볼 곳:

- [A Tour of Go (한국어)](https://go-tour-ko.appspot.com/) — 브라우저에서 바로 실행되는 공식 투어
- [Effective Go](https://go.dev/doc/effective_go) — "Go답게 쓰는 법". 2장, 5장 풀고 나서 읽으면 이해가 잘 된다
- [pkg.go.dev](https://pkg.go.dev/std) — 표준 라이브러리 문서. `go doc strings.Fields`로 터미널에서도 볼 수 있다
