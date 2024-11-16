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

package model

import (
	"context"
	"net"
)

type UserCtxKey struct{}
type ClientIPKey struct{}

func SetUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, UserCtxKey{}, user)
}

func GetUser(ctx context.Context) *User {
	v, ok := ctx.Value(UserCtxKey{}).(*User)
	if ok {
		return v
	}

	return nil
}

func SetClientIP(ctx context.Context, ip net.IP) context.Context {
	return context.WithValue(ctx, ClientIPKey{}, ip)
}

func GetClientIP(ctx context.Context) net.IP {
	ip, ok := ctx.Value(ClientIPKey{}).(net.IP)
	if ok {
		return ip
	}
	return nil
}
