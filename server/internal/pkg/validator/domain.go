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

// 单个 DNS 标签：小写字母、数字、连字符，1~63 位，不能以连字符开头或结尾。
var labelRe = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

// reservedLabels 保留前缀，避免与管理后台或常见基础服务冲突。
var reservedLabels = map[string]struct{}{
	"www": {}, "mail": {}, "smtp": {}, "imap": {}, "pop": {}, "pop3": {},
	"ftp": {}, "ns": {}, "ns1": {}, "ns2": {}, "mx": {}, "localhost": {},
}

// ValidateLabel 校验域名前缀（单个 DNS 标签），返回小写形式。
func ValidateLabel(raw string) (string, error) {
	label := strings.ToLower(strings.TrimSpace(raw))
	if label == "" {
		return "", fmt.Errorf("域名前缀不能为空")
	}
	if len(label) > 63 {
		return "", fmt.Errorf("域名前缀不能超过 63 个字符")
	}
	if strings.Contains(label, ".") {
		return "", fmt.Errorf("域名前缀只能是一级名称，不能包含点号")
	}
	if !labelRe.MatchString(label) {
		return "", fmt.Errorf("域名前缀只能包含小写字母、数字和连字符，且不能以连字符开头或结尾")
	}
	if _, ok := reservedLabels[label]; ok {
		return "", fmt.Errorf("域名前缀 %s 为系统保留名称", label)
	}
	return label, nil
}

// SubdomainLabel 若 host 恰好是 base 下的一级子域名，返回其前缀；否则返回空串。
func SubdomainLabel(host, base string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	base = strings.ToLower(strings.TrimSpace(base))
	if base == "" || !strings.HasSuffix(host, "."+base) {
		return ""
	}
	label := strings.TrimSuffix(host, "."+base)
	if label == "" || strings.Contains(label, ".") {
		return ""
	}
	return label
}
