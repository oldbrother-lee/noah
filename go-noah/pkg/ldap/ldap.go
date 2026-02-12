package ldap

import (
	"crypto/tls"
	"fmt"

	"github.com/go-ldap/ldap/v3"
	"github.com/spf13/viper"
)

// UserInfo LDAP 用户信息
type UserInfo struct {
	Username string
	Nickname string
	Email    string
	Mobile   string
}

// Auth LDAP 认证
func Auth(conf *viper.Viper, username, password string) (*UserInfo, error) {
	// 检查 LDAP 是否启用
	if !conf.GetBool("ldap.enable") {
		return nil, fmt.Errorf("LDAP not enabled")
	}

	// 获取 LDAP 配置
	host := conf.GetString("ldap.host")
	port := conf.GetInt("ldap.port")
	useSSL := conf.GetBool("ldap.use_ssl")
	baseDN := conf.GetString("ldap.base_dn")
	bindDN := conf.GetString("ldap.bind_dn")
	bindPass := conf.GetString("ldap.bind_pass")
	userFilter := conf.GetString("ldap.user_filter")
	nicknameAttr := conf.GetString("ldap.attributes.nickname")
	emailAttr := conf.GetString("ldap.attributes.email")
	mobileAttr := conf.GetString("ldap.attributes.mobile")

	// 1. 连接到 LDAP 服务器
	addr := fmt.Sprintf("%s:%d", host, port)
	var conn *ldap.Conn
	var err error
	if useSSL {
		conn, err = ldap.DialTLS("tcp", addr, &tls.Config{InsecureSkipVerify: true})
	} else {
		conn, err = ldap.Dial("tcp", addr)
	}
	if err != nil {
		return nil, fmt.Errorf("LDAP connect failed: %v", err)
	}
	defer conn.Close()

	// 2. 使用管理员账号绑定（用于搜索用户）
	if bindDN != "" && bindPass != "" {
		err = conn.Bind(bindDN, bindPass)
		if err != nil {
			return nil, fmt.Errorf("LDAP admin bind failed: %v", err)
		}
	}

	// 3. 搜索用户
	filter := fmt.Sprintf(userFilter, username)
	searchRequest := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		nil, // 获取所有属性
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil || len(sr.Entries) != 1 {
		return nil, fmt.Errorf("LDAP user not found or multiple entries: %v", err)
	}

	userEntry := sr.Entries[0]
	userDN := userEntry.DN

	// 4. 使用用户 DN 和密码进行绑定（验证密码）
	err = conn.Bind(userDN, password)
	if err != nil {
		return nil, fmt.Errorf("LDAP user bind failed (password mismatch): %v", err)
	}

	// 5. 返回用户信息
	return &UserInfo{
		Username: username,
		Nickname: userEntry.GetAttributeValue(nicknameAttr),
		Email:    userEntry.GetAttributeValue(emailAttr),
		Mobile:   userEntry.GetAttributeValue(mobileAttr),
	}, nil
}

