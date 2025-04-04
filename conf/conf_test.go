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

package conf

import (
	"os"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestParse(t *testing.T) {
	cfg := &Conf{
		HttpEnable:     false,
		HttpListen:     "",
		HttpsEnable:    false,
		HttpsListen:    "",
		TlsKeyPem:      "",
		TlsCertPem:     "",
		TlsKeyPemPath:  "",
		TlsCertPemPath: "",
		Library: []*LibraryConf{
			{
				Name:       "",
				MountPoint: "",
				Prefix:     "",
			},
		},
		Scope: []*ScopeConf{
			{
				Name:       "",
				Library:    "",
				Include:    nil,
				Exclude:    nil,
				Permission: nil,
			},
		},
		User: []*UserConf{
			{
				Username:   "",
				Credential: "",
				Scope:      nil,
			},
		},
		Security: &SecurityConf{
			PasswordRetryPerFiveMinute: 0,
			BanUserWrongPwd:            false,
			BanIpWrongPwd:              false,
		},
	}

	data, _ := toml.Marshal(cfg)
	fp, err := os.OpenFile("./config.example.toml", os.O_RDWR|os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		t.Log(err)
		return
	}
	fp.Write(data)
	fp.Sync()
	fp.Close()
}
