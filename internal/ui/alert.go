package ui

import (
	"fmt"
	"strconv"

	"jijin/internal/model"
	"jijin/internal/service"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// AlertUI 智能提醒界面
type AlertUI struct {
	content  fyne.CanvasObject
	window   fyne.Window
	ruleList *widget.List
	rules    []model.AlertRule
}

// NewAlertUI 创建智能提醒界面
func NewAlertUI() *AlertUI {
	ui := &AlertUI{}
	ui.build()
	return ui
}

func (u *AlertUI) build() {
	title := widget.NewLabel("智能提醒")
	title.TextStyle = fyne.TextStyle{Bold: true}

	addBtn := widget.NewButton("添加提醒规则", u.showAddDialog)

	u.ruleList = widget.NewList(
		func() int { return len(u.rules) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("规则名称"),
				widget.NewButton("删除", nil),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(u.rules) {
				return
			}
			rule := u.rules[id]
			box := obj.(*fyne.Container)
			label := box.Objects[0].(*widget.Label)
			label.SetText(fmt.Sprintf("%s - %s (阈值%.1f%%)",
				rule.FundName, rule.AlertType, rule.Threshold))
		},
	)

	u.content = container.NewBorder(
		container.NewVBox(title, addBtn, widget.NewSeparator()),
		nil, nil, nil,
		u.ruleList,
	)
}

// showAddDialog 显示添加规则对话框
func (u *AlertUI) showAddDialog() {
	codeEntry := widget.NewEntry()
	codeEntry.SetPlaceHolder("基金代码")

	thresholdEntry := widget.NewEntry()
	thresholdEntry.SetPlaceHolder("涨跌幅阈值(%)")
	thresholdEntry.SetText("3")

	typeSelect := widget.NewSelect(
		[]string{"盘中涨跌", "连涨连跌"},
		nil,
	)
	typeSelect.SetSelected("盘中涨跌")

	form := dialog.NewForm("添加提醒规则", "添加", "取消",
		[]*widget.FormItem{
			widget.NewFormItem("基金代码", codeEntry),
			widget.NewFormItem("提醒类型", typeSelect),
			widget.NewFormItem("阈值", thresholdEntry),
		},
		func(ok bool) {
			if !ok || codeEntry.Text == "" {
				return
			}
			threshold, _ := strconv.ParseFloat(thresholdEntry.Text, 64)
			alertType := service.AlertTypePriceChange
			if typeSelect.Selected == "连涨连跌" {
				alertType = service.AlertTypeConsecutive
			}
			rule := &model.AlertRule{
				FundCode:  codeEntry.Text,
				FundName:  codeEntry.Text,
				AlertType: alertType,
				Threshold: threshold,
				Direction: "both",
				Enabled:   true,
			}
			service.GetAlertService().CreateAlertRule(rule)
			u.Refresh()
		}, u.window)
	form.Show()
}

// Content 返回界面内容
func (u *AlertUI) Content() fyne.CanvasObject {
	return u.content
}

// SetWindow 设置窗口引用
func (u *AlertUI) SetWindow(w fyne.Window) {
	u.window = w
}

// Refresh 刷新界面
func (u *AlertUI) Refresh() {
	rules, _ := service.GetAlertService().GetAlertRules("")
	u.rules = rules
	u.ruleList.Refresh()
}
