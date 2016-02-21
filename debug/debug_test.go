package debug

import (
	"fmt"
	"strings"
	"testing"
)

func TestTraceCallFrame(t *testing.T) {
	frames := TraceCallFrame(0, 10)
	sz := len(frames)
	s := make([]string, sz)
	for i := 0; i < sz; i++ {
		s[i] = fmt.Sprintf("\t%s", frames[i])
	}
	t.Logf("panic: %s\n%s", "test panic", strings.Join(s, "\n"))
}
