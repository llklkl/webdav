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
	"log/slog"

	"github.com/llklkl/webdav/conf"
)

type Scope struct {
	name    string
	include MatchGroup
	exclude MatchGroup
	perm    Perm
}

type ScopeGroup []*Scope

func (s ScopeGroup) Match(path string, perm Perm) (bool, bool) {
	hasMatched := false
	hasPermission := false
	for _, scope := range s {
		matched, permission := scope.Match(path, perm)
		hasMatched = hasMatched || matched
		hasPermission = hasPermission || permission
		if matched && permission {
			break
		}
	}

	return hasMatched, hasPermission
}

func NewScope(scp *conf.ScopeConf) *Scope {
	include, err := NewMatchGroup(scp.Include)
	if err != nil {
		slog.Warn("include syntax error", slog.Any("include", scp.Include), slog.Any("err", err))
	}
	exclude, err := NewMatchGroup(scp.Exclude)
	if err != nil {
		slog.Warn("exclude syntax error", slog.Any("exclude", scp.Exclude), slog.Any("err", err))
	}
	s := &Scope{
		name:    scp.Name,
		include: include,
		exclude: exclude,
	}
	s.perm.FromString(scp.Permission)
	return s
}

func (s *Scope) Match(name string, perm Perm) (matched, permission bool) {
	if !s.perm.Check(perm) {
		return
	}
	permission = true

	if s.exclude.Match(name) {
		return
	}

	if s.include.Match(name) {
		matched = true
		return
	}

	return
}
