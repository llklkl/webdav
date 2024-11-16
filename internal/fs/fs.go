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
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"golang.org/x/net/webdav"

	"github.com/llklkl/webdav/conf"
	"github.com/llklkl/webdav/internal/model"
)

func clearPath(path string) string {
	if path == "" || path[0] != '/' {
		path = "/" + path
	}
	return filepath.Clean(path)
}

type Fs struct {
	name       string
	mountPoint string

	root webdav.Dir

	userScope map[string]ScopeGroup
}

func NewFs(cfg *conf.Conf, library *conf.LibraryConf) *Fs {
	fs := &Fs{
		name:       library.Name,
		mountPoint: clearPath(library.MountPoint),
		root:       webdav.Dir(library.MountPoint),
		userScope:  map[string]ScopeGroup{},
	}

	for _, scp := range cfg.Scope {
		if scp.Library != library.Name {
			continue
		}
		for _, user := range cfg.User {
			if !slices.Contains(user.Scope, scp.Name) {
				continue
			}
			fs.userScope[user.Username] = append(fs.userScope[user.Username], NewScope(scp))
		}
	}

	return fs
}

func (f *Fs) getScope(ctx context.Context) ScopeGroup {
	user := model.GetUser(ctx)
	return f.userScope[user.Username]
}

func (f *Fs) checkPermission(ctx context.Context, name string, needPerm Perm) error {
	scope := f.getScope(ctx)
	if matched, permission := scope.Match(name, needPerm); matched && permission {
		return nil
	} else if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		slog.Debug("permission forbidden", slog.String("name", name),
			slog.Bool("matched", matched),
			slog.String("needPerm", needPerm.String()))
	}

	return os.ErrPermission
}

func (f *Fs) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	const needPerm = PermCreateFolder

	name = clearPath(name)
	if err := f.checkPermission(ctx, name, needPerm); err != nil {
		return err
	}

	return f.root.Mkdir(ctx, name, perm)
}

type fileFilter struct {
	webdav.File
	dir   string
	scope ScopeGroup
}

func newFileFilter(dir string, f webdav.File, scope ScopeGroup) *fileFilter {
	return &fileFilter{
		File:  f,
		dir:   dir,
		scope: scope,
	}
}

func (f *fileFilter) Readdir(n int) ([]os.FileInfo, error) {
	infos, err := f.File.Readdir(n)
	if err != nil {
		return nil, err
	}
	filtered := infos[:0]
	for i := range infos {
		if matched, _ := f.scope.Match(filepath.Join(f.dir, infos[i].Name()), PermRead); matched {
			filtered = append(filtered, infos[i])
		}
	}
	return infos, nil
}

func (f *Fs) filterDir(ctx context.Context, name string) (string, error) {
	name = clearPath(name)
	needPerm := PermRead
	if err := f.checkPermission(ctx, name, needPerm); err != nil {
		return "", err
	}
	return name, nil
}

func (f *Fs) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	name = clearPath(name)

	needPerm := PermRead
	if flag&os.O_WRONLY != 0 {
		needPerm |= PermWrite
	}
	if flag&os.O_RDWR != 0 {
		needPerm |= PermWrite
	}
	if flag&os.O_CREATE != 0 {
		needPerm |= PermCreateFile
	}

	if err := f.checkPermission(ctx, name, needPerm); err != nil {
		return nil, err
	}

	file, err := f.root.OpenFile(ctx, name, flag, perm)
	if err != nil {
		return file, err
	}

	return newFileFilter(filepath.Dir(name), file, f.getScope(ctx)), nil
}

func (f *Fs) RemoveAll(ctx context.Context, name string) error {
	name = clearPath(name)
	needPerm := PermDelete

	if err := f.checkPermission(ctx, name, needPerm); err != nil {
		return err
	}

	return f.root.RemoveAll(ctx, name)
}

func (f *Fs) Rename(ctx context.Context, oldName, newName string) error {
	oldName = clearPath(oldName)
	needPerm := PermRename

	if err := f.checkPermission(ctx, oldName, needPerm); err != nil {
		return err
	}

	return f.root.Rename(ctx, oldName, newName)
}

func (f *Fs) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	name = clearPath(name)
	needPerm := PermRead

	if err := f.checkPermission(ctx, name, needPerm); err != nil {
		return nil, err
	}

	return f.root.Stat(ctx, name)
}
