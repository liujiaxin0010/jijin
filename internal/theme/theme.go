package theme

import (
	"image/color"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// AppTheme 自定义主题
type AppTheme struct{}

var _ fyne.Theme = (*AppTheme)(nil)

// 颜色定义
var (
	// 主色调 - 蓝色系
	primaryColor   = color.RGBA{R: 64, G: 128, B: 255, A: 255}
	primaryDark    = color.RGBA{R: 48, G: 96, B: 200, A: 255}

	// 背景色
	bgColor        = color.RGBA{R: 248, G: 250, B: 252, A: 255}
	cardBgColor    = color.RGBA{R: 255, G: 255, B: 255, A: 255}

	// 文字颜色
	textColor      = color.RGBA{R: 30, G: 41, B: 59, A: 255}
	textMutedColor = color.RGBA{R: 100, G: 116, B: 139, A: 255}

	// 状态颜色
	successColor   = color.RGBA{R: 34, G: 197, B: 94, A: 255}
	dangerColor    = color.RGBA{R: 239, G: 68, B: 68, A: 255}
	warningColor   = color.RGBA{R: 245, G: 158, B: 11, A: 255}

	// 边框和分割线
	borderColor    = color.RGBA{R: 226, G: 232, B: 240, A: 255}
)

// Color 返回主题颜色
func (t *AppTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		return primaryColor
	case theme.ColorNameBackground:
		return bgColor
	case theme.ColorNameButton:
		return primaryColor
	case theme.ColorNameForeground:
		return textColor
	case theme.ColorNameDisabled:
		return textMutedColor
	case theme.ColorNamePlaceHolder:
		return textMutedColor
	case theme.ColorNameHover:
		return color.RGBA{R: 240, G: 245, B: 255, A: 255}
	case theme.ColorNameFocus:
		return primaryColor
	case theme.ColorNameSelection:
		return color.RGBA{R: 200, G: 220, B: 255, A: 255}
	case theme.ColorNameInputBackground:
		return cardBgColor
	case theme.ColorNameSeparator:
		return borderColor
	case theme.ColorNameSuccess:
		return successColor
	case theme.ColorNameWarning:
		return warningColor
	case theme.ColorNameError:
		return dangerColor
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

// Font 返回字体
func (t *AppTheme) Font(style fyne.TextStyle) fyne.Resource {
	return loadChineseFont()
}

// Icon 返回图标
func (t *AppTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size 返回尺寸
func (t *AppTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameInnerPadding:
		return 12
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 20
	case theme.SizeNameSubHeadingText:
		return 16
	case theme.SizeNameCaptionText:
		return 12
	case theme.SizeNameInputBorder:
		return 1
	case theme.SizeNameScrollBar:
		return 12
	case theme.SizeNameScrollBarSmall:
		return 4
	default:
		return theme.DefaultTheme().Size(name)
	}
}

// 中文字体缓存
var chineseFontResource fyne.Resource

// loadChineseFont 加载中文字体
func loadChineseFont() fyne.Resource {
	if chineseFontResource != nil {
		return chineseFontResource
	}

	// Windows 字体路径 - 优先使用单独的TTF文件
	fontPaths := []string{
		"C:\\Windows\\Fonts\\simhei.ttf",   // 黑体 (TTF格式，兼容性最好)
		"C:\\Windows\\Fonts\\simsun.ttc",   // 宋体
		"C:\\Windows\\Fonts\\msyh.ttf",     // 微软雅黑 TTF
		"C:\\Windows\\Fonts\\msyh.ttc",     // 微软雅黑 TTC
	}

	for _, path := range fontPaths {
		if data, err := os.ReadFile(path); err == nil {
			chineseFontResource = fyne.NewStaticResource("chinese_font.ttf", data)
			return chineseFontResource
		}
	}

	// 回退到默认字体
	return theme.DefaultTheme().Font(fyne.TextStyle{})
}

// GetSuccessColor 获取成功颜色(绿色)
func GetSuccessColor() color.Color {
	return successColor
}

// GetDangerColor 获取危险颜色(红色)
func GetDangerColor() color.Color {
	return dangerColor
}

// GetPrimaryColor 获取主色
func GetPrimaryColor() color.Color {
	return primaryColor
}

// GetMutedColor 获取次要文字颜色
func GetMutedColor() color.Color {
	return textMutedColor
}

// GetCardBgColor 获取卡片背景色
func GetCardBgColor() color.Color {
	return cardBgColor
}
