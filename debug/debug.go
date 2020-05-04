package debug

import(
    "os"
    "log"
)

// DebugMode mode is set by seting the env var GO_DEBUG to any value.
// If set, some operations will log to /tmp/golog.log
var DebugMode bool

const logFile string = "/tmp/golog.log"
// GoLog writes a string to a file
func GoLog(logEntry string) {
    if  DebugMode {
        f, err := os.OpenFile(logFile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
        if err != nil {
            panic(err)
        }
        defer f.Close()
        
        log.SetOutput(f)
        log.Println(logEntry)
    }
}