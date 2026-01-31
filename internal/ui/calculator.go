package ui

import (
	"fmt"
	"strconv"
	"time"

	"jijin/internal/service"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// CalculatorUI 定投计算器UI
type CalculatorUI struct {
	content *fyne.Container

	// 输入
	codeEntry      *widget.Entry
	amountEntry    *widget.Entry
	frequencySelect *widget.Select
	startDateEntry *widget.Entry
	endDateEntry   *widget.Entry

	// 结果
	resultCard     *widget.Card
	investLabel    *widget.Label
	valueLabel     *widget.Label
	profitLabel    *widget.Label
	rateLabel      *widget.Label
	annualLabel    *widget.Label
	countLabel     *widget.Label
	avgCostLabel   *widget.Label
}

// NewCalculatorUI 创建定投计算器UI
func NewCalculatorUI() *CalculatorUI {
	c := &CalculatorUI{}
	c.build()
	return c
}

// build 构建UI
func (c *CalculatorUI) build() {
	// 输入表单
	c.codeEntry = widget.NewEntry()
	c.codeEntry.SetPlaceHolder("如: 000001")

	c.amountEntry = widget.NewEntry()
	c.amountEntry.SetPlaceHolder("每期定投金额")
	c.amountEntry.SetText("1000")

	c.frequencySelect = widget.NewSelect(
		[]string{"每日", "每周", "每月"},
		nil,
	)
	c.frequencySelect.SetSelected("每月")

	// 默认日期
	now := time.Now()
	oneYearAgo := now.AddDate(-1, 0, 0)

	c.startDateEntry = widget.NewEntry()
	c.startDateEntry.SetText(oneYearAgo.Format("2006-01-02"))

	c.endDateEntry = widget.NewEntry()
	c.endDateEntry.SetText(now.Format("2006-01-02"))

	calcBtn := widget.NewButton("计算收益", c.calculate)

	inputForm := widget.NewForm(
		widget.NewFormItem("基金代码", c.codeEntry),
		widget.NewFormItem("定投金额(元)", c.amountEntry),
		widget.NewFormItem("定投频率", c.frequencySelect),
		widget.NewFormItem("开始日期", c.startDateEntry),
		widget.NewFormItem("结束日期", c.endDateEntry),
	)

	inputCard := widget.NewCard("定投参数", "", container.NewVBox(
		inputForm,
		calcBtn,
	))

	// 结果显示
	c.investLabel = widget.NewLabel("-")
	c.valueLabel = widget.NewLabel("-")
	c.profitLabel = widget.NewLabel("-")
	c.rateLabel = widget.NewLabel("-")
	c.annualLabel = widget.NewLabel("-")
	c.countLabel = widget.NewLabel("-")
	c.avgCostLabel = widget.NewLabel("-")

	resultContent := container.NewGridWithColumns(2,
		widget.NewLabel("总投入:"), c.investLabel,
		widget.NewLabel("当前市值:"), c.valueLabel,
		widget.NewLabel("总收益:"), c.profitLabel,
		widget.NewLabel("收益率:"), c.rateLabel,
		widget.NewLabel("年化收益:"), c.annualLabel,
		widget.NewLabel("定投次数:"), c.countLabel,
		widget.NewLabel("平均成本:"), c.avgCostLabel,
	)

	c.resultCard = widget.NewCard("计算结果", "", resultContent)

	// 说明
	helpText := `定投计算器说明：
1. 输入基金代码（如沪深300指数：000300）
2. 设置每期定投金额
3. 选择定投频率（每日/每周/每月）
4. 设置起止日期
5. 点击计算按钮查看历史回测结果

注：计算基于历史净值数据，仅供参考`

	helpCard := widget.NewCard("使用说明", "", widget.NewLabel(helpText))

	// 布局
	c.content = container.NewVBox(
		inputCard,
		c.resultCard,
		helpCard,
	)
}

// calculate 执行计算
func (c *CalculatorUI) calculate() {
	code := c.codeEntry.Text
	if code == "" {
		dialog.ShowError(fmt.Errorf("请输入基金代码"), fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}

	amount, err := strconv.ParseFloat(c.amountEntry.Text, 64)
	if err != nil || amount <= 0 {
		dialog.ShowError(fmt.Errorf("请输入有效的定投金额"), fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}

	startDate, err := time.Parse("2006-01-02", c.startDateEntry.Text)
	if err != nil {
		dialog.ShowError(fmt.Errorf("开始日期格式错误，请使用YYYY-MM-DD"), fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}

	endDate, err := time.Parse("2006-01-02", c.endDateEntry.Text)
	if err != nil {
		dialog.ShowError(fmt.Errorf("结束日期格式错误，请使用YYYY-MM-DD"), fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}

	// 频率转换
	frequency := "monthly"
	switch c.frequencySelect.Selected {
	case "每日":
		frequency = "daily"
	case "每周":
		frequency = "weekly"
	case "每月":
		frequency = "monthly"
	}

	// 显示计算中
	c.investLabel.SetText("计算中...")
	c.valueLabel.SetText("-")
	c.profitLabel.SetText("-")
	c.rateLabel.SetText("-")
	c.annualLabel.SetText("-")
	c.countLabel.SetText("-")
	c.avgCostLabel.SetText("-")

	go func() {
		result, err := service.GetCalculatorService().CalculateInvestment(code, amount, frequency, startDate, endDate)
		if err != nil {
			c.investLabel.SetText("计算失败: " + err.Error())
			return
		}

		if result == nil {
			c.investLabel.SetText("无数据")
			return
		}

		c.investLabel.SetText(fmt.Sprintf("¥%.2f", result.TotalInvest))
		c.valueLabel.SetText(fmt.Sprintf("¥%.2f", result.CurrentValue))
		c.profitLabel.SetText(fmt.Sprintf("¥%.2f", result.TotalProfit))
		c.rateLabel.SetText(fmt.Sprintf("%.2f%%", result.ProfitRate))
		c.annualLabel.SetText(fmt.Sprintf("%.2f%%", result.AnnualReturn))
		c.countLabel.SetText(fmt.Sprintf("%d次", result.InvestCount))
		c.avgCostLabel.SetText(fmt.Sprintf("¥%.4f", result.AvgCost))
	}()
}

// Content 获取内容
func (c *CalculatorUI) Content() fyne.CanvasObject {
	return c.content
}
