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

package middleware

import (
	"crypto/subtle"
	"encoding/binary"
	"log/slog"
	"net"
	"net/http"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"golang.org/x/time/rate"

	"github.com/llklkl/webdav/conf"
	"github.com/llklkl/webdav/internal/model"
)

const MaxUsernamePasswordLength = 256

type ip struct {
	hi uint64
	lo uint64
}

func (i ip) IsZero() bool {
	return i.hi == 0 && i.lo == 0
}

type Basic struct {
	users    map[string]string
	security *conf.SecurityConf

	bannedUsers *lru.Cache[string, *rate.Limiter]
	bannedIp    *lru.Cache[ip, *rate.Limiter]
}

func NewBasic(cfg *conf.Conf) *Basic {
	b := &Basic{
		users:    map[string]string{},
		security: cfg.Security,
	}
	b.bannedUsers, _ = lru.New[string, *rate.Limiter](1024)
	b.bannedIp, _ = lru.New[ip, *rate.Limiter](1024)
	for _, u := range cfg.User {
		b.users[u.Username] = u.Credential
	}

	return b
}

func (b *Basic) Name() string {
	return "Basic AUTH"
}

func (b *Basic) Serve(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user *model.User
		var realPwd string
		var username, password string
		var ok bool
		var clientIp = model.GetClientIP(r.Context())
		var cip = b.convertIp(clientIp)

		if b.banClient(cip) {
			slog.Info("client has been banned", slog.String("ip", clientIp.String()))
			goto WrongPassword
		}

		username, password, ok = r.BasicAuth()
		if !ok {
			goto MarkBanned
		}
		if len(username) > MaxUsernamePasswordLength || len(password) > MaxUsernamePasswordLength {
			goto MarkBanned
		}
		if b.banUser(username) {
			slog.Info("user has been banned", slog.String("username", username))
			goto WrongPassword
		}
		realPwd = b.users[username]
		if subtle.ConstantTimeCompare([]byte(realPwd), []byte(password)) == 1 {
			goto Next
		} else {
			goto MarkBanned
		}

	MarkBanned:
		b.markBanned(cip, username)

	WrongPassword:
		w.Header().Set("WWW-Authenticate", "Basic realm=\"webdav\"")
		w.WriteHeader(http.StatusUnauthorized)
		return

	Next:
		user = &model.User{
			Username: username,
		}
		b.unmarkBanned(cip, username)
		next.ServeHTTP(w, r.WithContext(model.SetUser(r.Context(), user)))
	})
}

func (b *Basic) markBanned(cip ip, username string) {
	if b.security.BanIpWrongPwd && !cip.IsZero() {
		if !b.bannedIp.Contains(cip) {
			l := rate.NewLimiter(rate.Every(5*time.Minute/time.Duration(b.security.PasswordRetryPerFiveMinute)), 5)
			b.bannedIp.ContainsOrAdd(cip, l)
		}
	}
	if b.security.BanUserWrongPwd && len(username) > 0 {
		if !b.bannedUsers.Contains(username) {
			l := rate.NewLimiter(rate.Every(5*time.Minute/time.Duration(b.security.PasswordRetryPerFiveMinute)), 5)
			b.bannedUsers.ContainsOrAdd(username, l)
		}
	}
}

func (b *Basic) unmarkBanned(cip ip, username string) {
	if b.security.BanIpWrongPwd {
		b.bannedIp.Remove(cip)
	}
	if b.security.BanUserWrongPwd {
		b.bannedUsers.Remove(username)
	}
}

func (b *Basic) banClient(cip ip) bool {
	if !b.security.BanIpWrongPwd || cip.IsZero() {
		return false
	}
	limiter, ok := b.bannedIp.Get(cip)
	if !ok {
		return false
	}

	return !limiter.Allow()
}

func (b *Basic) banUser(username string) bool {
	if !b.security.BanUserWrongPwd {
		return false
	}
	limiter, ok := b.bannedUsers.Get(username)
	if !ok {
		return false
	}

	return !limiter.Allow()
}

func (b *Basic) convertIp(i net.IP) ip {
	r := ip{}
	if i == nil {
		return r
	}
	if len(i) == net.IPv4len {
		r.lo = uint64(binary.BigEndian.Uint32(i))
	} else if len(i) == net.IPv6len {
		r.lo = binary.BigEndian.Uint64(i[:8])
		r.hi = binary.BigEndian.Uint64(i[8:16])
	}

	return r
}
