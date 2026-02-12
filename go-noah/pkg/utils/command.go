package utils

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"syscall"
)

// Command 执行系统命令并实时推送输出
// ctx: 上下文，用于取消命令执行
// ch: 输出通道，用于推送命令的 stdout 和 stderr（调用者负责关闭）
// cmd: 要执行的命令字符串（会通过 bash -c 执行）
func Command(ctx context.Context, ch chan<- string, cmd string) error {
	c := exec.CommandContext(ctx, "bash", "-c", cmd)
	
	// 标准输出
	stdout, err := c.StdoutPipe()
	if err != nil {
		return err
	}
	
	// 标准错误
	stderr, err := c.StderrPipe()
	if err != nil {
		return err
	}
	
	var wg sync.WaitGroup
	// 需要读取 stdout 和 stderr
	wg.Add(2)
	go read(ctx, &wg, stderr, ch)
	go read(ctx, &wg, stdout, ch)
	
	// 启动命令
	if err := c.Start(); err != nil {
		return err
	}
	
	// 等待输出结束
	wg.Wait()
	
	// 获取退出状态
	var exitStatus int
	if err := c.Wait(); err != nil {
		if ex, ok := err.(*exec.ExitError); ok {
			if status, ok := ex.Sys().(syscall.WaitStatus); ok {
				exitStatus = status.ExitStatus()
			}
		}
		// 如果 exitStatus 为 0，说明命令正常退出（可能被取消）
		if exitStatus == 0 {
			// 检查是否是 context 取消
			if ctx.Err() != nil {
				return ctx.Err()
			}
		}
	}
	
	if exitStatus != 0 {
		return fmt.Errorf("cmd exit status %d", exitStatus)
	}
	
	return nil
}

// read 从 io.Reader 读取数据并发送到通道
func read(ctx context.Context, wg *sync.WaitGroup, std io.ReadCloser, ch chan<- string) {
	reader := bufio.NewReader(std)
	defer wg.Done()
	
	for {
		select {
		case <-ctx.Done():
			return
		default:
			readString, err := reader.ReadString('\n')
			if err != nil || err == io.EOF {
				return
			}
			ch <- readString
		}
	}
}

