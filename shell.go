// powershell provides a PowerShell session in which
// you can execute commands serially (one at a time).
//
// It detects the current code page when a session is
// created. And when running a command, it encodes
// the command and decodes the output from stdout or
// stderr with the encoding corresponding to the code page.
package powershell

import (
	"bytes"
	crand "crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/text/encoding/unicode"
	"io"
	"math/rand"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
)

const exeFilename = "powershell.exe"
const newline = "\r\n"

const (
	boundaryPrefix            = "$command"
	boundaryPrefixLen         = 8
	boundaryRandomPartByteLen = 12
)

// Shell represents a PowerShell session.
type Shell struct {
	codePage       int
	enc            encoding.Encoding
	cmd            *exec.Cmd
	stdin          io.WriteCloser
	stdout         io.ReadCloser
	stderr         io.ReadCloser
	rnd            *rand.Rand
	boundaryRndBuf [boundaryRandomPartByteLen]byte
	boundaryBuf    [boundaryPrefixLen + 2*boundaryRandomPartByteLen]byte
}

// ErrUnsupportedCodePage is the error returned from the New
// method if the detected code page is not in the Encodings map.
var ErrUnsupportedCodePage = errors.New("unsupported code page")

// Encodings contains a mapping from code page to encoding.
// Only code page 932 and 65001 are supported by default.
// To use with other code pages, you need to add an entry
// before calling the New method.
var Encodings = map[int]encoding.Encoding{
	932:   japanese.ShiftJIS,
	949:   korean.EUCKR,
	//65001: encoding.Nop,
	65001:	unicode.UTF8,
}

// New creates a new PowerShell session.
func New() (*Shell, error) {
	s, err := newShell()
	if err != nil {
		return nil, err
	}

	cp, err := s.detectCodePage()
	if err != nil {
		return nil, err
	}

	enc := Encodings[cp]
	if enc == nil {
		return nil, ErrUnsupportedCodePage
	}

	s.codePage = cp
	s.enc = enc
	return s, nil
}

func newShell() (*Shell, error) {
	exePath, err := exec.LookPath(exeFilename)
	if err != nil {
		return nil, fmt.Errorf("need powershell.exe: %w", err)
	}

	cmd := exec.Command(exePath, "-NoLogo", "-NoExit", "-Command", "-")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("stderr pipe: %w", err)
	}

	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("start powershell: %w", err)
	}

	rnd := rand.New(rand.NewSource(randSeed()))

	s := &Shell{
		enc:    encoding.Nop,
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
		rnd:    rnd,
	}
	copy(s.boundaryBuf[:], []byte(boundaryPrefix))
	return s, nil
}

// CodePage returns the detected code page of the session.
func (s *Shell) CodePage() int {
	return s.codePage
}

func (s *Shell) detectCodePage() (int, error) {
	out, err := s.Exec("chcp")
	if err != nil {
		return 0, fmt.Errorf("get codepage: %s", err)
	}
	out = strings.TrimRight(out, " \r\n")
	i := strings.LastIndex(out, ": ")
	if i == -1 {
		return 0, errors.New("invalid codepage output")
	}
	cp, err := strconv.Atoi(out[i+len(": "):])
	if err != nil {
		return 0, errors.New("non-numeric codepage")
	}
	return cp, nil
}

// Exec execute a command in the session.
// This method is not goroutine safe.
func (s *Shell) Exec(cmd string) (stdout string, err error) {
	// wrap the command in special markers so we know when to stop reading from the pipes
	boundary := s.randomBoundary()
	full := fmt.Sprintf("%s; echo '%s'; [Console]::Error.WriteLine('%s')%s", cmd, boundary, boundary, newline)
	full, err = s.enc.NewEncoder().String(full)
	if err != nil {
		return "", fmt.Errorf("encode command: %s", err)
	}
	_, err = s.stdin.Write([]byte(full))
	if err != nil {
		return "", fmt.Errorf("write command: %s", err)
	}

	var stderr string
	var wg sync.WaitGroup
	wg.Add(2)
	go readOutput(s.stdout, s.enc.NewDecoder(), &stdout, boundary, &wg)
	go readOutput(s.stderr, s.enc.NewDecoder(), &stderr, boundary, &wg)
	wg.Wait()
	if len(stderr) > 0 {
		return stdout, errors.New(stderr)
	}
	return stdout, nil
}

// Exit closes the session.
func (s *Shell) Exit() error {
	_, err := s.stdin.Write([]byte("exit" + newline))
	if err != nil {
		return fmt.Errorf("write exit: %s", err)
	}

	err = s.stdin.Close()
	if err != nil {
		return fmt.Errorf("close stdin: %s", err)
	}

	return nil
}

func (s *Shell) randomBoundary() string {
	_, _ = s.rnd.Read(s.boundaryRndBuf[:])
	hex.Encode(s.boundaryBuf[boundaryPrefixLen:], s.boundaryRndBuf[:])
	return string(s.boundaryBuf[:])
}

func readOutput(r io.Reader, dec *encoding.Decoder, out *string, boundary string, wg *sync.WaitGroup) {
	var bout []byte
	defer func() {
		*out = string(bout)
		wg.Done()
	}()

	marker := []byte(boundary + newline)
	const bufsize = 64
	buf := make([]byte, bufsize)
	for {
		n, err := r.Read(buf)
		if err != nil {
			return
		}

		decoded, err := dec.Bytes(buf[:n])
		if err != nil {
			return
		}

		bout = append(bout, decoded...)
		if bytes.HasSuffix(bout, marker) {
			bout = bout[:len(bout)-len(marker)]
			return
		}
	}
}

func randSeed() int64 {
	var b [8]byte
	if _, err := crand.Read(b[:]); err != nil {
		return time.Now().UnixNano()
	}
	return int64(binary.LittleEndian.Uint64(b[:]))
}
