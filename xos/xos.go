/*
 * Copyright 2012-2019 Li Kexian
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * A toolkit for Golang development
 * https://www.likexian.com/
 */

package xos

import (
	"bytes"
	"context"
	"fmt"
	"github.com/likexian/gokit/xfile"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
	"time"
)

// Version returns package version
func Version() string {
	return "0.2.0"
}

// Author returns package author
func Author() string {
	return "[Li Kexian](https://www.likexian.com/)"
}

// License returns package license
func License() string {
	return "Licensed under the Apache License 2.0"
}

// Exec exec command and returns
func Exec(cmd string, args ...string) (stdout, stderr string, err error) {
	bufOut := new(bytes.Buffer)
	bufErr := new(bytes.Buffer)

	c := exec.Command(cmd, args...)
	c.Stdout = bufOut
	c.Stderr = bufErr
	err = c.Run()

	return bufOut.String(), bufErr.String(), err
}

// TimeoutExec exec command with timeout and returns
func TimeoutExec(timeout int, cmd string, args ...string) (stdout, stderr string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	bufOut := new(bytes.Buffer)
	bufErr := new(bytes.Buffer)

	c := exec.Command(cmd, args...)
	c.Stdout = bufOut
	c.Stderr = bufErr
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err = c.Start(); err != nil {
		return
	}

	end := make(chan bool, 1)
	defer close(end)

	go func() {
		select {
		case <-ctx.Done():
			syscall.Kill(-c.Process.Pid, syscall.SIGKILL)
			return
		case <-end:
			return
		}
	}()

	if err = c.Wait(); err != nil {
		return
	}

	return bufOut.String(), bufErr.String(), err
}

// SetUser Set process user
func SetUser(user string) (err error) {
	uid, gid, err := LookupUser(user)
	if err != nil {
		return
	}

	err = SetGid(gid)
	if err == nil {
		err = SetUid(uid)
	}

	return
}

// LookupUser returns the uid and gid of user
func LookupUser(name string) (uid, gid int, err error) {
	u, err := user.Lookup(name)
	if err != nil {
		return
	}

	uid, err = strconv.Atoi(u.Uid)
	if err == nil {
		gid, err = strconv.Atoi(u.Gid)
	}

	return
}

// SetUid set the uid of process
func SetUid(uid int) (err error) {
	_, _, errno := syscall.RawSyscall(syscall.SYS_SETUID, uintptr(uid), 0, 0)
	if errno != 0 {
		err = errno
	}

	return
}

// SetGid set the gid of process
func SetGid(gid int) (err error) {
	_, _, errno := syscall.RawSyscall(syscall.SYS_SETGID, uintptr(gid), 0, 0)
	if errno != 0 {
		err = errno
	}

	return
}

// WritePid write pid to file path
func WritePid(path string) error {
	return xfile.WriteText(path, fmt.Sprintf("%d\n", os.Getpid()))
}
