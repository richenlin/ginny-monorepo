# ginny-cli
Ginny command line tool.


## Install

系统需要安装 Go，在 $GOPATH 目录下执行以下命令安装。成功后一个名为 ginny 的二进制可执行文件会被安装至 $GOPATH/bin/ 文件夹中。

*注意： 该命令行工具支持MacOs、Linux系统，windows系统请使用gitBash或者Cygwin*

方法一： 

```sh
go install github.com/goriller/ginny-cli/ginny@latest

```

方法二：编译安装

```sh
git clone github.com/goriller/ginny-cli.git

# mac:
cd ginny-cli/ginny && CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ginny
cp -f ginny $GOPATH/bin/

# linux
cd ginny-cli/ginny && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ginny
cp -f ginny $GOPATH/bin/


# windows
cd ginny-cli/ginny && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ginny
cp -f ginny $GOPATH/bin/
```
## Dependencies

* protoc：

    Mac OS:

  ```sh
  brew install protoc
  ```
  >如果没有安装homebrew,也可以使用下面下载二进制包的方法

  Windwos及其他系统：
  
  >进入[protobuf release](https://github.com/protocolbuffers/protobuf/>releases) 页面，选择适合自己操作系统的压缩包文件下载
  >解压缩后将bin/protoc二进制文件移动到环境变量的任意path下，如$GOPATH/bin，这里不建议直接将其和系统的一下path放在一起。

* protoc-gen-go：

  ```sh
  // mac
  cd $GOPATH && go install github.com/golang/protobuf/protoc-gen-go@latest
  ```

* go wire:
  
  ```sh
  cd $GOPATH && go get github.com/google/wire/cmd/wire
  ```

* protoc-gen-validate:
  
  ```sh
  cd $GOPATH && go install github.com/envoyproxy/protoc-gen-validate@latest
  ```

* goimports：

  ```sh
  cd $GOPATH && go get golang.org/x/tools/cmd/goimports
  ```

* mockgen：

  参考 https://github.com/golang/mock 进行安装

  ```sh
  cd $GOPATH && go install github.com/golang/mock/mockgen@v1.6.0
  ```

* make 
  
  >Mac OS以及Linux系统一般都具备make命令，解决windows执行make命令的问题:
  >[How to run "make" command in gitbash in windows?](https://gist.github.com/evanwill/0207876c3243bbb6863e65ec5dc3f058)

## Usage
```sh

Command line tool for Ginny project bestpractice

Usage:
  ginny [flags]
  ginny [command]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  help        Help about any command
  new         Create a new Ginny project
  component   Create component file
  service     Create service file by proto file

Flags:
  -h, --help   help for ginny

Use "ginny [command] --help" for more information about a command.
```
### 创建项目

根据 [ginny-template](https://github.com/goriller/ginny-template) 模板创建新项目

```sh
Create a new Ginny project from template

Usage:
  ginny new [flags]

Flags:
      --grpc            Create a grpc service project
  -h, --help            help for new
  -m, --module string   Define the project module, ex: github.com/demo
```

可根据参数创建http 或者 grpc服务的项目，默认创建http 项目

```sh
$ ginny new hellodemo

```
还可以定制项目 module地址:
```sh

$ ginny new hellodemo -m github.com/xxx/hellodemo
```

### proto服务层service

```sh
$ ginny service 

```
创建的service文件在项目 internal/service目录。proto协议增加接口后，可以重复执行此命令生成方法实现模板。

### 创建其他扩展

例如 repo、logic等:

```sh
$ ginny component user -t repo

```
创建的user_repo文件在项目 internal/repo目录

