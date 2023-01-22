package log

import (
	"fmt"
	"testing"
)

func TestLog(t *testing.T) {
	for _, level := range []Level{InfoLevel, WarnLevel, ErrorLevel, DebugLevel, TraceLevel} {
		t.Run(fmt.Sprintf("%d", level), func(t *testing.T) {
			SetLevel(level)
			Info("info")
			Infof("%s", "hello")

			Warn("warn")
			Warnf("%s", "hello")

			Error("error")
			Errorf("%s", "hello")

			Debug("debug")
			Debugf("%s", "hello")

			Trace("trace")
			Tracef("%s", "hello")
		})
	}
}
