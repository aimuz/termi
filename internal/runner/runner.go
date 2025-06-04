package runner

import (
	"fmt"
	"os"
	"os/exec"
)

// Run 执行 shell 命令，并将标准输入输出直接连接到当前终端，实现完整交互体验。
func Run(cmdStr string) error {
	fmt.Println("---------------------------")
	cmd := exec.Command("bash", "-c", cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		return err
	}
	// 等待命令结束，同时让用户实时看到输出 / 与之交互
	return cmd.Wait()
}
