package main

import (
	"archive/zip"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func main() {
	args := parseCommandLineArgs()
	zipFileName, selfName := prepareFileInfo()
	zipBuf := createZipBuffer(zipFileName, selfName)
	// saveZipFileToLocal(zipFileName, zipBuf)
	ts, tc, client := setupGitHubClient(args.token)
	commitZipToRepo(args, zipFileName, zipBuf, ts, tc, client)
}

func saveZipFileToLocal(zipFileName string, buf *bytes.Buffer) {
	err := ioutil.WriteFile(zipFileName, buf.Bytes(), 0644)
	if err != nil {
		log.Fatalf("无法将ZIP文件 %s 保存到本地。错误：%v", zipFileName, err)
	}
	log.Printf("已将ZIP文件 %s 保存到本地", zipFileName)
}

func parseCommandLineArgs() (args struct {
	username, repo, branch, message, token string
}) {
	flag.StringVar(&args.username, "username", "", "GitHub 用户名")
	flag.StringVar(&args.repo, "repo", "", "GitHub 仓库名")
	flag.StringVar(&args.branch, "branch", "main", "GitHub 分支名")
	flag.StringVar(&args.message, "message", "Automated commit for zipped files", "GitHub 提交信息")
	flag.StringVar(&args.token, "token", "", "GitHub 个人访问令牌")

	flag.Parse()

	if args.username == "" || args.repo == "" || args.token == "" {
		log.Fatal("请提供 GitHub 用户名、仓库名和访问令牌")
	}

	log.Printf("已设置 GitHub 用户名：%s，仓库名：%s，分支名：%s，提交信息：%s\n", args.username, args.repo, args.branch, args.message)

	return args
}

func prepareFileInfo() (zipFileName, selfName string) {
	currentDate := time.Now().Format("2006-01-02")
	zipFileName = fmt.Sprintf("uploaded_files_%s.zip", currentDate)

	selfPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	selfName = filepath.Base(selfPath)

	log.Printf("ZIP 文件名设置为：%s\n", zipFileName)

	return zipFileName, selfName
}

func createZipBuffer(zipFileName, selfName string) *bytes.Buffer {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	log.Println("开始创建 ZIP 文件并添加文件...")

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || info.Name() == selfName {
			return nil
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		safePath := filepath.ToSlash(path)
		f, err := zipWriter.Create(safePath)
		if err != nil {
			return err
		}

		_, err = f.Write(data)
		if err != nil {
			return err
		}

		log.Printf("已添加文件：%s 到 ZIP 文件\n", path)

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	err = zipWriter.Close()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("ZIP 文件创建成功")

	return &buf
}

func setupGitHubClient(token string) (ts oauth2.TokenSource, tc *http.Client, client *github.Client) {
	log.Println("正在设置 GitHub 客户端")

	ts = oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc = oauth2.NewClient(context.Background(), ts)
	client = github.NewClient(tc)

	log.Println("GitHub 客户端设置完成")

	return ts, tc, client
}

func commitZipToRepo(args struct {
	username, repo, branch, message, token string
}, zipFileName string, zipBuf *bytes.Buffer, ts oauth2.TokenSource, tc *http.Client, client *github.Client) {
	log.Printf("开始将 ZIP 文件提交到仓库：%s 分支：%s\n", args.repo, args.branch)

	opts := &github.RepositoryContentFileOptions{
		Message: github.String(args.message),
		Content: zipBuf.Bytes(),
		Branch:  github.String(args.branch),
	}

	// 检查文件是否存在
	fileContent, _, _, err := client.Repositories.GetContents(context.Background(), args.username, args.repo, zipFileName, &github.RepositoryContentGetOptions{
		Ref: args.branch,
	})

	if err != nil {
		if _, ok := err.(*github.ErrorResponse); ok && strings.Contains(err.Error(), "404 Not Found") {
			log.Println("文件不存在，创建新文件")
			// 文件不存在，创建新文件
			_, _, err = client.Repositories.CreateFile(context.Background(), args.username, args.repo, zipFileName, opts)
			if err != nil {
				log.Fatal(err)
			}
			log.Println("成功提交 ZIP 文件")
			log.Println("推送完成")
		} else {
			// 发生其他错误
			log.Fatal(err)
		}
	} else {
		// 文件存在，更新文件
		log.Print("文件已存在，更新文件")
		opts.SHA = fileContent.SHA
		_, _, err = client.Repositories.UpdateFile(context.Background(), args.username, args.repo, zipFileName, opts)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("成功更新 ZIP 文件")
		log.Println("推送完成")
	}
}
