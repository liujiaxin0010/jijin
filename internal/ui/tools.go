package ui

import (
	"fmt"

	"jijin/internal/repository"
	"jijin/internal/service"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ToolsUI 工具箱界面
type ToolsUI struct {
	content    fyne.CanvasObject
	window     fyne.Window
	resultArea *widget.Label
}

// NewToolsUI 创建工具箱界面
func NewToolsUI() *ToolsUI {
	ui := &ToolsUI{}
	ui.build()
	return ui
}

func (u *ToolsUI) build() {
	u.resultArea = widget.NewLabel("点击上方按钮查看结果")
	u.resultArea.Wrapping = fyne.TextWrapWord

	// 结果显示区域（带滚动）
	resultScroll := container.NewVScroll(u.resultArea)
	resultScroll.SetMinSize(fyne.NewSize(0, 120))
	resultCard := widget.NewCard("分析结果", "", resultScroll)

	// 回本预测卡片
	recoveryBtn := widget.NewButton("计算回本时间", u.showRecoveryDialog)
	recoveryCard := widget.NewCard("回本预测", "预测亏损基金回本时间", recoveryBtn)

	// 盈利概率卡片
	profitBtn := widget.NewButton("计算盈利概率", u.showProfitDialog)
	profitCard := widget.NewCard("盈利概率", "计算未来盈利概率", profitBtn)

	// 波段信号卡片
	signalBtn := widget.NewButton("生成信号", u.showSignalDialog)
	signalCard := widget.NewCard("波段信号", "技术指标买卖信号", signalBtn)

	// 基金扫雷卡片
	riskBtn := widget.NewButton("风险扫描", u.scanRisk)
	riskCard := widget.NewCard("基金扫雷", "识别风险基金", riskBtn)

	// 基金排行卡片
	rankBtn := widget.NewButton("查看排行", u.showRanking)
	rankCard := widget.NewCard("基金排行", "持仓收益排行", rankBtn)

	grid := container.NewGridWithColumns(2,
		recoveryCard, profitCard,
		signalCard, riskCard,
		rankCard,
	)

	u.content = container.NewBorder(
		nil, resultCard, nil, nil,
		container.NewVScroll(grid),
	)
}

// Content 返回界面内容
func (u *ToolsUI) Content() fyne.CanvasObject {
	return u.content
}

// SetWindow 设置窗口引用
func (u *ToolsUI) SetWindow(w fyne.Window) {
	u.window = w
}

// showRecoveryDialog 显示回本预测对话框
func (u *ToolsUI) showRecoveryDialog() {
	holdings, _ := repository.GetAllHoldings()
	if len(holdings) == 0 {
		u.resultArea.SetText("暂无持仓数据")
		return
	}

	var lossHoldings []string
	for _, h := range holdings {
		if h.Profit() < 0 {
			lossHoldings = append(lossHoldings,
				fmt.Sprintf("%s (亏损%.2f%%)", h.FundName, h.ProfitRate()))
		}
	}

	if len(lossHoldings) == 0 {
		u.resultArea.SetText("恭喜！所有持仓均已盈利")
		return
	}

	result := "亏损持仓回本预测:\n"
	for _, h := range holdings {
		if h.Profit() < 0 {
			pred, err := service.GetPredictionService().PredictRecoveryTime(h.ID)
			if err == nil {
				result += fmt.Sprintf("\n%s: 预计%d天回本 (置信度%.0f%%)",
					h.FundName, pred.PredictedDays, pred.ConfidenceLevel)
			}
		}
	}
	u.resultArea.SetText(result)
}

// showProfitDialog 显示盈利概率对话框
func (u *ToolsUI) showProfitDialog() {
	codeEntry := widget.NewEntry()
	codeEntry.SetPlaceHolder("输入基金代码")

	form := dialog.NewForm("盈利概率计算", "计算", "取消",
		[]*widget.FormItem{
			widget.NewFormItem("基金代码", codeEntry),
		},
		func(ok bool) {
			if !ok || codeEntry.Text == "" {
				return
			}
			periods := []int{30, 90, 180, 365}
			results, err := service.GetPredictionService().CalculateProfitProbability(codeEntry.Text, periods)
			if err != nil {
				u.resultArea.SetText("计算失败: " + err.Error())
				return
			}
			text := fmt.Sprintf("基金 %s 盈利概率:\n", codeEntry.Text)
			for _, r := range results {
				text += fmt.Sprintf("\n持有%d天: %.1f%% (预期收益%.2f%%)",
					r.Period, r.ProfitProb, r.ExpectedReturn)
			}
			u.resultArea.SetText(text)
		}, u.window)
	form.Show()
}

// showSignalDialog 显示波段信号
func (u *ToolsUI) showSignalDialog() {
	holdings, _ := repository.GetAllHoldings()
	if len(holdings) == 0 {
		u.resultArea.SetText("暂无持仓数据")
		return
	}

	result := "波段信号分析:\n"
	for _, h := range holdings {
		signal, err := service.GetSignalService().GenerateSignal(h.FundCode)
		if err == nil {
			result += fmt.Sprintf("\n%s: %s (强度%.0f%%)\n  %s",
				h.FundName, signal.SignalType, signal.SignalStrength, signal.Reason)
		}
	}
	u.resultArea.SetText(result)
}

// scanRisk 风险扫描
func (u *ToolsUI) scanRisk() {
	holdings, _ := repository.GetAllHoldings()
	if len(holdings) == 0 {
		u.resultArea.SetText("暂无持仓数据")
		return
	}

	result := "风险扫描结果:\n"
	for _, h := range holdings {
		risk, err := service.GetRiskService().AnalyzeFundRisk(h.FundCode)
		if err == nil {
			result += fmt.Sprintf("\n%s: 风险等级%d/5 (回撤%.1f%%)\n  %s",
				h.FundName, risk.RiskLevel, risk.MaxDrawdown, risk.Suggestion)
		}
	}
	u.resultArea.SetText(result)
}

// showRanking 显示排行
func (u *ToolsUI) showRanking() {
	items, _ := service.GetRankingService().GetHoldingRanking()
	if len(items) == 0 {
		u.resultArea.SetText("暂无持仓数据")
		return
	}

	result := "持仓收益排行:\n"
	for i, item := range items {
		result += fmt.Sprintf("\n%d. %s: %.2f%%", i+1, item.FundName, item.Value)
	}
	u.resultArea.SetText(result)
}

// Refresh 刷新界面
func (u *ToolsUI) Refresh() {
}
