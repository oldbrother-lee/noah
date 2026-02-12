package insight

import (
	"context"
	"go-noah/pkg/global"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// WebSocketHandlerApp 全局 Handler 实例
var WebSocketHandlerApp = new(WebSocketHandler)

// WebSocketHandler WebSocket 处理
type WebSocketHandler struct{}

// upgrader WebSocket 升级器
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源（生产环境应该限制）
	},
}

// HandleWebSocket WebSocket 处理器
// @Summary WebSocket 连接
// @Description 建立 WebSocket 连接，订阅指定频道的 Redis 消息
// @Tags WebSocket
// @Accept json
// @Produce json
// @Param channel path string true "频道名称（通常是工单ID）"
// @Success 101 {string} string "Switching Protocols"
// @Router /ws/{channel} [get]
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	channel := c.Param("channel")
	if channel == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channel parameter is required"})
		return
	}

	// 升级 HTTP 连接为 WebSocket
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		global.Logger.Error("WebSocket upgrade failed", zap.Error(err))
		return
	}
	defer ws.Close()

	global.Logger.Info("WebSocket client connected", zap.String("channel", channel))

	// 如果没有 Redis，直接关闭连接
	if global.Redis == nil {
		global.Logger.Warn("Redis is not configured, closing WebSocket connection", zap.String("channel", channel))
		return
	}

	// 监听客户端是否断开连接
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				// 检查是否是正常的关闭错误
				// 1005 (no status) 是客户端正常关闭时的常见情况，不应该记录为错误
				errStr := err.Error()
				isNormalClose := strings.Contains(errStr, "1005") || strings.Contains(errStr, "no status")
				
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) && !isNormalClose {
					global.Logger.Error("WebSocket read error", zap.Error(err))
				} else {
					// 正常关闭，记录为调试日志
					global.Logger.Debug("WebSocket connection closed", zap.Error(err))
				}
				return
			}
		}
	}()

	// 订阅 Redis 频道
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()
	
	sub := global.Redis.Subscribe(ctx, channel)
	defer sub.Close()

	// 创建消息通道
	msgChan := make(chan string)

	// 启动 Redis 消息接收 goroutine
	go func() {
		defer close(msgChan)
		for {
			msg, err := sub.ReceiveMessage(ctx)
			if err != nil {
				// 检查是否是正常的关闭错误
				if err == redis.Nil || err == context.Canceled {
					return
				}
				// 检查是否是连接关闭导致的错误
				if ctx.Err() == context.Canceled {
					return
				}
				// 检查是否是网络连接关闭导致的错误（正常关闭）
				errStr := err.Error()
				if strings.Contains(errStr, "use of closed network connection") || 
				   strings.Contains(errStr, "connection closed") {
					global.Logger.Debug("Redis subscription closed", zap.Error(err), zap.String("channel", channel))
					return
				}
				// 其他错误才记录为错误
				global.Logger.Error("Redis subscribe error", zap.Error(err), zap.String("channel", channel))
				return
			}
			select {
			case msgChan <- msg.Payload:
			case <-ctx.Done():
				return
			}
		}
	}()

	// 主循环：处理消息和连接关闭
	for {
		select {
		case <-done:
			// 客户端断开连接，取消 context 以停止 Redis 订阅
			cancel()
			global.Logger.Info("WebSocket client disconnected", zap.String("channel", channel))
			return
		case <-ctx.Done():
			// Context 被取消，退出循环
			return
		case msg, ok := <-msgChan:
			if !ok {
				// 消息通道关闭
				return
			}
			// 将 Redis 消息写入 WebSocket，推送给客户端
			global.Logger.Debug("Sending message to WebSocket client", zap.String("channel", channel), zap.String("message", msg))
			if err := ws.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				// 检查是否是正常的关闭错误
				// 1005 (no status) 是客户端正常关闭时的常见情况，不应该记录为错误
				errStr := err.Error()
				isNormalClose := strings.Contains(errStr, "1005") || strings.Contains(errStr, "no status")
				
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) && !isNormalClose {
					global.Logger.Error("WebSocket write error", zap.Error(err), zap.String("channel", channel))
				} else {
					global.Logger.Debug("WebSocket write connection closed", zap.Error(err), zap.String("channel", channel))
				}
				cancel()
				return
			}
		}
	}
}

