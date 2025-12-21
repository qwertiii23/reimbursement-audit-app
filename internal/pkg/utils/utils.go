package utils

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

// GenerateUUID 生成UUID
func GenerateUUID() string {
	// TODO: 实现UUID生成逻辑
	return ""
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// GenerateRandomNumber 生成随机数字
func GenerateRandomNumber(length int) string {
	const charset = "0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// MD5 计算MD5哈希
func MD5(data string) string {
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// SHA256 计算SHA256哈希
func SHA256(data string) string {
	// TODO: 实现SHA256哈希计算
	return ""
}

// Base64Encode Base64编码
func Base64Encode(data string) string {
	// TODO: 实现Base64编码
	return ""
}

// Base64Decode Base64解码
func Base64Decode(data string) (string, error) {
	// TODO: 实现Base64解码
	return "", nil
}

// IsEmpty 检查字符串是否为空
func IsEmpty(str string) bool {
	return strings.TrimSpace(str) == ""
}

// IsNotEmpty 检查字符串是否不为空
func IsNotEmpty(str string) bool {
	return !IsEmpty(str)
}

// Trim 去除字符串两端的空白字符
func Trim(str string) string {
	return strings.TrimSpace(str)
}

// Contains 检查字符串是否包含子字符串
func Contains(str, substr string) bool {
	return strings.Contains(str, substr)
}

// ContainsIgnoreCase 忽略大小写检查字符串是否包含子字符串
func ContainsIgnoreCase(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}

// Split 分割字符串
func Split(str, sep string) []string {
	return strings.Split(str, sep)
}

// Join 连接字符串数组
func Join(elems []string, sep string) string {
	return strings.Join(elems, sep)
}

// Replace 替换字符串
func Replace(str, old, new string) string {
	return strings.ReplaceAll(str, old, new)
}

// ToUpperCase 转换为大写
func ToUpperCase(str string) string {
	return strings.ToUpper(str)
}

// ToLowerCase 转换为小写
func ToLowerCase(str string) string {
	return strings.ToLower(str)
}

// IsEmail 检查是否为有效的邮箱地址
func IsEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

// IsPhone 检查是否为有效的手机号
func IsPhone(phone string) bool {
	pattern := `^1[3-9]\d{9}$`
	matched, _ := regexp.MatchString(pattern, phone)
	return matched
}

// IsIDCard 检查是否为有效的身份证号
func IsIDCard(idCard string) bool {
	// 简单的身份证号验证，实际应用中需要更复杂的验证
	pattern := `^[1-9]\d{5}(18|19|20)\d{2}((0[1-9])|(1[0-2]))(([0-2][1-9])|10|20|30|31)\d{3}[0-9Xx]$`
	matched, _ := regexp.MatchString(pattern, idCard)
	return matched
}

// IsURL 检查是否为有效的URL
func IsURL(url string) bool {
	pattern := `^(https?|ftp):\/\/[^\s/$.?#].[^\s]*$`
	matched, _ := regexp.MatchString(pattern, url)
	return matched
}

// IsIP 检查是否为有效的IP地址
func IsIP(ip string) bool {
	pattern := `^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`
	matched, _ := regexp.MatchString(pattern, ip)
	return matched
}

// FormatDate 格式化日期
func FormatDate(t time.Time, layout string) string {
	return t.Format(layout)
}

// ParseDate 解析日期字符串
func ParseDate(dateStr, layout string) (time.Time, error) {
	return time.Parse(layout, dateStr)
}

// Now 获取当前时间
func Now() time.Time {
	return time.Now()
}

// Unix 获取当前Unix时间戳
func Unix() int64 {
	return time.Now().Unix()
}

// UnixMilli 获取当前毫秒级Unix时间戳
func UnixMilli() int64 {
	return time.Now().UnixMilli()
}

// Sleep 休眠
func Sleep(d time.Duration) {
	time.Sleep(d)
}

// After 在指定时间后执行
func After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

// NewTimer 创建定时器
func NewTimer(d time.Duration) *time.Timer {
	return time.NewTimer(d)
}

// NewTicker 创建周期性定时器
func NewTicker(d time.Duration) *time.Ticker {
	return time.NewTicker(d)
}