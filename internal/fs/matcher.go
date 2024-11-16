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
	"errors"
	"path/filepath"
	"strings"
)

type Matcher interface {
	Match(path string) bool
}

type MatchGroup []Matcher

func NewMatchGroup(patterns []string) (MatchGroup, error) {
	var firstErr error
	mg := make(MatchGroup, 0, len(patterns))
	for _, p := range patterns {
		if m, err := NewMatcher(p); err == nil {
			mg = append(mg, m)
		} else if firstErr == nil {
			firstErr = err
		}
	}

	return mg, firstErr
}

func (m MatchGroup) Match(path string) bool {
	for i := range m {
		if m[i].Match(path) {
			return true
		}
	}

	return false
}

func NewMatcher(pattern string) (Matcher, error) {
	if strings.HasPrefix(pattern, "file:") {
		pattern = pattern[5:]
		return newMatchFile(pattern), nil
	} else if strings.HasPrefix(pattern, "dir:") {
		pattern = pattern[4:]
		return newMatchDir(pattern), nil
	}

	return nil, errors.New("unsupported pattern")
}

type MatchFile struct {
	ext     bool
	pattern string
}

func newMatchFile(pattern string) *MatchFile {
	if strings.HasPrefix(pattern, "*.") {
		return &MatchFile{
			ext:     true,
			pattern: pattern[1:],
		}
	} else {
		return &MatchFile{
			ext:     false,
			pattern: pattern,
		}
	}
}

func (m *MatchFile) Match(path string) bool {
	if m.ext {
		return strings.EqualFold(filepath.Ext(path), m.pattern)
	} else {
		return path == m.pattern
	}
}

type MatchDir string

func newMatchDir(pattern string) MatchDir {
	pattern = clearPath(pattern)
	return MatchDir(pattern)
}

func (m MatchDir) Match(path string) bool {
	return strings.HasPrefix(path, string(m))
}
