package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"go-noah/pkg/global"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// GhostControl 通过 Unix socket 发送命令给 gh-ost
// socketPath: Unix socket 路径，格式：/tmp/gh-ost.{database}.{table}.sock
// command: 要发送的命令
//   - throttle: 暂停执行
//   - unthrottle: 恢复执行
//   - panic: 取消执行（紧急停止）
//   - chunk-size=xxx: 设置 chunk size（例如：chunk-size=1000）
//   - max-lag-millis=xxx: 设置最大延迟（例如：max-lag-millis=2000）
func GhostControl(socketPath, command string) error {
	// 连接 Unix socket
	conn, err := net.DialTimeout("unix", socketPath, 5*time.Second)
	if err != nil {
		global.Logger.Error("Failed to connect to gh-ost socket",
			zap.String("socket_path", socketPath),
			zap.String("command", command),
			zap.Error(err),
		)
		return fmt.Errorf("连接 gh-ost socket 失败: %w", err)
	}
	defer conn.Close()

	// 设置写入超时
	if err := conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
		global.Logger.Error("Failed to set write deadline",
			zap.String("socket_path", socketPath),
			zap.Error(err),
		)
		return fmt.Errorf("设置写入超时失败: %w", err)
	}

	// 发送命令（gh-ost 命令以换行符结尾）
	commandWithNewline := command + "\n"
	_, err = conn.Write([]byte(commandWithNewline))
	if err != nil {
		global.Logger.Error("Failed to send command to gh-ost socket",
			zap.String("socket_path", socketPath),
			zap.String("command", command),
			zap.Error(err),
		)
		return fmt.Errorf("发送命令失败: %w", err)
	}

	global.Logger.Info("Successfully sent command to gh-ost socket",
		zap.String("socket_path", socketPath),
		zap.String("command", command),
	)

	return nil
}

// GetGhostSocketPath 根据数据库名和表名生成 gh-ost socket 路径
func GetGhostSocketPath(database, table string) string {
	return fmt.Sprintf("/tmp/gh-ost.%s.%s.sock", database, table)
}

// GetGhostSocketPathFromOrderID 从 Redis 获取 gh-ost socket 路径（根据 order_id）
// 如果 Redis 中没有，尝试从工单信息推断 socket 路径（需要提供数据库名和表名）
func GetGhostSocketPathFromOrderID(orderID string, database, table string) (string, error) {
	ctx := context.Background()

	// 首先尝试从 Redis 获取
	if global.Redis != nil {
		key := fmt.Sprintf("ghost:socket:%s", orderID)
		socketPath, err := global.Redis.Get(ctx, key).Result()
		if err == nil {
			// 验证 socket 文件是否存在
			if _, err := os.Stat(socketPath); err == nil {
				return socketPath, nil
			}
			global.Logger.Warn("Redis 中的 socket 路径文件不存在，尝试推断路径",
				zap.String("order_id", orderID),
				zap.String("socket_path", socketPath),
			)
		} else if err != redis.Nil {
			// 非 nil 错误，返回错误
			return "", fmt.Errorf("获取 socket 路径失败: %w", err)
		}
		// redis.Nil 错误，继续尝试推断
	}

	// Redis 中没有，尝试从工单信息推断 socket 路径
	if database != "" && table != "" {
		socketPath := GetGhostSocketPath(database, table)
		// 检查 socket 文件是否存在
		if _, err := os.Stat(socketPath); err == nil {
			// 文件存在，保存到 Redis 并返回
			if global.Redis != nil && orderID != "" {
				_ = SetGhostSocketPathToOrderID(orderID, socketPath) // 保存失败不影响返回
			}
			return socketPath, nil
		}
		global.Logger.Warn("推断的 socket 路径文件不存在",
			zap.String("order_id", orderID),
			zap.String("database", database),
			zap.String("table", table),
			zap.String("socket_path", socketPath),
		)
	}

	return "", fmt.Errorf("未找到 gh-ost socket 路径，可能执行已完成或未开始")
}

// SetGhostSocketPathToOrderID 将 gh-ost socket 路径保存到 Redis（根据 order_id）
func SetGhostSocketPathToOrderID(orderID, socketPath string) error {
	if global.Redis == nil {
		return fmt.Errorf("Redis 未配置")
	}

	key := fmt.Sprintf("ghost:socket:%s", orderID)
	ctx := context.Background()
	// 设置过期时间为 24 小时（gh-ost 执行完成后会清理）
	err := global.Redis.Set(ctx, key, socketPath, 24*time.Hour).Err()
	if err != nil {
		global.Logger.Error("Failed to save ghost socket path to Redis",
			zap.String("order_id", orderID),
			zap.String("socket_path", socketPath),
			zap.Error(err),
		)
		return fmt.Errorf("保存 socket 路径失败: %w", err)
	}

	global.Logger.Info("Successfully saved ghost socket path to Redis",
		zap.String("order_id", orderID),
		zap.String("socket_path", socketPath),
	)

	return nil
}

// SaveGhostProgressToRedis 将 gh-ost 进度信息保存到 Redis 缓存
// orderID: 工单ID
// progressData: 进度数据（map[string]interface{}，包含 percent, current, total, eta, operation 等）
func SaveGhostProgressToRedis(orderID string, progressData map[string]interface{}) error {
	if global.Redis == nil {
		return fmt.Errorf("Redis 未配置")
	}

	key := fmt.Sprintf("ghost:progress:%s", orderID)
	ctx := context.Background()

	// 将进度数据序列化为 JSON
	jsonData, err := json.Marshal(progressData)
	if err != nil {
		return fmt.Errorf("序列化进度数据失败: %w", err)
	}

	// 设置过期时间为 24 小时（gh-ost 执行完成后会自动过期）
	err = global.Redis.Set(ctx, key, jsonData, 24*time.Hour).Err()
	if err != nil {
		global.Logger.Error("Failed to save ghost progress to Redis",
			zap.String("order_id", orderID),
			zap.Error(err),
		)
		return fmt.Errorf("保存进度到 Redis 失败: %w", err)
	}

	return nil
}

// GetGhostProgressFromRedis 从 Redis 获取 gh-ost 最新进度信息
// orderID: 工单ID
// 返回: 进度数据（map[string]interface{}）和错误
func GetGhostProgressFromRedis(orderID string) (map[string]interface{}, error) {
	if global.Redis == nil {
		return nil, fmt.Errorf("Redis 未配置")
	}

	key := fmt.Sprintf("ghost:progress:%s", orderID)
	ctx := context.Background()

	jsonData, err := global.Redis.Get(ctx, key).Result()
	if err == redis.Nil {
		// Redis 中没有数据，返回 nil（不是错误）
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("从 Redis 获取进度失败: %w", err)
	}

	// 反序列化 JSON 数据
	var progressData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &progressData); err != nil {
		return nil, fmt.Errorf("解析进度数据失败: %w", err)
	}

	return progressData, nil
}
