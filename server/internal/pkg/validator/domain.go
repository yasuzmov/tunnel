// Package validator 存放可独立单测的输入校验。
// 域名会写入 nginx 配置文件，必须在进模板之前拒绝一切特殊字符。
package validator

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

// 允许字母数字连字符及点号组成的 FQDN，至少包含一个点，防止写成 localhost 注入。
var domainRe = regexp.MustCompile(`(?i)^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?(\.[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?)+$`)

// ValidateDomain 校验公网域名。拒绝空格、路径分隔符、通配符与 IP，避免拼进 nginx 配置。
func ValidateDomain(raw string) (string, error) {
	domain := strings.ToLower(strings.TrimSpace(raw))
	if domain == "" {
		return "", fmt.Errorf("域名不能为空")
	}
	if len(domain) > 253 {
		return "", fmt.Errorf("域名过长")
	}
	if strings.Contains(domain, "..") || strings.ContainsAny(domain, " /\\;{}$`'\"\n\r\t") {
		return "", fmt.Errorf("域名包含非法字符")
	}
	if !domainRe.MatchString(domain) {
		return "", fmt.Errorf("域名格式不合法")
	}
	if net.ParseIP(domain) != nil {
		return "", fmt.Errorf("不支持使用 IP 作为域名")
	}
	return domain, nil
}
