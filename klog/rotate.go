package klog

import (
	"Lplot/utils"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"
)

const (
	// number of seconds in a day
	secondInDay = 24 * 60 * 60
	// number of seconds in a hour
	secondInHour = 60 * 60
	// number of seconds in a minute
	secondInMinute = 60
)

// This log writer sends output to a file
type TimeFileLogWriter struct {
	basename    string // without path info
	filename    string // include filepath, might abs path or relative path
	absFileName string // abs path
	file        *os.File

	rotateSwitch bool
	rotateUnit   string // 'D', 'H', 'M'
	backupCount  int    // If backupCount is > 0, when rollover is done,

	interval   int64
	suffix     string         // suffix of log file
	fileFilter *regexp.Regexp // for removing old log files

	rolloverAt int64 // time.Unix()
}

func NewTimeFileLogWriter(fName string) *TimeFileLogWriter {
	w := &TimeFileLogWriter{
		basename:     fName,
		filename:     filepath.Join(logConfig.Path, fName),
		rotateSwitch: logConfig.RotateSwitch,
		rotateUnit:   logConfig.RotateUnit,
		backupCount:  logConfig.RotateCount,
	}

	// get abs path
	if path, err := filepath.Abs(w.filename); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "NewTimeFileLogWriter(%s): %s\n", w.filename, err)
		return nil
	} else {
		w.absFileName = path
	}

	// prepare for w.interval, w.suffix and w.fileFilter
	w.prepare()

	if w.rotateSwitch {
		if err := w.doRotate(); err != nil {
			//panic(fmt.Errorf("NewTimeFileLogWriter doRotate(%q): %s\n", w.basename, err))
			return nil
		}
	} else {
		if err := w.doCreate(); err != nil {
			//panic(fmt.Errorf("NewTimeFileLogWriter doCreate(%q): %s\n", w.basename, err))
			return nil
		}
	}

	return w
}

func (w *TimeFileLogWriter) Write(p []byte) (n int, err error) {
	// Guard against concurrent writes
	//rl.mutex.Lock()
	//defer rl.mutex.Unlock()

	if w.rotateSwitch && w.shouldRollover() {
		if err := w.doRotate(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "log rotate(%s): %s\n", w.basename, err)
		}
	}

	return w.file.Write(p)
}

func (w *TimeFileLogWriter) prepare() {
	var regRule string

	switch w.rotateUnit {
	case "D": // day
		w.interval = 60 * 60 * 24
		w.suffix = "%Y%m%d"
		regRule = `^\d{4}\d{2}\d{2}$`
	case "M": // minute
		w.interval = 60
		w.suffix = "%Y%m%d%H%M"
		regRule = `^\d{4}\d{2}\d{2}\d{2}\d{2}$`
	case "H": // hour, by default
		fallthrough
	default:
		w.interval = 60 * 60
		w.suffix = "%Y%m%d%H"
		regRule = `^\d{4}\d{2}\d{2}\d{2}$`
	}
	w.fileFilter = regexp.MustCompile(regRule)

	fInfo, err := os.Stat(w.filename)
	var t time.Time
	if err == nil {
		t = fInfo.ModTime() // 最后修改时间
	} else {
		t = time.Now()
	}

	w.rolloverAt = w.computeRollover(t)
}

func (w *TimeFileLogWriter) computeRollover(currTime time.Time) int64 {
	var result int64
	t := currTime.Local()

	if w.rotateUnit == "D" {
		// r is the number of seconds left between now and nextDay
		r := secondInDay - ((t.Hour()*60+t.Minute())*60 + t.Second())
		result = currTime.Unix() + int64(r)
	} else if w.rotateUnit == "H" {
		// r is the number of seconds left between now and the next hour
		r := secondInHour - (t.Minute()*60 + t.Second())
		result = currTime.Unix() + int64(r)
	} else {
		// r is the number of seconds left between now and the next minute
		r := secondInMinute - (t.Second())
		result = currTime.Unix() + int64(r)
	}
	return result
}

func (w *TimeFileLogWriter) shouldRollover() bool {
	t := time.Now().Unix()

	if t >= w.rolloverAt {
		return true
	} else {
		return false
	}
}

// 创建新的文件
func (w *TimeFileLogWriter) doCreate() error {
	if w.file != nil {
		return nil
	}

	// 重新打开文件
	fd, err := os.OpenFile(w.filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	w.file = fd

	return nil
}

// 文件自动切割功能
func (w *TimeFileLogWriter) doRotate() error {
	// Close any log file that may be open
	if w.file != nil {
		_ = w.file.Close()
	}

	if w.shouldRollover() {
		// rename file to backup name
		if err := w.moveToBackup(); err != nil {
			return err
		}
	}

	// remove files, according to backupCount
	if w.backupCount > 0 {
		for _, fileName := range w.getFilesToDelete() {
			_ = os.Remove(fileName)
		}
	}
	// Open the log file
	fd, err := os.OpenFile(w.filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	w.file = fd

	w.adjustRolloverAt()

	return nil
}

// 同种类型的日志按时间排序，保留最新的 log.rotate.count 个日志文件
func (w *TimeFileLogWriter) getFilesToDelete() []string {
	dirName := filepath.Dir(w.filename)
	baseName := filepath.Base(w.filename)

	var result []string
	fileInfos, err := ioutil.ReadDir(dirName)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", w.filename, err)
		return result
	}

	prefix := baseName + "."
	plen := len(prefix)
	for _, fileInfo := range fileInfos {
		fileName := fileInfo.Name()
		if len(fileName) >= plen {
			if fileName[:plen] == prefix {
				suffix := fileName[plen:]
				if w.fileFilter.MatchString(suffix) {
					result = append(result, filepath.Join(dirName, fileName))
				}
			}
		}
	}

	sort.Sort(sort.StringSlice(result))

	if len(result) < w.backupCount {
		result = result[0:0]
	} else {
		result = result[:len(result)-w.backupCount]
	}
	return result
}

// rename file to backup name
func (w *TimeFileLogWriter) moveToBackup() error {
	_, err := os.Lstat(w.filename)
	if err != nil {
		return nil
	}
	// file exists

	// get the time that this sequence started at and make it a TimeTuple
	t := time.Unix(w.rolloverAt-w.interval, 0).Local()
	fName := w.absFileName + "." + utils.Format(w.suffix, t)

	// remove the file with fName if exist
	if _, err := os.Stat(fName); err == nil {
		err = os.Remove(fName)
		if err != nil {
			return fmt.Errorf("Rotate: %s\n", err)
		}
	}

	// Rename the file to its new found home
	err = os.Rename(w.absFileName, fName)
	if err != nil {
		return fmt.Errorf("Rotate: %s\n", err)
	}
	return nil
}

// 更新下次切割时间
func (w *TimeFileLogWriter) adjustRolloverAt() {
	currTime := time.Now()
	newRolloverAt := w.computeRollover(currTime)

	for newRolloverAt <= currTime.Unix() {
		newRolloverAt = newRolloverAt + w.interval
	}

	w.rolloverAt = newRolloverAt
}
