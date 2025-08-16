package utils
import (
	"fmt"
	"time"
)
type TimeHelper struct{}
func (th *TimeHelper) FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0f秒", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.0f分钟", d.Minutes())
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%.1f小时", d.Hours())
	} else {
		days := int(d.Hours() / 24)
		hours := d.Hours() - float64(days*24)
		return fmt.Sprintf("%d天%.0f小时", days, hours)
	}
}
func (th *TimeHelper) FormatTimestamp(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)
	if diff < time.Minute {
		return "刚刚"
	} else if diff < time.Hour {
		return fmt.Sprintf("%.0f分钟前", diff.Minutes())
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%.0f小时前", diff.Hours())
	} else if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d天前", days)
	} else {
		return t.Format("01-02 15:04")
	}
}
func (th *TimeHelper) IsToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.YearDay() == now.YearDay()
}
type StringHelper struct{}
func (sh *StringHelper) TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	runes := []rune(s)
	if len(runes) <= maxLength {
		return s
	}
	if maxLength > 3 {
		return string(runes[:maxLength-3]) + "..."
	}
	return string(runes[:maxLength])
}
func (sh *StringHelper) IsEmpty(s string) bool {
	return len(s) == 0
}
func (sh *StringHelper) SanitizeFilename(filename string) string {
	forbidden := []rune{'<', '>', ':', '"', '/', '\\', '|', '?', '*'}
	runes := []rune(filename)
	result := make([]rune, 0, len(runes))
	for _, r := range runes {
		isForbidden := false
		for _, f := range forbidden {
			if r == f {
				isForbidden = true
				break
			}
		}
		if !isForbidden && r >= 32 { // 过滤控制字符
			result = append(result, r)
		} else {
			result = append(result, '_')
		}
	}
	return string(result)
}
// ColorHelper 颜色辅助工具
type ColorHelper struct{}
// GetContrastColor 获取对比色（黑或白）
func (ch *ColorHelper) GetContrastColor(r, g, b uint8) (uint8, uint8, uint8) {
	// 计算亮度
	brightness := (int(r)*299 + int(g)*587 + int(b)*114) / 1000
	if brightness > 128 {
		return 0, 0, 0 // 返回黑色
	}
	return 255, 255, 255 // 返回白色
}
// ValidatorHelper 验证辅助工具
type ValidatorHelper struct{}
// IsValidPort 验证端口号是否有效
func (vh *ValidatorHelper) IsValidPort(port int) bool {
	return port > 0 && port <= 65535
}
// IsValidIPAddress 验证IP地址是否有效（简单验证）
func (vh *ValidatorHelper) IsValidIPAddress(ip string) bool {
	// 简单的IP地址格式验证
	if len(ip) == 0 {
		return false
	}
	// TODO: 实现更完整的IP地址验证
	return true
}
// SafeExecute 安全执行函数，捕获panic
func SafeExecute(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered from panic: %v\n", r)
		}
	}()
	fn()
}
// SafeExecuteWithReturn 安全执行函数并返回错误
func SafeExecuteWithReturn(fn func() error) error {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic recovered: %v", r)
			fmt.Printf("Recovered from panic: %v\n", r)
		}
	}()
	err = fn()
	return err
}