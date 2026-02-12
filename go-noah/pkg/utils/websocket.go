package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"go-noah/pkg/global"

	"go.uber.org/zap"
)

// PublishMSG 发布消息结构
type PublishMSG struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// PublishMessageToChannel 发布消息到 Redis 频道
func PublishMessageToChannel(channel string, data interface{}, renderType string) error {
	if global.Redis == nil {
		global.Logger.Warn("Redis is not configured, cannot publish message", zap.String("channel", channel))
		return nil
	}

	var msg PublishMSG
	msg.Type = renderType
	msg.Data = data

	jsonData, err := json.Marshal(msg)
	if err != nil {
		global.Logger.Error("Failed to marshal message", zap.Error(err))
		return err
	}

	ctx := context.Background()
	if err := global.Redis.Publish(ctx, channel, jsonData).Err(); err != nil {
		global.Logger.Error("Failed to publish message to Redis", zap.Error(err), zap.String("channel", channel))
		return err
	}

	// 调试日志：记录消息发布成功（使用 Info 级别，确保能看到）
	global.Logger.Info("Published message to Redis channel",
		zap.String("channel", channel),
		zap.String("type", renderType),
		zap.String("data", fmt.Sprintf("%.200v", data)), // 限制长度，避免日志过长
	)

	return nil
}

