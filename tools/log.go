package tools

import (
	"fmt"
	"log"
	"strings"
	"sync"
)

// color palette map
var (
	colorOff    = "\033[0m"
	colorRed    = "\033[0;31m"
	colorGreen  = "\033[0;32m"
	colorOrange = "\033[0;33m"
	colorBlue   = "\033[0;34m"
	colorPurple = "\033[0;35m"
	colorCyan   = "\033[0;36m"
	colorGray   = "\033[0;37m"

	// Level prefix template
	levelFatalPrefix = "[FATAL] "
	levelErrorPrefix = "[ERROR] "
	levelWarnPrefix  = "[WARN]  "
	levelInfoPrefix  = "[INFO]  "
	levelDebugPrefix = "[DEBUG] "
	levelTracePrefix = "[TRACE] "

	// Level definition
	levelFatal = 6
	levelError = 5
	levelWarn  = 4
	levelInfo  = 3
	levelDebug = 2
	levelTrace = 1

	levelFatalName = "FATAL"
	levelErrorName = "ERROR"
	levelWarnName  = "WARN"
	levelInfoName  = "INFO"
	levelDebugName = "DEBUG"
	levelTraceName = "TRACE"
)

type LevelOption struct {
	level  int
	name   string
	color  string
	prefix string
}

type OKLogger struct {
	mu    sync.RWMutex
	color bool
	level int
}

var levelOptions = map[int]LevelOption{
	levelFatal: {level: levelFatal, name: levelFatalName, color: colorRed, prefix: levelFatalPrefix},
	levelError: {level: levelError, name: levelErrorName, color: colorOrange, prefix: levelErrorPrefix},
	levelWarn:  {level: levelWarn, name: levelWarnName, color: colorPurple, prefix: levelWarnPrefix},
	levelInfo:  {level: levelFatal, name: levelInfoName, color: colorGreen, prefix: levelInfoPrefix},
	levelDebug: {level: levelDebug, name: levelDebugName, color: colorCyan, prefix: levelDebugPrefix},
	levelTrace: {level: levelTrace, name: levelTraceName, color: colorGray, prefix: levelTracePrefix},
}

var okLogger = &OKLogger{
	color: true,
	level: levelInfo,
}

// WithColor explicitly turn on colorful features on the log
func WithColor() {
	okLogger.mu.Lock()
	defer okLogger.mu.Unlock()
	okLogger.color = true
}

// WithoutColor explicitly turn off colorful features on the log
func WithoutColor() {
	okLogger.mu.Lock()
	defer okLogger.mu.Unlock()
	okLogger.color = false
}

func SetLevel(newLevel int) {
	okLogger.mu.Lock()
	defer okLogger.mu.Unlock()

	if newLevel >= levelTrace && newLevel <= levelFatal {
		okLogger.level = newLevel
	}
}

func SetLevelByName(levelName string) {
	okLogger.mu.Lock()
	defer okLogger.mu.Unlock()

	for level, option := range levelOptions {
		if strings.EqualFold(levelName, option.name) {
			okLogger.level = level
		}

	}
}

func output(level int, payload string) {
	if level < okLogger.level {
		return
	}

	option := levelOptions[level]

	if okLogger.color {
		lines := strings.Split(payload, "\n")
		var colorLines []string
		for _, line := range lines {
			if len(colorLines) > 0 {
				colorLines = append(colorLines, fmt.Sprintf("%s%v%s", option.color, line, colorOff))
			} else {
				colorLines = append(colorLines, fmt.Sprintf("%s%v%v%s", option.color, option.prefix, line, colorOff))
			}
		}
		log.Print(strings.Join(colorLines, "\n"))
	} else {
		if strings.HasSuffix(payload, "\n") {
			log.Printf("%v%v", option.prefix, payload)
		} else {
			log.Printf("%v%v\n", option.prefix, payload)
		}
	}
}

func Fatalf(format string, v ...interface{}) {
	output(levelFatal, fmt.Sprintf(format, v...))
}

func Errorf(format string, v ...interface{}) {
	output(levelError, fmt.Sprintf(format, v...))
}

func Warnf(format string, v ...interface{}) {
	output(levelWarn, fmt.Sprintf(format, v...))
}

func Infof(format string, v ...interface{}) {
	output(levelInfo, fmt.Sprintf(format, v...))
}

func Debugf(format string, v ...interface{}) {
	output(levelDebug, fmt.Sprintf(format, v...))
}

func Tracef(format string, v ...interface{}) {
	output(levelTrace, fmt.Sprintf(format, v...))
}

func Fatal(v ...interface{}) {
	output(levelFatal, fmt.Sprint(v...))
}

func Error(v ...interface{}) {
	output(levelError, fmt.Sprint(v...))
}

func Warn(v ...interface{}) {
	output(levelWarn, fmt.Sprint(v...))
}

func Info(v ...interface{}) {
	output(levelInfo, fmt.Sprint(v...))
}

func Debug(v ...interface{}) {
	output(levelDebug, fmt.Sprint(v...))
}

func Trace(v ...interface{}) {
	output(levelTrace, fmt.Sprint(v...))
}
