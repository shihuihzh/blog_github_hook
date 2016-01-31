package main

import (
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/golang/glog"
)

var path string

// 执行命令
func execute(name string, arg ...string) {

	cmd := exec.Command(name, arg...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("StdoutPipe: " + err.Error())
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Println("StderrPipe: ", err.Error())
		return
	}

	if err := cmd.Start(); err != nil {
		log.Println("Start: ", err.Error())
		return
	}

	bytesErr, err := ioutil.ReadAll(stderr)
	if err != nil {
		log.Println("ReadAll stderr: ", err.Error())
		return
	}

	if len(bytesErr) != 0 {
		glog.Infof("stderr is not nil: %s", bytesErr)
		return
	}

	bytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		log.Println("ReadAll stdout: ", err.Error())
		return
	}

	if err := cmd.Wait(); err != nil {
		log.Println("Wait: ", err.Error())
		return
	}
	glog.Infof("stdout: %s", bytes)
}

// 创建字体
func fontCreate() {
	glog.Infof("Start Create Fonts ...")

	execute("node", "./script/font-creator.js")

	glog.Infof("Create Fonts Success ...")

	// 上传文件
	uploadFile()

}

// 上传文件
func uploadFile() {
	glog.Infof("Start Upload ...")

	go execute("node", "./script/upload-fonts.js")
	go execute("node", "./script/upload-css.js")
	go execute("node", "./script/upload-js.js")
	go execute("node", "./script/upload-images.js")

	glog.Infof("Upload Success ...")
}

// 同步 git 仓库
func gitSync() {
	glog.Infof("Start git pull ...")

	execute("git", "pull")

	glog.Infof("git pull Success ...")

	// 上传文件
	fontCreate()

}

func hookHandler(w http.ResponseWriter, r *http.Request) {
	go gitSync()
	io.WriteString(w, "Hello, world!")
}

func main() {

	// 初始化日志
	flag.Parse()

	// 初始化当前程序目录路径
	file, _ := exec.LookPath(os.Args[0])
	path, _ = filepath.Abs(filepath.Dir(file))

	mux := http.NewServeMux()
	mux.HandleFunc("/github_hook.json", hookHandler)

	glog.Infof("Github-hook Server Start ...")
	err := http.ListenAndServe(":8888", mux)

	if err != nil {
		glog.Fatal("ListenAndServe: ", err.Error())
	}

}
