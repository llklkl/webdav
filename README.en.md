## webdav

[简体中文](README.md)｜[English](README.en.md)

A high-performance WebDAV server. Supported features:

1. Mount multiple folders simultaneously
2. Support folder/file-level access scope and access permission control
3. Support user-level access scope control
4. Support authentication
    1. Currently only supports HTTP basic authentication
    2. Other authentication methods are planned to be supported
5. Support independent listening of HTTP/HTTPS
6. Support IP/user-level password anti-burst processing

## Configuration

### Noun Explanation

+ library: Resource library. Can configure specific directories that need to be shared, as well as the WebDAV access
  prefix.
+ scope: Access scope. Configure the access scope and permissions of a resource library, and multiple scopes can control
  permissions more finely.
+ user: User. Users who access the resource library, users indirectly have access to the resource library through the
  scope, and each user can set multiple access scopes.

### Configuration Example

```toml
# Enable HTTP listening
http_enable = true
# HTTP listening address
http_listen = "127.0.0.1:8080"
# Enable HTTPS listening
https_enable = true
# HTTPS listening address
https_listen = "127.0.0.1:8443"
# Specify certificate and private key, base64 format. Choose one of the two
tls_key_pem = ""
tls_cert_pem = ""
# Specify certificate and private key, file path
tls_key_perm_path = "pri.key"
tls_cert_perm_path = "cert.crt"

# Set resource library
[[library]]
# Resource library name
name = "media"
# Folder path where the resource library is mounted, must be an absolute path
mount_point = "/data/media"
# WebDAV prefix
prefix = "webdav"

[[library]]
name = "backup"
mount_point = "/data/backup"
prefix = "webdav2"

# Set access scope
[[scope]]
# Access scope name
name = "media"
# Associated resource library
library = "media"
# Allowed specified folders/files are defined with dir:
# Allowed specified file suffixes are defined with file:*.xxx
include = [
    "dir:/music",
    "dir:/vedio",
    "file:*.mp4"
]
# Excluded files/folders, the format is consistent with include
exclude = [
    "dir:/xxx",
    "file:xxx.mp4"
]
# Access permissions. All can be enabled with * instead
permission = ["read", "write", "create_file", "create_folder", "rename"]

[[scope]]
name = "backup"
library = "backup"
include = [
    "dir:/"
]
exclude = [
]
# Enable all permissions
permission = ["*"]

# Configure access users
[[user]]
# Username
username = "test"
# Password
credential = "test"
# User's accessible scopes
scope = ["media", "backup"]

# Security configuration
[security]
# Number of password retries allowed within 5 minutes
password_retry_per_five_minute = 10
# Whether to ban the user after exceeding the retry count
ban_user_wrong_pwd = true
# Whether to ban the IP after exceeding the retry count
ban_ip_wrong_pwd = true
```

## Running

### Binary run

```shell
webdav serve -c /path/to/config.toml
```

### Check if the configuration is correct

```shell
webdav verify -c /path/to/config.toml
```