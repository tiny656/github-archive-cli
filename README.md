# 上传档案至 GitHub

这个 Go 程序会自动将当前目录下的文件打包为 ZIP 文件，并将 ZIP 文件上传/更新至指定的 GitHub 仓库。

## 用法

1. 首先，需要安装 Go 语言环境
2. 编译代码：在代码所在目录，运行 `go build`，生成可执行文件
3. 运行可执行文件，使用必须的参数 `-username`、`-repo` 和 `-token`。例如：

```shell
./<可执行文件名> -username <您的GitHub用户名> -repo <您的GitHub仓库名> -token <您的GitHub访问令牌>
```

可选参数：

- `-branch`: 可选，指定要提交到的分支，默认为 `main` 分支
- `-message`: 可选，指定提交信息，默认为 `"Automated commit for zipped files"`

## 功能

1. 遍历当前目录下的所有文件和子目录
2. 将文件和子目录添加到 ZIP 文件中（当前程序可执行文件不包含在内）
3. 将 ZIP 文件上传或更新至指定 GitHub 仓库
4. 若文件已存在，程序会更新现有文件；若文件不存在，程序会创建新文件

使用此程序，您可以轻松将本地文件上传或更新至您的 GitHub 仓库。