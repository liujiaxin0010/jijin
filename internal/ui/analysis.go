package ui

import (
	"fmt"
	"image/color"

	"jijin/internal/service"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// AnalysisUI 收益分析UI
type AnalysisUI struct {
	content *fyne.Container

	// 汇总
	totalCostLabel   *widget.Label
	totalValueLabel  *widget.Label
	totalProfitLabel *widget.Label
	profitRateLabel  *widget.Label

	// 持仓分布
	distributionList *widget.List
	distributions    []distributionItem

	// 收益明细
	profitList *widget.List
	profits    []profitItem
}

type distributionItem struct {
	Name    string
	Value   float64
	Percent float64
	Color   color.Color
}

type profitItem struct {
	Name       string
	Cost       float64
	Value      float64
	Profit     float64
	ProfitRate float64
}

// NewAnalysisUI 创建收益分析UI
func NewAnalysisUI() *AnalysisUI {
	a := &AnalysisUI{}
	a.build()
	return a
}

// build 构建UI
func (a *AnalysisUI) build() {
	// 汇总卡片
	a.totalCostLabel = widget.NewLabel("¥0.00")
	a.totalCostLabel.TextStyle = fyne.TextStyle{Bold: true}

	a.totalValueLabel = widget.NewLabel("¥0.00")
	a.totalValueLabel.TextStyle = fyne.TextStyle{Bold: true}

	a.totalProfitLabel = widget.NewLabel("¥0.00")
	a.totalProfitLabel.TextStyle = fyne.TextStyle{Bold: true}

	a.profitRateLabel = widget.NewLabel("0.00%")
	a.profitRateLabel.TextStyle = fyne.TextStyle{Bold: true}

	summaryContent := container.NewGridWithColumns(4,
		container.NewVBox(widget.NewLabel("总投入"), a.totalCostLabel),
		container.NewVBox(widget.NewLabel("总市值"), a.totalValueLabel),
		container.NewVBox(widget.NewLabel("总收益"), a.totalProfitLabel),
		container.NewVBox(widget.NewLabel("收益率"), a.profitRateLabel),
	)

	summaryCard := widget.NewCard("资产汇总", "", summaryContent)

	// 持仓分布
	a.distributionList = widget.NewList(
		func() int {
			return len(a.distributions)
		},
		func() fyne.CanvasObject {
			colorBox := canvas.NewRectangle(color.RGBA{100, 149, 237, 255})
			colorBox.SetMinSize(fyne.NewSize(20, 20))
			return container.NewHBox(
				colorBox,
				widget.NewLabel("基金名称基金名称"),
				widget.NewLabel("¥10000.00"),
				widget.NewLabel("50.00%"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(a.distributions) {
				return
			}
			item := a.distributions[id]
			box := obj.(*fyne.Container)

			colorBox := box.Objects[0].(*canvas.Rectangle)
			colorBox.FillColor = item.Color
			colorBox.Refresh()

			box.Objects[1].(*widget.Label).SetText(item.Name)
			box.Objects[2].(*widget.Label).SetText(fmt.Sprintf("¥%.2f", item.Value))
			box.Objects[3].(*widget.Label).SetText(fmt.Sprintf("%.2f%%", item.Percent))
		},
	)

	distributionCard := widget.NewCard("持仓分布", "", a.distributionList)

	// 收益明细
	a.profitList = widget.NewList(
		func() int {
			return len(a.profits)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("基金名称基金名称"),
				widget.NewLabel("成本:¥10000"),
				widget.NewLabel("市值:¥12000"),
				widget.NewLabel("盈亏:¥2000"),
				widget.NewLabel("收益率:20%"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(a.profits) {
				return
			}
			item := a.profits[id]
			box := obj.(*fyne.Container)

			box.Objects[0].(*widget.Label).SetText(item.Name)
			box.Objects[1].(*widget.Label).SetText(fmt.Sprintf("成本:¥%.2f", item.Cost))
			box.Objects[2].(*widget.Label).SetText(fmt.Sprintf("市值:¥%.2f", item.Value))
			box.Objects[3].(*widget.Label).SetText(fmt.Sprintf("盈亏:¥%.2f", item.Profit))
			box.Objects[4].(*widget.Label).SetText(fmt.Sprintf("收益率:%.2f%%", item.ProfitRate))
		},
	)

	profitCard := widget.NewCard("收益明细", "", a.profitList)

	// 布局
	a.content = container.NewVBox(
		summaryCard,
		container.NewGridWithColumns(2,
			distributionCard,
			profitCard,
		),
	)
}

// Content 获取内容
func (a *AnalysisUI) Content() fyne.CanvasObject {
	return a.content
}

// Refresh 刷新数据
func (a *AnalysisUI) Refresh() {
	// 获取汇总数据
	totalCost, totalValue, totalProfit, profitRate := service.GetPortfolioService().GetPortfolioSummary()

	a.totalCostLabel.SetText(fmt.Sprintf("¥%.2f", totalCost))
	a.totalValueLabel.SetText(fmt.Sprintf("¥%.2f", totalValue))
	a.totalProfitLabel.SetText(fmt.Sprintf("¥%.2f", totalProfit))
	a.profitRateLabel.SetText(fmt.Sprintf("%.2f%%", profitRate))

	// 获取持仓数据
	holdings, _ := service.GetPortfolioService().GetAllHoldings()

	// 计算分布
	colors := []color.Color{
		color.RGBA{100, 149, 237, 255}, // 蓝色
		color.RGBA{60, 179, 113, 255},  // 绿色
		color.RGBA{255, 165, 0, 255},   // 橙色
		color.RGBA{238, 130, 238, 255}, // 紫色
		color.RGBA{255, 99, 71, 255},   // 红色
		color.RGBA{64, 224, 208, 255},  // 青色
	}

	a.distributions = make([]distributionItem, len(holdings))
	a.profits = make([]profitItem, len(holdings))

	for i, h := range holdings {
		value := h.MarketValue()
		percent := 0.0
		if totalValue > 0 {
			percent = value / totalValue * 100
		}

		a.distributions[i] = distributionItem{
			Name:    h.FundName,
			Value:   value,
			Percent: percent,
			Color:   colors[i%len(colors)],
		}

		a.profits[i] = profitItem{
			Name:       h.FundName,
			Cost:       h.Cost,
			Value:      value,
			Profit:     h.Profit(),
			ProfitRate: h.ProfitRate(),
		}
	}

	a.distributionList.Refresh()
	a.profitList.Refresh()
}
