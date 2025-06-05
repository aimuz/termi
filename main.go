package main

import (
	"fmt"
	"os"
	"strings"

	"termi.sh/termi/internal/ui"
)


func main() {
	if len(os.Args) < 2 {
		fmt.Println("请在命令后输入自然语言，例如：\n  termi 我想对 baidu.com 发起 ping")
		os.Exit(1)
	}

	query := strings.Join(os.Args[1:], " ")
	
	if err := ui.RunApp(query); err != nil {
		fmt.Printf("应用出错: %v\n", err)
		os.Exit(1)
	}
}
