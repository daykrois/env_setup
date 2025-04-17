package main

import (
	"env_setup/utils"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// const url string = "https://download.java.net/java/GA/jdk24/1f9ff9062db4449d8ca828c504ffae90/36/GPL/openjdk-24_windows-x64_bin.zip"

var homeDir string
var jdkPath string
var setupDir string

var url string
var filterKeyword string

func init() {
	homeDir, _ = os.UserHomeDir()
	jdkPath = filepath.Join(homeDir, ".env/jdk.zip")
	setupDir = filepath.Join(homeDir, ".env")

	url = "https://jdk.java.net/archive/"
	filterKeyword = "zip"
}

func main() {
	// utils.Download(url, jdkPath)
	// utils.UnzipWithProgress(jdkPath, setupDir)
	// utils.AddToPath("Environment", "abc")

	links, err := utils.GetJDKLinks(url, filterKeyword)
	if err != nil {
		log.Fatalf("错误: %v", err)
	}

	fmt.Println("已获取的 JDK 下载链接如下：")
	for jdkName, link := range links {
		fmt.Printf("JDK 名称: %s\n下载链接: %s\n\n", jdkName, link)
	}
}
