# osd-tool

腾讯云(cos)或阿里云(oss)的对象存储 目录上传、下载工具，当前只是简单的全量上传和下载，用于跨服务商迁移文件、跨设备迁移文件、内容备份等场景

[![Go](https://github.com/HayWolf/osd-tool/actions/workflows/build.yml/badge.svg?branch=master)](https://github.com/HayWolf/osd-tool/actions/workflows/build.yml)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/HayWolf/osd-tool)](https://github.com/HayWolf/osd-tool/releases)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/HayWolf/osd-tool)](https://github.com/HayWolf/osd-tool/blob/master/go.mod)
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg)](http://copyfree.org)
[![Go Report](https://goreportcard.com/badge/github.com/HayWolf/osd-tool)](https://goreportcard.com/report/github.com/HayWolf/osd-tool)

## 用法

```shell
# 有golang环境时可以通过go install下载安装，或者直接下载当前项目发布的包
go install github.com/HayWolf/osd-tool@latest

# 首次使用时候可以通过init命令初始化配置文件模版，默认生成config.yaml文件
osd-tool init

# 配置好相应的配置内容...

# 把配置文件中配置的upload list上传到对象存储
osd-tool upload

# 把配置文件中配置的download list下载到本地
osd-tool download

# 升级当前程序
osd-tool --upgrade
```

### 存储器类型配置

```yaml
# 存储对象 cos 或者 oss （分别是腾讯云和阿里云）
# 下方需要对应配置 cos或oss的密钥等信息
storage: cos
```

### 上传配置

在配置文件中配置要上传的目录和目标路径，source为本地路径，dest为cos路径。比如下方配置将会把本地的sync1目录下的文件及文件夹上传到COS的/syncTest/dir1目录下：

```yaml
upload:
  list:
    - source: /Users/HayWolf/Downloads/sync1
      dest: /syncTest/dir1
    - source: /Users/HayWolf/Downloads/sync2
      dest: /syncTest/2dir
  ignore: [ .git, .DS_Store ] # 需要忽略的文件和文件夹
```

### 下载配置

在配置文件中配置要下载的目录和目标路径，source为cos路径，dest为本地路径。比如下方配置将会把cos上的syncTest目录下的文件及子目录下载到本地的downloadTest目录下：

```yaml
download:
  list:
    - source: /syncTest
      dest: /Users/HayWolf/Downloads/downloadTest
```

### 对象存储配置

```yaml
osd:
  secret_id:
  secret_key:
  bucket: # 存储桶的名称，注意cos的存储桶名称带有APPID，
  region:  # 替换成存储桶的区域代码，比如Oss的cn-shenzhen，比如Cos的ap-guangzhou
  timeout: 300 #单位：秒
```