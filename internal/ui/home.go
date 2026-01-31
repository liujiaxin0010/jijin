package ui

import (
	"fmt"
	"image/color"

	"jijin/internal/service"
	apptheme "jijin/internal/theme"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// HomeUI 首页UI
type HomeUI struct {
	content   *fyne.Container
	onRefresh func()

	// 汇总信息
	totalCostLabel   *widget.Label
	totalValueLabel  *widget.Label
	totalProfitLabel *widget.Label
	profitRateLabel  *widget.Label

	// 持仓列表
	holdingList *fyne.Container
	holdings    []holdingItem
}

type holdingItem struct {
	Name       string
	Code       string
	Value      float64
	Profit     float64
	ProfitRate float64
	Nav        float64
}

// NewHomeUI 创建首页UI
func NewHomeUI(onRefresh func()) *HomeUI {
	h := &HomeUI{
		onRefresh: onRefresh,
	}
	h.build()
	return h
}

// createSummaryCard 创建汇总卡片
func (h *HomeUI) createSummaryCard(title string, valueLabel *widget.Label, icon fyne.Resource, bgColor color.Color) fyne.CanvasObject {
	bg := canvas.NewRectangle(bgColor)
	bg.CornerRadius = 8

	titleLabel := widget.NewLabel(title)
	titleLabel.TextStyle = fyne.TextStyle{}

	valueLabel.TextStyle = fyne.TextStyle{Bold: true}
	valueLabel.Alignment = fyne.TextAlignCenter

	iconWidget := widget.NewIcon(icon)

	content := container.NewVBox(
		container.NewHBox(iconWidget, titleLabel),
		layout.NewSpacer(),
		valueLabel,
	)

	return container.NewStack(
		bg,
		container.NewPadded(container.NewPadded(content)),
	)
}

// build 构建UI
func (h *HomeUI) build() {
	// 汇总卡片
	h.totalCostLabel = widget.NewLabel("¥0.00")
	h.totalValueLabel = widget.NewLabel("¥0.00")
	h.totalProfitLabel = widget.NewLabel("¥0.00")
	h.profitRateLabel = widget.NewLabel("0.00%")

	// 使用不同颜色的卡片
	costCard := h.createSummaryCard("总投入", h.totalCostLabel,
		theme.InfoIcon(), color.RGBA{R: 240, G: 249, B: 255, A: 255})

	valueCard := h.createSummaryCard("总市值", h.totalValueLabel,
		theme.VisibilityIcon(), color.RGBA{R: 240, G: 253, B: 244, A: 255})

	profitCard := h.createSummaryCard("总收益", h.totalProfitLabel,
		theme.MoveUpIcon(), color.RGBA{R: 254, G: 243, B: 199, A: 255})

	rateCard := h.createSummaryCard("收益率", h.profitRateLabel,
		theme.DocumentIcon(), color.RGBA{R: 254, G: 226, B: 226, A: 255})

	summaryGrid := container.NewGridWithColumns(4,
		costCard, valueCard, profitCard, rateCard,
	)

	// 标题栏
	titleLabel := widget.NewLabelWithStyle("我的持仓", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	refreshBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), h.onRefresh)

	header := container.NewBorder(nil, nil, titleLabel, refreshBtn)

	// 持仓列表容器
	h.holdingList = container.NewVBox()

	// 空状态提示
	emptyLabel := widget.NewLabel("暂无持仓，请在「搜索」页面添加基金")
	emptyLabel.Alignment = fyne.TextAlignCenter
	h.holdingList.Add(container.NewCenter(emptyLabel))

	// 滚动容器
	scroll := container.NewVScroll(h.holdingList)
	scroll.SetMinSize(fyne.NewSize(0, 400))

	// 主内容卡片
	contentCard := widget.NewCard("", "", container.NewBorder(
		header,
		nil, nil, nil,
		scroll,
	))

	h.content = container.NewBorder(
		container.NewVBox(
			container.NewPadded(summaryGrid),
			widget.NewSeparator(),
		),
		nil, nil, nil,
		container.NewPadded(contentCard),
	)
}

// createHoldingCard 创建持仓卡片
func (h *HomeUI) createHoldingCard(item holdingItem) fyne.CanvasObject {
	bg := canvas.NewRectangle(apptheme.GetCardBgColor())
	bg.CornerRadius = 6

	// 基金名称和代码
	nameLabel := widget.NewLabelWithStyle(item.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	codeLabel := widget.NewLabel(item.Code)
	codeLabel.TextStyle = fyne.TextStyle{}

	// 净值
	navLabel := widget.NewLabel(fmt.Sprintf("净值: %.4f", item.Nav))

	// 市值
	valueLabel := widget.NewLabelWithStyle(
		fmt.Sprintf("¥%.2f", item.Value),
		fyne.TextAlignTrailing,
		fyne.TextStyle{Bold: true},
	)

	// 收益和收益率
	profitText := fmt.Sprintf("%.2f", item.Profit)
	rateText := fmt.Sprintf("%.2f%%", item.ProfitRate)

	if item.Profit >= 0 {
		profitText = "+" + profitText
		rateText = "+" + rateText
	}

	profitLabel := widget.NewLabel(profitText)
	rateLabel := widget.NewLabel(rateText)

	if item.Profit >= 0 {
		profitLabel.Importance = widget.SuccessImportance
		rateLabel.Importance = widget.SuccessImportance
	} else {
		profitLabel.Importance = widget.DangerImportance
		rateLabel.Importance = widget.DangerImportance
	}

	leftContent := container.NewVBox(
		nameLabel,
		container.NewHBox(codeLabel, navLabel),
	)

	rightContent := container.NewVBox(
		valueLabel,
		container.NewHBox(layout.NewSpacer(), profitLabel, rateLabel),
	)

	content := container.NewBorder(nil, nil, leftContent, rightContent)

	return container.NewStack(
		bg,
		container.NewPadded(content),
	)
}

// Content 获取内容
func (h *HomeUI) Content() fyne.CanvasObject {
	return h.content
}

// Refresh 刷新数据
func (h *HomeUI) Refresh() {
	// 获取汇总数据
	totalCost, totalValue, totalProfit, profitRate := service.GetPortfolioService().GetPortfolioSummary()

	h.totalCostLabel.SetText(fmt.Sprintf("¥%.2f", totalCost))
	h.totalValueLabel.SetText(fmt.Sprintf("¥%.2f", totalValue))

	profitText := fmt.Sprintf("¥%.2f", totalProfit)
	rateText := fmt.Sprintf("%.2f%%", profitRate)

	if totalProfit >= 0 {
		profitText = "+" + profitText
		rateText = "+" + rateText
		h.totalProfitLabel.Importance = widget.SuccessImportance
		h.profitRateLabel.Importance = widget.SuccessImportance
	} else {
		h.totalProfitLabel.Importance = widget.DangerImportance
		h.profitRateLabel.Importance = widget.DangerImportance
	}

	h.totalProfitLabel.SetText(profitText)
	h.profitRateLabel.SetText(rateText)

	// 获取持仓列表
	holdings, _ := service.GetPortfolioService().GetAllHoldings()
	h.holdings = make([]holdingItem, len(holdings))

	// 清空列表
	h.holdingList.RemoveAll()

	if len(holdings) == 0 {
		emptyLabel := widget.NewLabel("暂无持仓，请在「搜索」页面添加基金")
		emptyLabel.Alignment = fyne.TextAlignCenter
		h.holdingList.Add(container.NewPadded(container.NewCenter(emptyLabel)))
	} else {
		for i, holding := range holdings {
			h.holdings[i] = holdingItem{
				Name:       holding.FundName,
				Code:       holding.FundCode,
				Value:      holding.MarketValue(),
				Profit:     holding.Profit(),
				ProfitRate: holding.ProfitRate(),
				Nav:        holding.CurrentNav,
			}
			h.holdingList.Add(h.createHoldingCard(h.holdings[i]))
		}
	}

	h.holdingList.Refresh()
}
