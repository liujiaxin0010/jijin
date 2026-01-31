package ui

import (
	"fmt"
	"strconv"
	"time"

	"jijin/internal/model"
	"jijin/internal/service"
	apptheme "jijin/internal/theme"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// PortfolioUI 持仓管理UI
type PortfolioUI struct {
	content *fyne.Container

	holdingList *fyne.Container
	holdings    []model.Holding

	// 交易记录
	txList *fyne.Container
	txs    []model.Transaction
}

// NewPortfolioUI 创建持仓管理UI
func NewPortfolioUI() *PortfolioUI {
	p := &PortfolioUI{}
	p.build()
	return p
}

// build 构建UI
func (p *PortfolioUI) build() {
	// 持仓列表
	p.holdingList = container.NewVBox()
	holdingScroll := container.NewVScroll(p.holdingList)
	holdingScroll.SetMinSize(fyne.NewSize(0, 400))

	holdingCard := widget.NewCard("持仓列表", "管理您的基金持仓", holdingScroll)

	// 交易记录列表
	p.txList = container.NewVBox()
	txScroll := container.NewVScroll(p.txList)
	txScroll.SetMinSize(fyne.NewSize(0, 400))

	txCard := widget.NewCard("交易记录", "查看历史交易", txScroll)

	// 标签页
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("持仓", theme.ListIcon(), holdingCard),
		container.NewTabItemWithIcon("交易记录", theme.HistoryIcon(), txCard),
	)

	p.content = container.NewPadded(tabs)
}

// createHoldingItem 创建持仓项
func (p *PortfolioUI) createHoldingItem(h model.Holding) fyne.CanvasObject {
	bg := canvas.NewRectangle(apptheme.GetCardBgColor())
	bg.CornerRadius = 6

	// 基金信息
	nameLabel := widget.NewLabelWithStyle(h.FundName, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	codeLabel := widget.NewLabel(h.FundCode)
	codeLabel.Importance = widget.LowImportance

	// 份额和成本
	sharesLabel := widget.NewLabel(fmt.Sprintf("份额: %.2f", h.Shares))
	costLabel := widget.NewLabel(fmt.Sprintf("成本: ¥%.2f", h.Cost))

	// 市值和盈亏
	marketValue := h.MarketValue()
	profit := h.Profit()
	profitRate := h.ProfitRate()

	valueLabel := widget.NewLabelWithStyle(fmt.Sprintf("¥%.2f", marketValue), fyne.TextAlignTrailing, fyne.TextStyle{Bold: true})

	profitText := fmt.Sprintf("%.2f (%.2f%%)", profit, profitRate)
	if profit >= 0 {
		profitText = "+" + profitText
	}
	profitLabel := widget.NewLabel(profitText)
	if profit >= 0 {
		profitLabel.Importance = widget.SuccessImportance
	} else {
		profitLabel.Importance = widget.DangerImportance
	}

	// 操作按钮
	buyBtn := widget.NewButtonWithIcon("买入", theme.ContentAddIcon(), func() {
		p.showBuyDialog(h)
	})
	buyBtn.Importance = widget.HighImportance

	sellBtn := widget.NewButtonWithIcon("卖出", theme.ContentRemoveIcon(), func() {
		p.showSellDialog(h)
	})
	sellBtn.Importance = widget.WarningImportance

	deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		dialog.ShowConfirm("确认删除", fmt.Sprintf("确定要删除 %s 吗？", h.FundName), func(ok bool) {
			if ok {
				service.GetPortfolioService().DeleteHolding(h.ID)
				p.Refresh()
			}
		}, fyne.CurrentApp().Driver().AllWindows()[0])
	})

	leftContent := container.NewVBox(
		nameLabel,
		container.NewHBox(codeLabel, sharesLabel, costLabel),
	)

	rightContent := container.NewVBox(
		valueLabel,
		profitLabel,
	)

	info := container.NewBorder(nil, nil, leftContent, rightContent)
	buttons := container.NewHBox(layout.NewSpacer(), buyBtn, sellBtn, deleteBtn)

	content := container.NewVBox(info, buttons)

	return container.NewStack(bg, container.NewPadded(content))
}

// createTxItem 创建交易记录项
func (p *PortfolioUI) createTxItem(tx model.Transaction) fyne.CanvasObject {
	bg := canvas.NewRectangle(apptheme.GetCardBgColor())
	bg.CornerRadius = 4

	dateLabel := widget.NewLabel(tx.TradeDate.Format("2006-01-02"))
	dateLabel.Importance = widget.LowImportance

	nameLabel := widget.NewLabelWithStyle(tx.FundName, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	typeText := "买入"
	var typeImportance widget.Importance = widget.SuccessImportance
	if tx.Type == "sell" {
		typeText = "卖出"
		typeImportance = widget.DangerImportance
	}
	typeLabel := widget.NewLabel(typeText)
	typeLabel.Importance = typeImportance

	amountLabel := widget.NewLabel(fmt.Sprintf("¥%.2f", tx.Amount))
	sharesLabel := widget.NewLabel(fmt.Sprintf("%.2f份", tx.Shares))
	navLabel := widget.NewLabel(fmt.Sprintf("净值: %.4f", tx.NetValue))

	content := container.NewHBox(
		dateLabel,
		nameLabel,
		typeLabel,
		layout.NewSpacer(),
		amountLabel,
		sharesLabel,
		navLabel,
	)

	return container.NewStack(bg, container.NewPadded(content))
}

// showBuyDialog 显示买入对话框
func (p *PortfolioUI) showBuyDialog(h model.Holding) {
	amountEntry := widget.NewEntry()
	amountEntry.SetPlaceHolder("买入金额")

	navEntry := widget.NewEntry()
	navEntry.SetPlaceHolder("成交净值")
	if h.CurrentNav > 0 {
		navEntry.SetText(fmt.Sprintf("%.4f", h.CurrentNav))
	}

	feeEntry := widget.NewEntry()
	feeEntry.SetPlaceHolder("手续费(可选)")
	feeEntry.SetText("0")

	form := widget.NewForm(
		widget.NewFormItem("买入金额", amountEntry),
		widget.NewFormItem("成交净值", navEntry),
		widget.NewFormItem("手续费", feeEntry),
	)

	win := fyne.CurrentApp().Driver().AllWindows()[0]

	dialog.ShowCustomConfirm("买入 - "+h.FundName, "确认买入", "取消", form, func(ok bool) {
		if !ok {
			return
		}

		amount, err := strconv.ParseFloat(amountEntry.Text, 64)
		if err != nil || amount <= 0 {
			dialog.ShowError(fmt.Errorf("请输入有效的金额"), win)
			return
		}

		nav, err := strconv.ParseFloat(navEntry.Text, 64)
		if err != nil || nav <= 0 {
			dialog.ShowError(fmt.Errorf("请输入有效的净值"), win)
			return
		}

		fee, _ := strconv.ParseFloat(feeEntry.Text, 64)

		err = service.GetPortfolioService().Buy(h.FundCode, amount, nav, fee, time.Now())
		if err != nil {
			dialog.ShowError(err, win)
			return
		}

		dialog.ShowInformation("成功", "买入成功", win)
		p.Refresh()
	}, win)
}

// showSellDialog 显示卖出对话框
func (p *PortfolioUI) showSellDialog(h model.Holding) {
	sharesEntry := widget.NewEntry()
	sharesEntry.SetPlaceHolder("卖出份额")

	navEntry := widget.NewEntry()
	navEntry.SetPlaceHolder("成交净值")
	if h.CurrentNav > 0 {
		navEntry.SetText(fmt.Sprintf("%.4f", h.CurrentNav))
	}

	feeEntry := widget.NewEntry()
	feeEntry.SetPlaceHolder("手续费(可选)")
	feeEntry.SetText("0")

	availableLabel := widget.NewLabel(fmt.Sprintf("可用份额: %.2f", h.Shares))
	availableLabel.Importance = widget.LowImportance

	form := widget.NewForm(
		widget.NewFormItem("卖出份额", sharesEntry),
		widget.NewFormItem("成交净值", navEntry),
		widget.NewFormItem("手续费", feeEntry),
		widget.NewFormItem("", availableLabel),
	)

	win := fyne.CurrentApp().Driver().AllWindows()[0]

	dialog.ShowCustomConfirm("卖出 - "+h.FundName, "确认卖出", "取消", form, func(ok bool) {
		if !ok {
			return
		}

		shares, err := strconv.ParseFloat(sharesEntry.Text, 64)
		if err != nil || shares <= 0 {
			dialog.ShowError(fmt.Errorf("请输入有效的份额"), win)
			return
		}

		if shares > h.Shares {
			dialog.ShowError(fmt.Errorf("卖出份额不能超过持有份额"), win)
			return
		}

		nav, err := strconv.ParseFloat(navEntry.Text, 64)
		if err != nil || nav <= 0 {
			dialog.ShowError(fmt.Errorf("请输入有效的净值"), win)
			return
		}

		fee, _ := strconv.ParseFloat(feeEntry.Text, 64)

		err = service.GetPortfolioService().Sell(h.FundCode, shares, nav, fee, time.Now())
		if err != nil {
			dialog.ShowError(err, win)
			return
		}

		dialog.ShowInformation("成功", "卖出成功", win)
		p.Refresh()
	}, win)
}

// Content 获取内容
func (p *PortfolioUI) Content() fyne.CanvasObject {
	return p.content
}

// Refresh 刷新数据
func (p *PortfolioUI) Refresh() {
	// 刷新持仓
	p.holdings, _ = service.GetPortfolioService().GetAllHoldings()
	p.holdingList.RemoveAll()

	if len(p.holdings) == 0 {
		emptyLabel := widget.NewLabel("暂无持仓")
		emptyLabel.Alignment = fyne.TextAlignCenter
		p.holdingList.Add(container.NewCenter(emptyLabel))
	} else {
		for _, h := range p.holdings {
			p.holdingList.Add(p.createHoldingItem(h))
		}
	}
	p.holdingList.Refresh()

	// 刷新交易记录
	p.txs, _ = service.GetPortfolioService().GetTransactions("")
	p.txList.RemoveAll()

	if len(p.txs) == 0 {
		emptyLabel := widget.NewLabel("暂无交易记录")
		emptyLabel.Alignment = fyne.TextAlignCenter
		p.txList.Add(container.NewCenter(emptyLabel))
	} else {
		for _, tx := range p.txs {
			p.txList.Add(p.createTxItem(tx))
		}
	}
	p.txList.Refresh()
}
