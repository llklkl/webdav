## webdav

[简体中文](README.md)｜[English](README.en.md)

一个高性能的 webdav 服务器。支持的特性：

1. 同时挂载多个文件夹；
2. 支持文件夹/文件级别的访问范围、访问权限控制；
3. 支持用户级别访问范围控制；
4. 支持鉴权；
    1. 目前仅支持 http basic 鉴权；
    2. 其他鉴权方式计划支持。。
5. 支持 http/https 独立监听；
6. 支持 ip/用户级别密码防爆破处理

## 配置

### 名词解释

+ library: 资源库。可配置具体需要共享的目录，以及webdav的访问前缀。
+ scope: 访问范围。配置某个资源库的访问范围和权限，多个 scope 可以更精细化的控制权限。
+ user: 用户。访问资源库的用户，用户通过 scope 间接拥有资源库的访问权限，每个用户可以设置多个访问范围。

### 配置示例

```toml
# 开启 http 监听
http_enable = true
# http 监听地址
http_listen = "127.0.0.1:8080"
# 开启 https 监听
https_enable = true
# https 监听地址
https_listen = "127.0.0.1:8443"
# 指定证书和私钥, base64格式. 两者二选一
tls_key_pem = ""
tls_cert_pem = ""
# 指定证书和私钥, 文件路径
tls_key_pem_path = "pri.key"
tls_cert_pem_path = "cert.crt"

# 设置资源库
[[library]]
# 资源库名称
name = "media"
# 资源库挂载的文件夹路径，必须是绝对路径
mount_point = "/data/media"
# webdav 前缀
prefix = "webdav"

[[library]]
name = "backup"
mount_point = "/data/backup"
prefix = "webdav2"

# 设置访问范围
[[scope]]
# 访问范围名称
name = "media"
# 关联的资源库
library = "media"
# 允许指定的文件夹/文件用 dir: 来定义
# 允许指定文件后缀用 file:*.xxx 来定义
include = [
    "dir:/music",
    "dir:/vedio",
    "file:*.mp4"
]
# 排除的文件/文件夹，格式与 include 一致
exclude = [
    "dir:/xxx",
    "file:xxx.mp4"
]
# 访问权限。全部开启可用 * 代替
permission = ["read", "write", "create_file", "create_folder", "rename"]

[[scope]]
name = "backup"
library = "backup"
include = [
    "dir:/"
]
exclude = [
]
# 开启全部权限
permission = ["*"]

# 配置访问用户
[[user]]
# 用户名
username = "test"
# 密码
credential = "test"
# 用户可访问的范围
scope = ["media", "backup"]

# 安全配置
[security]
# 5 分钟内，可重试密码的次数
password_retry_per_five_minute = 10
# 超过重试次数之后，是否要禁用该用户
ban_user_wrong_pwd = true
# 超过重试次数之后，是否要禁用 ip
ban_ip_wrong_pwd = true
```

## 运行

### 二进制运行

```shell
webdav serve -c /path/to/config.toml
```

### 检查配置是否正确

```shell
webdav verify -c /path/to/config.toml
```