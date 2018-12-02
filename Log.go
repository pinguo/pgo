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

    rotateNone   = 0
    rotateHourly = 1
    rotateDaily  = 2
)

// LevelToString convert int level to string
func LevelToString(level int) string {
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

// StringToLevel convert string to int level
func StringToLevel(level string) int {
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
        levels |= StringToLevel(v)
    }

    return levels
}

// LogItem represent an item of log
type LogItem struct {
    When    time.Time
    Level   int
    Name    string
    LogId   string
    Trace   string
    Message string
}

// Log the log component, configuration:
// log:
//     levels: "ALL"
//     traceLevels: "DEBUG"
//     chanLen: 1000
//     flushInterval: "60s"
//     targets:
//         info:
//             class: "@pgo/FileTarget"
//             levels: "DEBUG,INFO,NOTICE"
//             filePath: "@runtime/info.log"
//             maxLogFile: 10
//             rotate: "daily"
//         error: {
//             class: "@pgo/FileTarget"
//             levels: "WARN,ERROR,FATAL"
//             filePath: "@runtime/error.log"
//             maxLogFile: 10
//             rotate: "daily"
type Log struct {
    levels        int
    chanLen       int
    traceLevels   int
    flushInterval time.Duration
    targets       map[string]ITarget
    msgChan       chan *LogItem
    wg            sync.WaitGroup
}

func (d *Log) Construct() {
    d.levels = LevelAll
    d.chanLen = 1000
    d.traceLevels = LevelDebug
    d.flushInterval = 60 * time.Second
}

func (d *Log) Init() {
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

// SetLevels set levels to handle, default "ALL"
func (d *Log) SetLevels(v interface{}) {
    if _, ok := v.(string); ok {
        d.levels = parseLevels(v.(string))
    } else if _, ok := v.(int); ok {
        d.levels = v.(int)
    } else {
        panic(fmt.Sprintf("Log: invalid levels: %v", v))
    }
}

// SetChanLen set length of log channel, default 1000
func (d *Log) SetChanLen(len int) {
    d.chanLen = len
}

// SetTraceLevels set levels to trace, default "DEBUG"
func (d *Log) SetTraceLevels(v interface{}) {
    if _, ok := v.(string); ok {
        d.traceLevels = parseLevels(v.(string))
    } else if _, ok := v.(int); ok {
        d.traceLevels = v.(int)
    } else {
        panic(fmt.Sprintf("Log: invalid trace levels: %v", v))
    }
}

// SetFlushInterval set interval to flush log, default "60s"
func (d *Log) SetFlushInterval(v string) {
    if flushInterval, err := time.ParseDuration(v); err != nil {
        panic(fmt.Sprintf("Log: parse flushInterval error, val:%s, err:%s", v, err.Error()))
    } else {
        d.flushInterval = flushInterval
    }
}

// SetTargets set output target, ConsoleTarget will be used if no targets specified
func (d *Log) SetTargets(targets map[string]interface{}) {
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

// GetLogger get a new logger with name and id specified
func (d *Log) GetLogger(name, logId string) *Logger {
    return &Logger{name, logId, d}
}

// GetProfiler get a new profiler
func (d *Log) GetProfiler() *Profiler {
    return &Profiler{}
}

// Flush close msg chan and wait loop end
func (d *Log) Flush() {
    close(d.msgChan)
    d.wg.Wait()
}

func (d *Log) isHandling(level int) bool {
    return level&d.levels != 0
}

func (d *Log) addItem(item *LogItem) {
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

func (d *Log) loop() {
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

// Logger
type Logger struct {
    name  string
    logId string
    log   *Log
}

func (l *Logger) init(name, logId string, log *Log) {
    l.name, l.logId, l.log = name, logId, log
}

func (l *Logger) logMsg(level int, format string, v ...interface{}) {
    if !l.log.isHandling(level) {
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

    l.log.addItem(item)
}

func (l *Logger) Debug(format string, v ...interface{}) {
    l.logMsg(LevelDebug, format, v...)
}

func (l *Logger) Info(format string, v ...interface{}) {
    l.logMsg(LevelInfo, format, v...)
}

func (l *Logger) Notice(format string, v ...interface{}) {
    l.logMsg(LevelNotice, format, v...)
}

func (l *Logger) Warn(format string, v ...interface{}) {
    l.logMsg(LevelWarn, format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
    l.logMsg(LevelError, format, v...)
}

func (l *Logger) Fatal(format string, v ...interface{}) {
    l.logMsg(LevelFatal, format, v...)
}

// Profiler
type Profiler struct {
    pushLog      []string
    counting     map[string][2]int
    profile      map[string][2]int
    profileStack map[string]time.Time
}

func (p *Profiler) reset() {
    p.pushLog = nil
    p.counting = nil
    p.profile = nil
    p.profileStack = nil
}

// PushLog add push log, the push log string is key=Util.ToString(v)
func (p *Profiler) PushLog(key string, v interface{}) {
    if p.pushLog == nil {
        p.pushLog = make([]string, 0)
    }

    pl := key + "=" + Util.ToString(v)
    p.pushLog = append(p.pushLog, pl)
}

// Counting add counting info, the counting string is key=sum(hit)/sum(total)
func (p *Profiler) Counting(key string, hit, total int) {
    if p.counting == nil {
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

// ProfileStart mark start of profile
func (p *Profiler) ProfileStart(key string) {
    if p.profileStack == nil {
        p.profileStack = make(map[string]time.Time)
    }

    p.profileStack[key] = time.Now()
}

// ProfileStop mark stop of profile
func (p *Profiler) ProfileStop(key string) {
    if startTime, ok := p.profileStack[key]; ok {
        delete(p.profileStack, key)
        p.ProfileAdd(key, time.Now().Sub(startTime))
    }
}

// ProfileAdd add profile info, the profile string is key=sum(elapse)/count
func (p *Profiler) ProfileAdd(key string, elapse time.Duration) {
    if p.profile == nil {
        p.profile = make(map[string][2]int)
    }

    v, _ := p.profile[key]
    v[0] += int(elapse.Nanoseconds() / 1e6)
    v[1] += 1

    p.profile[key] = v
}

// GetPushLogString get push log string
func (p *Profiler) GetPushLogString() string {
    if len(p.pushLog) == 0 {
        return ""
    }

    return strings.Join(p.pushLog, " ")
}

// GetCountingString get counting info string
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

// GetProfileString get profile info string
func (p *Profiler) GetProfileString() string {
    if len(p.profile) == 0 {
        return ""
    }

    ps := make([]string, 0)
    for k, v := range p.profile {
        ps = append(ps, fmt.Sprintf("%s=%dms/%d", k, v[0], v[1]))
    }

    return strings.Join(ps, " ")
}

// Target base class of output
type Target struct {
    levels    int
    formatter IFormatter
}

// SetLevels set levels for target, eg. "DEBUG,INFO,NOTICE"
func (t *Target) SetLevels(v interface{}) {
    if _, ok := v.(string); ok {
        t.levels = parseLevels(v.(string))
    } else if _, ok := v.(int); ok {
        t.levels = v.(int)
    } else {
        panic(fmt.Sprintf("Target: invalid levels: %v", v))
    }
}

// SetFormatter set user-defined log formatter, eg. "Lib/Log/Formatter"
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

// IsHandling check whether this target is handling the log item
func (t *Target) IsHandling(level int) bool {
    return t.levels&level != 0
}

// Format format log item to string
func (t *Target) Format(item *LogItem) string {
    // call user-defined formatter if exists
    if t.formatter != nil {
        return t.formatter.Format(item)
    }

    // default log format: [time][logId][name][level][trace]: message\n
    return fmt.Sprintf("[%s][%s][%s][%s]%s: %s\n",
        item.When.Format("2006/01/02 15:04:05.000"),
        item.LogId,
        item.Name,
        LevelToString(item.Level),
        item.Trace,
        item.Message,
    )
}

// ConsoleTarget target for console
type ConsoleTarget struct {
    Target
}

func (c *ConsoleTarget) Construct() {
    c.levels = LevelAll
}

// Process write log to stdout
func (c *ConsoleTarget) Process(item *LogItem) {
    if !c.IsHandling(item.Level) {
        return
    }

    os.Stdout.WriteString(c.Format(item))
}

// Flush flush log to stdout
func (c *ConsoleTarget) Flush(final bool) {
    os.Stdout.Sync()
}

// FileTarget target for file, configuration:
// info:
//     class: "@pgo/FileTarget"
//     levels: "DEBUG,INFO,NOTICE"
//     filePath: "@runtime/info.log"
//     maxLogFile: 10
//     maxBufferByte: 10485760
//     maxBufferLine: 10000
//     rotate: "daily"
type FileTarget struct {
    Target
    filePath      string
    maxLogFile    int
    maxBufferByte int
    maxBufferLine int
    rotate        int

    buffer        bytes.Buffer
    lastRotate    time.Time
    curBufferLine int
}

func (f *FileTarget) Construct() {
    f.filePath = "@runtime/app.log"
    f.maxLogFile = 10
    f.maxBufferByte = 10 * 1024 * 1024
    f.maxBufferLine = 10000
    f.rotate = rotateDaily
    f.levels = LevelAll
}

func (f *FileTarget) Init() {
    f.filePath = GetAlias(f.filePath)
    h, e := os.OpenFile(f.filePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
    if e != nil {
        panic(fmt.Sprintf("FileTarget: failed to open file: %s, e: %s", f.filePath, e))
    }

    defer h.Close()

    stat, e := h.Stat()
    if e != nil {
        panic(fmt.Sprintf("FileTarget: failed to stat file: %s, e: %s", f.filePath, e))
    }

    f.curBufferLine = 0
    f.lastRotate = stat.ModTime()
    f.buffer.Grow(f.maxBufferByte)
}

// SetFilePath set file path, default "@runtime/app.log"
func (f *FileTarget) SetFilePath(filePath string) {
    f.filePath = filePath
}

// SetMaxLogFile set max log backups, default 10
func (f *FileTarget) SetMaxLogFile(maxLogFile int) {
    f.maxLogFile = maxLogFile
}

// SetMaxBufferByte set max buffer bytes, default 10MB
func (f *FileTarget) SetMaxBufferByte(maxBufferByte int) {
    f.maxBufferByte = maxBufferByte
}

// SetMaxBufferLine set max buffer lines, default 10000
func (f *FileTarget) SetMaxBufferLine(maxBufferLine int) {
    f.maxBufferLine = maxBufferLine
}

// SetRotate set rotate policy(none, hourly, daily), default "daily"
func (f *FileTarget) SetRotate(rotate string) {
    switch strings.ToUpper(rotate) {
    case "NONE":
        f.rotate = rotateNone
    case "HOURLY":
        f.rotate = rotateHourly
    case "DAILY":
        f.rotate = rotateDaily
    default:
        panic("FileTarget: invalid rotate:" + rotate)
    }
}

// Process check and rotate log file if rotate is enable,
// write log to buffer, flush buffer to file if buffer is full.
func (f *FileTarget) Process(item *LogItem) {
    if !f.IsHandling(item.Level) {
        return
    }

    // rotate log file
    if f.shouldRotate(item.When) {
        f.rotateLog(item.When)
    }

    // write log to buffer
    f.buffer.WriteString(f.Format(item))
    f.curBufferLine++

    // flush buffer to file
    if f.curBufferLine >= f.maxBufferLine || f.buffer.Len() >= f.maxBufferByte {
        f.Flush(false)
    }
}

// Flush flush log buffer to file
func (f *FileTarget) Flush(final bool) {
    // nothing to flush
    if f.curBufferLine == 0 {
        return
    }

    // open log file to write
    h, e := os.OpenFile(f.filePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
    if e != nil {
        panic(fmt.Sprintf("FileTarget: failed to open file: %s, e: %s", f.filePath, e))
    }

    defer h.Close()

    // write log buffer to file
    f.buffer.WriteTo(h)
    f.buffer.Reset()
    f.curBufferLine = 0
}

func (f *FileTarget) shouldRotate(now time.Time) bool {
    if f.rotate == rotateHourly {
        return now.Hour() != f.lastRotate.Hour() || now.Day() != f.lastRotate.Day()
    } else if f.rotate == rotateDaily {
        return now.Day() != f.lastRotate.Day()
    }

    return false
}

func (f *FileTarget) rotateLog(now time.Time) {
    layout, interval := "", time.Duration(0)
    if f.rotate == rotateHourly {
        layout = "2006010215"
        interval = time.Hour
    } else if f.rotate == rotateDaily {
        layout = "20060102"
        interval = time.Hour * 24
    } else {
        return
    }

    // flush and close file
    f.Flush(true)

    // move current file to backup file
    suffix := f.lastRotate.Format(layout)
    newPath := fmt.Sprintf("%s.%s", f.filePath, suffix)
    os.Rename(f.filePath, newPath)

    // update last rotate time
    f.lastRotate = now

    // clean backup file
    backups, _ := filepath.Glob(f.filePath + ".*")
    if len(backups) > 0 {
        for _, backup := range backups {
            ext := filepath.Ext(backup)
            d, e := time.ParseInLocation(layout, ext[1:], now.Location())
            if e == nil && int(now.Sub(d)/interval) > f.maxLogFile {
                os.Remove(backup)
            }
        }
    }
}
