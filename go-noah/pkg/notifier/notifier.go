package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-noah/pkg/global"
	"go-noah/pkg/utils"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// WechatNotifierConfig 企业微信通知配置
type WechatNotifierConfig struct {
	Webhook string
}

// DingTalkNotifierConfig 钉钉通知配置
type DingTalkNotifierConfig struct {
	Webhook  string
	Keywords string
}

// EmailNotifierConfig 邮件通知配置
type EmailNotifierConfig struct {
	Username string
	Host     string
	Port     int
	Password string
}

// Notifier 通知接口
type Notifier interface {
	SendMessage(subject string, users []string, msg string) error
}

// UserInfo 用户信息（用于@功能）
type UserInfo struct {
	Username string // 用户名（用于企业微信）
	Phone    string // 手机号（用于钉钉）
	Nickname string // 昵称（用于显示）
}

// WechatNotifier 企业微信通知实现
type WechatNotifier struct {
	Config WechatNotifierConfig
}

// SendMessage 发送企业微信消息
// users: 用户名列表，函数内部会查询用户信息用于@功能
func (w *WechatNotifier) SendMessage(subject string, users []string, msg string) error {
	// 批量查询用户信息
	userInfos := getUserInfos(users)
	var mentionedMobiles []string
	var mentionedUserIds []string

	for _, userInfo := range userInfos {
		// 优先使用 userid（mentioned_list）
		// 通常企业微信的 userid 就是用户名，如果企业微信配置了用户名作为 userid
		if userInfo.Username != "" {
			mentionedUserIds = append(mentionedUserIds, userInfo.Username)
		} else if userInfo.Phone != "" {
			// 如果没有用户名，使用手机号（mentioned_mobile_list）
			mentionedMobiles = append(mentionedMobiles, userInfo.Phone)
		}
	}

	// 使用 markdown 消息类型
	markdownContent := map[string]interface{}{
		"content": msg,
	}

	// 优先使用 userid 列表（mentioned_list）
	if len(mentionedUserIds) > 0 {
		markdownContent["mentioned_list"] = mentionedUserIds
	}
	// 如果有手机号列表，也添加（mentioned_mobile_list），作为备用
	if len(mentionedMobiles) > 0 {
		markdownContent["mentioned_mobile_list"] = mentionedMobiles
	}

	payload := map[string]interface{}{
		"msgtype":  "markdown",
		"markdown": markdownContent,
	}

	messageJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := http.Post(w.Config.Webhook, "application/json", bytes.NewBuffer(messageJSON))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		global.Logger.Info("企业微信消息发送成功",
			zap.Strings("users", users),
			zap.Strings("mentioned_mobiles", mentionedMobiles),
			zap.Strings("mentioned_userids", mentionedUserIds),
		)
	} else {
		// 读取响应体以便调试
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("发送企业微信消息失败: users=%v, mentioned_mobiles=%v, mentioned_userids=%v, status_code=%d, response=%s",
			users, mentionedMobiles, mentionedUserIds, resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// DingTalkNotifier 钉钉通知实现
type DingTalkNotifier struct {
	Config DingTalkNotifierConfig
}

// SendMessage 发送钉钉消息
// users: 用户名列表，函数内部会查询手机号用于@功能
func (d *DingTalkNotifier) SendMessage(subject string, users []string, msg string) error {
	// 批量查询用户手机号
	userInfos := getUserInfos(users)
	var phones []string
	var mentionTexts []string

	for _, userInfo := range userInfos {
		if userInfo.Phone != "" {
			phones = append(phones, userInfo.Phone)
			// 在消息末尾添加@格式：@手机号
			mentionTexts = append(mentionTexts, fmt.Sprintf("@%s", userInfo.Phone))
		}
	}

	// 如果有需要@的用户，在消息末尾添加@格式
	if len(mentionTexts) > 0 {
		msg = fmt.Sprintf("%s\n\n%s", msg, strings.Join(mentionTexts, " "))
	}

	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": msg,
		},
		"at": map[string]interface{}{
			"atMobiles": phones,
			"isAtAll":   false,
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化钉钉消息失败: %w", err)
	}

	global.Logger.Debug("发送钉钉消息",
		zap.String("webhook", d.Config.Webhook),
		zap.String("message", msg),
		zap.Strings("at_mobiles", phones),
		zap.String("payload", string(jsonPayload)),
	)

	resp, err := http.Post(d.Config.Webhook, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("发送钉钉HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取钉钉响应失败: %w", err)
	}
	responseBody := string(bodyBytes)

	global.Logger.Debug("钉钉响应",
		zap.Int("status_code", resp.StatusCode),
		zap.String("response_body", responseBody),
	)

	// 解析响应体，检查是否有错误
	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err == nil {
		if errcode, ok := result["errcode"].(float64); ok {
			if errcode == 0 {
				global.Logger.Info("钉钉消息发送成功", zap.Strings("users", users), zap.Strings("at_mobiles", phones))
				return nil
			} else {
				errmsg := ""
				if msg, ok := result["errmsg"].(string); ok {
					errmsg = msg
				}
				return fmt.Errorf("钉钉返回错误: errcode=%.0f, errmsg=%s", errcode, errmsg)
			}
		}
	}

	// 如果无法解析响应，检查HTTP状态码
	if resp.StatusCode == http.StatusOK {
		global.Logger.Info("钉钉消息发送成功", zap.Strings("users", users), zap.Strings("at_mobiles", phones))
		return nil
	} else {
		return fmt.Errorf("发送钉钉消息失败: status_code=%d, response=%s", resp.StatusCode, responseBody)
	}
}

// EmailNotifier 邮件通知实现
// 注意：需要安装 gopkg.in/gomail.v2 依赖才能使用
// 运行: go get gopkg.in/gomail.v2
type EmailNotifier struct {
	Config EmailNotifierConfig
}

// SendMessage 发送邮件
// 注意：此功能需要 gomail 库，暂时注释掉，需要时再启用
func (e *EmailNotifier) SendMessage(subject string, emails []string, msg string) error {
	// TODO: 需要安装 gopkg.in/gomail.v2 依赖
	// m := gomail.NewMessage()
	// m.SetHeader("From", e.Config.Username)
	// m.SetHeader("To", emails...)
	// m.SetHeader("Subject", subject)
	// m.SetBody("text/plain", msg)
	// d := gomail.NewDialer(e.Config.Host, e.Config.Port, e.Config.Username, e.Config.Password)
	// if err := d.DialAndSend(m); err != nil {
	// 	return err
	// }
	global.Logger.Info("邮件通知功能需要 gomail 库，暂时跳过", zap.Strings("emails", emails))
	return nil
}

// NewWechatNotifier 创建企业微信通知实例
func NewWechatNotifier(config WechatNotifierConfig) Notifier {
	return &WechatNotifier{Config: config}
}

// NewDingTalkNotifier 创建钉钉通知实例
func NewDingTalkNotifier(config DingTalkNotifierConfig) Notifier {
	return &DingTalkNotifier{Config: config}
}

// NewEmailNotifier 创建邮件通知实例
func NewEmailNotifier(config EmailNotifierConfig) Notifier {
	return &EmailNotifier{Config: config}
}

// SendMessage 发送通知消息（支持企业微信、钉钉、邮件）
func SendMessage(subject, orderID string, users []string, msg string) {
	// 获取通知配置
	conf := global.Conf
	if conf == nil {
		global.Logger.Warn("配置未初始化，跳过通知发送")
		return
	}

	// 获取通知URL
	noticeURL := conf.GetString("notify.notice_url")
	if noticeURL == "" {
		noticeURL = "http://localhost:8000"
	}
	orderURL := fmt.Sprintf("%s/das/orders-detail/%s", noticeURL, orderID)
	msg = fmt.Sprintf("%s\n\n工单地址：%s", msg, orderURL)

	// 去重用户列表
	newUsers := utils.RemoveDuplicate(users)
	if len(newUsers) == 0 {
		return
	}

	// 发送企业微信消息
	if conf.GetBool("notify.wechat.enable") {
		webhook := conf.GetString("notify.wechat.webhook")
		if webhook != "" {
			wechatConfig := WechatNotifierConfig{Webhook: webhook}
			wechatNotifier := NewWechatNotifier(wechatConfig)
			if err := wechatNotifier.SendMessage("", newUsers, msg); err != nil {
				global.Logger.Error("发送企业微信消息失败", zap.Error(err))
			}
		}
	}

	// 发送钉钉消息
	if conf.GetBool("notify.dingtalk.enable") {
		webhook := conf.GetString("notify.dingtalk.webhook")
		keywords := conf.GetString("notify.dingtalk.keywords")

		global.Logger.Debug("钉钉通知配置检查",
			zap.Bool("enable", conf.GetBool("notify.dingtalk.enable")),
			zap.String("webhook", webhook),
			zap.String("keywords", keywords),
		)

		if webhook != "" {
			dingTalkConfig := DingTalkNotifierConfig{
				Webhook:  webhook,
				Keywords: keywords,
			}
			dingTalkNotifier := NewDingTalkNotifier(dingTalkConfig)

			// 如果设置了关键词，确保消息中包含关键词
			withKeywordsMsg := msg
			if keywords != "" {
				// 检查消息中是否已包含关键词
				if !strings.Contains(msg, keywords) {
					withKeywordsMsg = fmt.Sprintf("%s\n\n关键词：%s", msg, keywords)
				}
			}

			if err := dingTalkNotifier.SendMessage("", newUsers, withKeywordsMsg); err != nil {
				global.Logger.Error("发送钉钉消息失败",
					zap.Error(err),
					zap.String("webhook", webhook),
					zap.String("keywords", keywords),
				)
			} else {
				global.Logger.Info("钉钉消息发送成功", zap.Strings("users", newUsers))
			}
		} else {
			global.Logger.Warn("钉钉通知已启用但webhook为空")
		}
	}

	// 发送邮件（需要 gomail 库支持）
	if conf.GetBool("notify.mail.enable") {
		mailHost := conf.GetString("notify.mail.host")
		mailPort := conf.GetInt("notify.mail.port")
		mailUsername := conf.GetString("notify.mail.username")
		mailPassword := conf.GetString("notify.mail.password")
		if mailHost != "" && mailPort > 0 && mailUsername != "" {
			emailConfig := EmailNotifierConfig{
				Host:     mailHost,
				Port:     mailPort,
				Username: mailUsername,
				Password: mailPassword,
			}
			emailNotifier := NewEmailNotifier(emailConfig)
			// 获取用户邮箱（需要从数据库查询）
			// 暂时跳过邮件发送，需要时再实现
			_ = emailNotifier
			global.Logger.Debug("邮件通知功能需要用户邮箱信息和 gomail 库，暂时跳过")
		}
	}
}

// SendOrderNotification 发送工单通知（封装常用场景）
// applicant: 申请人用户名
func SendOrderNotification(orderID, title, applicant string, receivers []string, msg string) {
	// 获取申请人昵称
	applicantNickname := getApplicantNickname(applicant)

	// 如果获取到昵称，替换消息中的申请人用户名为其昵称
	// 使用更精确的替换方式，避免误替换其他用户名
	if applicantNickname != "" && applicant != "" {
		// 替换 "用户{申请人用户名}" 为 "用户{申请人昵称}"
		msg = strings.ReplaceAll(msg, "用户"+applicant, "用户"+applicantNickname)
		// 替换 "{申请人用户名}提交了工单" 为 "{申请人昵称}提交了工单"
		msg = strings.ReplaceAll(msg, applicant+"提交了工单", applicantNickname+"提交了工单")
		// 替换 "{申请人用户名}创建了工单" 为 "{申请人昵称}创建了工单"
		msg = strings.ReplaceAll(msg, applicant+"创建了工单", applicantNickname+"创建了工单")
	}

	// 确保申请人也在接收列表中（使用用户名，因为通知系统需要用户名）
	allReceivers := append([]string{applicant}, receivers...)
	SendMessage(title, orderID, allReceivers, msg)
}

// getApplicantNickname 根据用户名获取昵称
func getApplicantNickname(username string) string {
	if username == "" {
		return ""
	}

	// 从全局数据库查询昵称
	if global.DB != nil {
		var nickname string
		err := global.DB.Table("admin_users").
			Select("nickname").
			Where("username = ?", username).
			Scan(&nickname).Error
		if err == nil && nickname != "" {
			return nickname
		}
	}

	return ""
}

// getUserInfos 批量查询用户信息（用户名、手机号、昵称）
func getUserInfos(usernames []string) []UserInfo {
	if len(usernames) == 0 {
		return []UserInfo{}
	}

	var userInfos []UserInfo
	if global.DB != nil {
		var users []struct {
			Username string
			Phone    string
			Nickname string
		}

		// 批量查询用户信息
		global.DB.Table("admin_users").
			Select("username, phone, nickname").
			Where("username IN ?", usernames).
			Find(&users)

		// 转换为 UserInfo 列表
		for _, user := range users {
			userInfos = append(userInfos, UserInfo{
				Username: user.Username,
				Phone:    user.Phone,
				Nickname: user.Nickname,
			})
		}
	}

	return userInfos
}
