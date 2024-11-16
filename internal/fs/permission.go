/*
 * Copyright (c) 2024 llklkl
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */

package fs

import (
	"strings"
)

type Perm int

const (
	PermNone         Perm = 0
	PermRead         Perm = 1 << 1
	PermWrite        Perm = 1 << 2
	PermDelete       Perm = 1 << 3
	PermCreateFolder Perm = 1 << 4
	PermCreateFile   Perm = 1 << 5
	PermRename       Perm = 1 << 6
)

func (p *Perm) String() string {
	b := strings.Builder{}
	b.Grow(64)
	w := func(s string) {
		if b.Len() > 0 {
			b.WriteByte('|')
		}
		b.WriteString(s)
	}
	var s string
	for _, x := range []Perm{PermRead, PermWrite, PermDelete, PermCreateFolder, PermCreateFile, PermRename} {
		if x&*p == 0 {
			continue
		}
		switch x {
		case PermRead:
			s = "read"
		case PermWrite:
			s = "write"
		case PermDelete:
			s = "delete"
		case PermCreateFile:
			s = "create_file"
		case PermCreateFolder:
			s = "create_folder"
		case PermRename:
			s = "rename"
		}
		w(s)
	}
	return b.String()
}

func (p *Perm) FromString(perms []string) {
	for _, perm := range perms {
		switch perm {
		case "read":
			*p |= PermRead
		case "write":
			*p |= PermWrite
		case "delete":
			*p |= PermDelete
		case "create_file":
			*p |= PermCreateFolder
		case "create_folder":
			*p |= PermCreateFolder
		case "rename":
			*p |= PermRename
		case "*":
			*p |= PermRead | PermWrite | PermDelete | PermCreateFile | PermCreateFolder | PermRename
		}
	}
}

func (p *Perm) Check(needPerm Perm) bool {
	return *p&needPerm == needPerm
}
