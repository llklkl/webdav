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
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"slices"

	"github.com/BurntSushi/toml"
)

type Conf struct {
	HttpEnable     bool   `toml:"http_enable"`
	HttpListen     string `toml:"http_listen"`
	HttpsEnable    bool   `toml:"https_enable"`
	HttpsListen    string `toml:"https_listen"`
	TlsKeyPem      string `toml:"tls_key_pem"`
	TlsCertPem     string `toml:"tls_cert_pem"`
	TlsKeyPemPath  string `toml:"tls_key_pem_path"`
	TlsCertPemPath string `toml:"tls_cert_pem_path"`

	Library []*LibraryConf `toml:"library"`
	Scope   []*ScopeConf   `toml:"scope"`
	User    []*UserConf    `toml:"user"`

	Security *SecurityConf `toml:"security"`
}

type LibraryConf struct {
	Name       string `toml:"name"`
	MountPoint string `toml:"mount_point"`
	Prefix     string `toml:"prefix"`
}

func ValidLibrary(cfg *Conf, conf *LibraryConf) error {
	if conf == nil {
		return errors.New("empty library configure")
	}
	if conf.Name == "" {
		return errors.New("empty library name")
	}
	if conf.MountPoint == "" {
		return fmt.Errorf("the mount point of library[%s] is empty", conf.Name)
	}
	if !filepath.IsAbs(conf.MountPoint) {
		return fmt.Errorf("mount point only support absolute path for library[%s]", conf.Name)
	}
	return nil
}

type ScopeConf struct {
	Name       string   `toml:"name"`
	Library    string   `toml:"library"`
	Include    []string `toml:"include"`
	Exclude    []string `toml:"exclude"`
	Permission []string `toml:"permission"`
}

func ValidScope(cfg *Conf, conf *ScopeConf) error {
	if conf == nil {
		return errors.New("empty scope configure")
	}
	if conf.Name == "" {
		return errors.New("empty scope name")
	}
	if !slices.ContainsFunc(cfg.Library, func(l *LibraryConf) bool { return l.Name == conf.Library }) {
		return fmt.Errorf("library[%s] not found", conf.Library)
	}
	for _, perm := range conf.Permission {
		if !slices.Contains([]string{
			"read",
			"write",
			"delete",
			"create_file",
			"create_folder",
			"rename",
			"*",
		}, perm) {
			return fmt.Errorf("scope[%s] permission[%s] is invalid", conf.Name, perm)
		}
	}

	return nil
}

type UserConf struct {
	Username   string   `toml:"username"`
	Credential string   `toml:"credential"`
	Scope      []string `toml:"scope"`
}

func ValidUser(cfg *Conf, conf *UserConf) error {
	if conf == nil {
		return errors.New("empty user configure")
	}
	if conf.Username == "" {
		return errors.New("empty user name")
	}
	for _, scope := range conf.Scope {
		if !slices.ContainsFunc(cfg.Scope, func(s *ScopeConf) bool { return s.Name == scope }) {
			return fmt.Errorf("the scope[%s] of user[%s] not found", scope, conf.Username)
		}
	}

	return nil
}

type SecurityConf struct {
	PasswordRetryPerFiveMinute int  `toml:"password_retry_per_five_minute"`
	BanUserWrongPwd            bool `toml:"ban_user_wrong_pwd"`
	BanIpWrongPwd              bool `toml:"ban_ip_wrong_pwd"`
}

func Valid(cfg *Conf) error {
	if cfg == nil {
		return errors.New("empty configure")
	}

	if !cfg.HttpEnable && !cfg.HttpsEnable {
		return errors.New("at least one of HTTP or HTTPS should be enabled")
	}
	if cfg.HttpEnable {
		if _, err := url.Parse("http://" + cfg.HttpListen); err != nil {
			return fmt.Errorf("bad format [http_listen]: %w", err)
		}
	}
	if cfg.HttpsEnable {
		if _, err := url.Parse("https://" + cfg.HttpsListen); err != nil {
			return fmt.Errorf("bad format [https_listen]: %w", err)
		}
		if (cfg.TlsCertPemPath == "" || cfg.TlsKeyPemPath == "") &&
			(cfg.TlsCertPem == "" || cfg.TlsKeyPem == "") {
			return errors.New("empty tls certificate")
		}
	}

	for _, lib := range cfg.Library {
		if err := ValidLibrary(cfg, lib); err != nil {
			return err
		}
	}
	for _, scope := range cfg.Scope {
		if err := ValidScope(cfg, scope); err != nil {
			return err
		}
	}
	for _, user := range cfg.User {
		if err := ValidUser(cfg, user); err != nil {
			return err
		}
	}

	return nil
}

func Parse(path string) (*Conf, error) {
	cfg := new(Conf)
	_, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return nil, err
	}
	if err := Valid(cfg); err != nil {
		return nil, err
	}
	if cfg.Security == nil {
		cfg.Security = &SecurityConf{}
	}
	return cfg, nil
}
