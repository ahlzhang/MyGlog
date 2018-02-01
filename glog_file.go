// Go support for leveled logs, analogous to https://code.google.com/p/google-glog/
//
// Copyright 2013 Google Inc. All Rights Reserved.
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

// File I/O for logs.

package glog

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// MaxSize is the maximum size of a log file in bytes. 100MB
var MaxSize uint64 = 1024 * 1024 * 100

// logDirs lists the candidate directories for new log files.
var logDirs []string

// If non-empty, overrides the choice of directory in which to write logs.
// See createLogDirs for the full list of possible destinations.
var logDir = flag.String("log_dir", test(), "If non-empty, write log files in this directory")

func test() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println(err)
	}
	dirctory := strings.Replace(dir, "\\", "/", 0)
	return dirctory + "/conf/log/"
}

func createLogDirs() {
	if *logDir != "" {
		logDirs = append(logDirs, *logDir)
	} else {
		logDirs = append(logDirs, os.TempDir())
	}
}

var (
	pid      = os.Getpid()
	program  = filepath.Base(os.Args[0])
	host     = "unknownhost"
	userName = "unknownuser"
)

func init() {
	h, err := os.Hostname()
	if err == nil {
		host = shortHostname(h)
	}

	current, err := user.Current()
	if err == nil {
		userName = current.Username
	}

	// Sanitize userName since it may contain filepath separators on Windows.
	userName = strings.Replace(userName, `\`, "_", -1)
}

// shortHostname returns its argument, truncating at the first period.
// For instance, given "www.google.com" it returns "www".
func shortHostname(hostname string) string {
	if i := strings.Index(hostname, "."); i >= 0 {
		return hostname[:i]
	}
	return hostname
}

// logName returns a new log file name containing tag, with start time t, and
// the name for the symlink for tag.
func logName(tag string, t time.Time) (name, link string) {
	//name = fmt.Sprintf("%s.%s.%s.log.%s.%04d%02d%02d-%02d%02d%02d.%d",
	name = fmt.Sprintf("%s.%s.%04d%02d%02d-%02d%02d%02d",
		program,
		//host,
		//userName,
		tag,
		t.Year(),
		t.Month(),
		t.Day(),
		t.Hour(),
		t.Minute(),
		t.Second()) /**,
		pid)**/
	return name, program + "." + tag
}

func logInitName(tag string) (name, link string) {
	//name = fmt.Sprintf("%s.%s.%s.log.%s.%04d%02d%02d-%02d%02d%02d.%d",
	name = fmt.Sprintf("%s.%s", program, tag)
	return name, tag
}

var onceLogDirs sync.Once

// create creates a new log file and returns the file and its filename, which
// contains tag ("INFO", "FATAL", etc.) and t.  If the file is created
// successfully, create also attempts to update the symlink for that tag, ignoring
// errors.
func create(tag string, t time.Time) (f *os.File, filename string, err error) {
	onceLogDirs.Do(createLogDirs)
	if len(logDirs) == 0 {
		return nil, "", errors.New("log: no log dirs")
	}
	//name, link := logName(tag, t)
	name, link := logInitName(tag)
	var lastErr error
	for _, dir := range logDirs {
		fname := filepath.Join(dir, name)

		if _, lastErr = os.Stat(fname); !os.IsNotExist(lastErr) {
			newName, _ := logName(tag, t)
			os.Rename(fname, filepath.Join(dir, newName))
		}

		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			lastErr = err
			break
		}

		f, err := os.Create(fname)
		if err == nil {
			symlink := filepath.Join(dir, link)

			os.Remove(symlink)        // ignore err
			os.Symlink(name, symlink) // ignore err
			return f, fname, nil
		}
		lastErr = err
	}
	return nil, "", fmt.Errorf("log: cannot create log: %v", lastErr)
}

func openFile(tag string) (*os.File, error) {
	onceLogDirs.Do(createLogDirs)
	if len(logDirs) == 0 {
		return nil, errors.New("log: no log dirs")
	}
	name, _ := logInitName(tag)

	lastErr := errors.New("文件不存在")
	var f *os.File
	for _, dir := range logDirs {
		fname := filepath.Join(dir, name)
		if _, lastErr = os.Stat(fname); !os.IsNotExist(lastErr) {
			f, lastErr = os.OpenFile(fname, os.O_RDWR, 0666)
			if lastErr == nil {
				return f, nil
			}
		}
	}

	return nil, lastErr
}
