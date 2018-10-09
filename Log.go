package pgo

import (
    "bytes"
    "fmt"
    "os"
    "path/filepath"
    "runtime"
    "strings"
    "sync"
    "time"

    "github.com/pinguo/pgo/Util"
)

const (
    LevelNone   = 0x00
    LevelDebug  = 0x01
    LevelInfo   = 0x02
    LevelNotice = 0x04
    LevelWarn   = 0x08
    LevelError  = 0x10
    LevelFatal  = 0x20
    LevelAll    = 0xFF
)

func levelToString(level int) string {
    switch level {
    case LevelNone:
        return "NONE"
    case LevelDebug:
        return "DEBUG"
    case LevelInfo:
        return "INFO"
    case LevelNotice:
        return "NOTICE"
    case LevelWarn:
        return "WARN"
    case LevelError:
        return "ERROR"
    case LevelFatal:
        return "FATAL"
    case LevelAll:
        return "ALL"
    default:
        panic(fmt.Sprintf("unknown log level: %x", level))
    }
}

func stringToLevel(level string) int {
    switch strings.ToUpper(level) {
    case "NONE":
        return LevelNone
    case "DEBUG":
        return LevelDebug
    case "INFO":
        return LevelInfo
    case "NOTICE":
        return LevelNotice
    case "WARN":
        return LevelWarn
    case "ERROR":
        return LevelError
    case "FATAL":
        return LevelFatal
    case "ALL":
        return LevelAll
    default:
        panic(fmt.Sprintf("unknown log level: %s", level))
    }
}

// parse comma separated level string to int format
// eg. `debug,info` => 0x03
func parseLevels(str string) int {
    levels := LevelNone
    parts := strings.Split(str, ",")

    for _, v := range parts {
        v = strings.TrimSpace(v)
        levels |= stringToLevel(v)
    }

    return levels
}

type LogItem struct {
    When    time.Time
    Level   int
    Name    string
    LogId   string
    Trace   string
    Message string
}

// log component, configuration:
// "log": {
//     "levels": "ALL",
//     "traceLevels": "DEBUG"
//     "chanLen": 1000,
//     "flushInterval": "60s",
//     "targets": {
//         "info": {
//             "class": "@pgo/FileTarget",
//             "levels": "DEBUG,INFO,NOTICE",
//             "filePath": "@runtime/info.log",
//             "maxLogFile": 10
//         },
//         "error": {
//             "class": "@pgo/FileTarget",
//             "levels": "WARN,ERROR,FATAL",
//             "filePath": "@runtime/error.log",
//             "maxLogFile": 10
//         }
//     }
// }
type Dispatcher struct {
    levels        int
    chanLen       int
    traceLevels   int
    flushInterval time.Duration
    targets       map[string]ITarget
    msgChan       chan *LogItem
    wg            sync.WaitGroup
}

func (d *Dispatcher) Construct() {
    d.levels = LevelAll
    d.chanLen = 1000
    d.traceLevels = LevelDebug
    d.flushInterval = 60 * time.Second
}

func (d *Dispatcher) Init() {
    d.msgChan = make(chan *LogItem, d.chanLen)

    if len(d.targets) == 0 {
        // use console target as default
        d.targets = make(map[string]ITarget)
        d.targets["console"] = CreateObject("@pgo/ConsoleTarget").(ITarget)
    }

    // start loop
    d.wg.Add(1)
    go d.loop()
}

func (d *Dispatcher) SetLevels(v interface{}) {
    if _, ok := v.(string); ok {
        d.levels = parseLevels(v.(string))
    } else if _, ok := v.(int); ok {
        d.levels = v.(int)
    } else {
        panic(fmt.Sprintf("Dispatcher: invalid levels: %v", v))
    }
}

func (d *Dispatcher) SetChanLen(len int) {
    d.chanLen = len
}

func (d *Dispatcher) SetTraceLevels(v interface{}) {
    if _, ok := v.(string); ok {
        d.traceLevels = parseLevels(v.(string))
    } else if _, ok := v.(int); ok {
        d.traceLevels = v.(int)
    } else {
        panic(fmt.Sprintf("Dispatcher: invalid trace levels: %v", v))
    }
}

func (d *Dispatcher) SetFlushInterval(interval string) {
    d.flushInterval, _ = time.ParseDuration(interval)
}

func (d *Dispatcher) SetTargets(targets map[string]interface{}) {
    d.targets = make(map[string]ITarget)

    for name, val := range targets {
        if config, ok := val.(map[string]interface{}); ok {
            if _, ok := config["class"]; !ok {
                config["class"] = "@pgo/ConsoleTarget"
            }
        }

        d.targets[name] = CreateObject(val).(ITarget)
    }
}

func (d *Dispatcher) GetLogger(name, logId string) *Logger {
    return &Logger{name, logId, d}
}

func (d *Dispatcher) GetProfiler() *Profiler {
    return &Profiler{}
}

func (d *Dispatcher) IsHandling(level int) bool {
    return level&d.levels != 0
}

func (d *Dispatcher) AddItem(item *LogItem) {
    if d.levels&item.Level != 0 {
        if d.traceLevels&item.Level != 0 {
            if _, file, line, ok := runtime.Caller(3); ok {
                if pos := strings.LastIndex(file, "src/"); pos > 0 {
                    file = file[pos+4:]
                }

                item.Trace = fmt.Sprintf("[%s:%d]", file, line)
            }
        }

        d.msgChan <- item
    }
}

// close msg chan and wait loop end
func (d *Dispatcher) Flush() {
    close(d.msgChan)
    d.wg.Wait()
}

func (d *Dispatcher) loop() {
    flushTimer := time.Tick(d.flushInterval)

    for {
        select {
        case item, ok := <-d.msgChan:
            for _, target := range d.targets {
                if ok {
                    target.Process(item)
                } else {
                    target.Flush(true)
                }
            }

            if !ok {
                goto end
            }
        case <-flushTimer:
            for _, target := range d.targets {
                target.Flush(false)
            }
        }
    }

end:
    d.wg.Done()
}

// logger component
type Logger struct {
    name       string
    logId      string
    dispatcher *Dispatcher
}

func (l *Logger) log(level int, format string, v ...interface{}) {
    if !l.dispatcher.IsHandling(level) {
        return
    }

    item := &LogItem{
        When:  time.Now(),
        Level: level,
        Name:  l.name,
        LogId: l.logId,
    }

    if len(v) == 0 {
        item.Message = format
    } else {
        item.Message = fmt.Sprintf(format, v...)
    }

    l.dispatcher.AddItem(item)
}

func (l *Logger) Debug(format string, v ...interface{}) {
    l.log(LevelDebug, format, v...)
}

func (l *Logger) Info(format string, v ...interface{}) {
    l.log(LevelInfo, format, v...)
}

func (l *Logger) Notice(format string, v ...interface{}) {
    l.log(LevelNotice, format, v...)
}

func (l *Logger) Warn(format string, v ...interface{}) {
    l.log(LevelWarn, format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
    l.log(LevelError, format, v...)
}

func (l *Logger) Fatal(format string, v ...interface{}) {
    l.log(LevelFatal, format, v...)
}

// profiler component
type Profiler struct {
    pushLog      []string
    counting     map[string][2]int
    profile      map[string][2]int
    profileStack map[string]time.Time
}

func (p *Profiler) Reset() {
    p.pushLog = nil
    p.counting = nil
    p.profile = nil
    p.profileStack = nil
}

func (p *Profiler) PushLog(key string, v interface{}) {
    if nil == p.pushLog {
        p.pushLog = make([]string, 0)
    }

    pl := fmt.Sprintf("%s=%s", key, Util.ToString(v))
    p.pushLog = append(p.pushLog, pl)
}

func (p *Profiler) Counting(key string, hit, total int) {
    if nil == p.counting {
        p.counting = make(map[string][2]int)
    }

    v := p.counting[key]

    if hit > 0 {
        v[0] += hit
    }

    if total <= 0 {
        total = 1
    }

    v[1] += total
    p.counting[key] = v
}

func (p *Profiler) ProfileStart(key string) {
    if nil == p.profileStack {
        p.profileStack = make(map[string]time.Time)
    }

    p.profileStack[key] = time.Now()
}

func (p *Profiler) ProfileStop(key string) {
    if nil == p.profile {
        p.profile = make(map[string][2]int)
    }

    if startTime, ok := p.profileStack[key]; ok {
        delete(p.profileStack, key)
        elapse := time.Now().Sub(startTime)

        v, _ := p.profile[key]
        v[0] += int(elapse.Nanoseconds() / 1e6)
        v[1] += 1

        p.profile[key] = v
    }
}

func (p *Profiler) GetPushLogString() string {
    if len(p.pushLog) == 0 {
        return ""
    }

    return strings.Join(p.pushLog, " ")
}

func (p *Profiler) GetCountingString() string {
    if len(p.counting) == 0 {
        return ""
    }

    cs := make([]string, 0)
    for k, v := range p.counting {
        cs = append(cs, fmt.Sprintf("%s=%d/%d", k, v[0], v[1]))
    }

    return strings.Join(cs, " ")
}

func (p *Profiler) GetProfileString() string {
    if len(p.profile) == 0 {
        return ""
    }

    ps := make([]string, 0)
    for k, v := range p.profile {
        ps = append(ps, fmt.Sprintf("%s=%d(ms)/%d", k, v[0], v[1]))
    }

    return strings.Join(ps, " ")
}

// base class of output target
type Target struct {
    levels    int
    formatter IFormatter
}

func (t *Target) SetLevels(v interface{}) {
    if _, ok := v.(string); ok {
        t.levels = parseLevels(v.(string))
    } else if _, ok := v.(int); ok {
        t.levels = v.(int)
    } else {
        panic(fmt.Sprintf("Target: invalid levels: %v", v))
    }
}

func (t *Target) SetFormatter(v interface{}) {
    if ptr, ok := v.(IFormatter); ok {
        t.formatter = ptr
    } else if class, ok := v.(string); ok {
        t.formatter = CreateObject(class).(IFormatter)
    } else if config, ok := v.(map[string]interface{}); ok {
        t.formatter = CreateObject(config).(IFormatter)
    } else {
        panic(fmt.Sprintf("Target: invalid formatter: %v", v))
    }
}

func (t *Target) IsHandling(level int) bool {
    return t.levels&level != 0
}

func (t *Target) Format(item *LogItem) string {
    if t.formatter != nil {
        return t.formatter.Format(item)
    }

    // [time][logId][name][level][trace]: message\n
    return fmt.Sprintf("[%s][%s][%s][%s]%s: %s\n",
        item.When.Format("2006/01/02 15:04:05.000"),
        item.LogId,
        item.Name,
        levelToString(item.Level),
        item.Trace,
        item.Message,
    )
}

// console output target
type ConsoleTarget struct {
    Target
}

func (c *ConsoleTarget) Process(item *LogItem) {
    if !c.IsHandling(item.Level) {
        return
    }

    os.Stdout.WriteString(c.Format(item))
}

func (c *ConsoleTarget) Flush(final bool) {
    os.Stdout.Sync()
}

// file output target
type FileTarget struct {
    Target
    filePath      string
    maxLogFile    int
    maxBufferByte int
    maxBufferLine int
    curBufferLine int
    buffer        bytes.Buffer
    file          *os.File
    lastRotate    time.Time
}

func (f *FileTarget) Construct() {
    f.filePath = "@runtime/app.log"
    f.maxLogFile = 10
    f.maxBufferByte = 1 << 20
    f.maxBufferLine = 1000
}

func (f *FileTarget) Init() {
    path := GetAlias(f.filePath)
    h, e := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
    if e != nil {
        panic(fmt.Sprintf("FileTarget: failed to open file: %s, e: %s", path, e))
    }

    stat, e := h.Stat()
    if e != nil {
        panic(fmt.Sprintf("FileTarget: failed to stat file: %s, e: %s", path, e))
    }

    f.file = h
    f.curBufferLine = 0
    f.lastRotate = stat.ModTime()
    f.buffer.Grow(f.maxBufferByte)
}

func (f *FileTarget) SetFilePath(filePath string) {
    f.filePath = filePath
}

func (f *FileTarget) SetMaxLogFile(maxLogFile int) {
    f.maxLogFile = maxLogFile
}

func (f *FileTarget) SetMaxBufferByte(maxBufferByte int) {
    f.maxBufferByte = maxBufferByte
}

func (f *FileTarget) SetMaxBufferLine(maxBufferLine int) {
    f.maxBufferLine = maxBufferLine
}

func (f *FileTarget) Process(item *LogItem) {
    if !f.IsHandling(item.Level) {
        return
    }

    // rotate file by day
    if item.When.Day() != f.lastRotate.Day() {
        f.rotate(item.When)
    }

    // write log to buffer
    f.buffer.WriteString(f.Format(item))
    f.curBufferLine++

    // flush buffer to file
    if f.curBufferLine >= f.maxBufferLine || f.buffer.Len() >= f.maxBufferByte {
        f.Flush(false)
    }
}

func (f *FileTarget) Flush(final bool) {
    f.curBufferLine = 0

    if f.file != nil {
        f.buffer.WriteTo(f.file)
    }

    f.buffer.Reset()

    if final {
        f.file.Close()
        f.file = nil
    }
}

func (f *FileTarget) rotate(now time.Time) {
    defer func() {
        f.lastRotate = now
    }()

    // flush and close file
    f.Flush(true)

    // move current file to backup file
    path := GetAlias(f.filePath)
    date := f.lastRotate.Format("20060102")
    newPath := fmt.Sprintf("%s.%s", path, date)
    os.Rename(path, newPath)

    // clean backup file
    backups, _ := filepath.Glob(path + ".*")
    if len(backups) > 0 {
        for _, backup := range backups {
            ext := filepath.Ext(backup)
            d, e := time.ParseInLocation("20060102", ext[1:], now.Location())
            if e == nil && int(now.Sub(d).Hours()/24) > f.maxLogFile {
                os.Remove(backup)
            }
        }
    }

    // recreate log file
    h, e := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
    if e != nil {
        panic(fmt.Sprintf("FileTarget: failed to open file: %s, e: %s", path, e))
    }

    f.file = h
}
