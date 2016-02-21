package debug

import (
	"fmt"
	"runtime"
)

type FrameInfo struct {
	Func string
	File string
	Line int
}

func (fi FrameInfo) String() string {
	return fmt.Sprintf("%s:%d [%s]", fi.File, fi.Line, fi.Func)
}

// skip indicate to skip recording frame
// skip=1 skip caller frame, skip=2 skip caller's caller frame
// max deep 32
func TraceCallFrame(skip, deep int) []FrameInfo {
	if deep > 32 {
		deep = 32
	}
	pc := make([]uintptr, deep)
	frames := make([]FrameInfo, deep)
	runtime.Callers(skip+2, pc)
	n := 0
	for i := 0; i < deep; i++ {
		if pc[i] != 0 {
			f := runtime.FuncForPC(pc[i])
			frames[i].File, frames[i].Line = f.FileLine(pc[i])
			frames[i].Func = f.Name()
			n++
		}
	}
	return frames[:n]
}
