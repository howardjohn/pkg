// Copyright Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"fmt"

	"github.com/go-logr/logr"
)

// zapLogger is a logr.Logger that uses Zap to log. This is needed to get
// libraries, namely Kubernetes/klog, that use logr, to use our standard logging.
// This enables standard formatting, scope filtering, and options. The logr
// interface does not have a concept of Debug/Info/Warn/Error as we do. Instead,
// logging is based on Verbosity levels, where 0 is the most important. We treat
// levels 0-3 as info level and 4+ as debug; there are no warnings. This
// threshold is fairly arbitrary based on inspection of Kubernetes usage and
// https://kubernetes.io/docs/reference/kubectl/cheatsheet/#kubectl-output-verbosity-and-debugging.
// Errors are passed through as errors.
// Zap does come with its own logr implementation, but we have chosen to re-implement to allow usage of
// our Scope - in particular, this allows changing the logging level of kubernetes logs by users.
type zapLogger struct {
	l      *Scope
	lvl    int
	lvlSet bool
}

const debugLevelThreshold = 3

func (zl *zapLogger) Enabled() bool {
	if zl.lvlSet && zl.lvl > debugLevelThreshold {
		return zl.l.DebugEnabled()
	}
	return zl.l.InfoEnabled()
}

// Logs will come in with newlines, but our logger auto appends newline
func trimNewline(msg string) string {
	if len(msg) == 0 {
		return msg
	}
	lc := len(msg) - 1
	if msg[lc] == '\n' {
		return msg[:lc]
	}
	return msg
}

func (zl *zapLogger) Info(msg string, keysAndVals ...interface{}) {
	if zl.lvlSet && zl.lvl > debugLevelThreshold {
		zl.l.Debug(trimNewline(msg), keysAndVals)
	} else {
		zl.l.Info(trimNewline(msg), keysAndVals)
	}
}

func (zl *zapLogger) Error(err error, msg string, keysAndVals ...interface{}) {
	if zl.l.ErrorEnabled() {
		if err == nil {
			zl.l.Error(trimNewline(msg), keysAndVals)
		} else {
			zl.l.Error(fmt.Sprintf("%v: %s", err.Error(), msg), keysAndVals)
		}
	}
}

func (zl *zapLogger) V(level int) logr.Logger {
	return &zapLogger{
		lvl:    zl.lvl + level,
		l:      zl.l,
		lvlSet: true,
	}
}

func (zl *zapLogger) WithValues(keysAndValues ...interface{}) logr.Logger {
	return newLogrAdapter(zl.l.WithLabels(keysAndValues...))
}

func (zl *zapLogger) WithName(name string) logr.Logger {
	return zl
}

// NewLogger creates a new logr.Logger using the given Zap Logger to log.
func newLogrAdapter(l *Scope) logr.Logger {
	return &zapLogger{
		l:      l,
		lvl:    0,
		lvlSet: false,
	}
}
