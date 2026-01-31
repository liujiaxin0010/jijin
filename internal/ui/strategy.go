package ui

import (
	"encoding/json"
	"fmt"
	"strconv"

	"jijin/internal/model"
	"jijin/internal/service"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// StrategyUI 策略配置UI
type StrategyUI struct {
	content *fyne.Container

	strategyList *widget.List
	strategies   []model.Strategy

	// 策略计算结果
	resultLabel *widget.Label
}

// NewStrategyUI 创建策略配置UI
func NewStrategyUI() *StrategyUI {
	s := &StrategyUI{}
	s.build()
	return s
}

// build 构建UI
func (s *StrategyUI) build() {
	// 添加策略按钮
	addBtn := widget.NewButton("新建策略", s.showAddDialog)

	// 策略列表
	s.strategyList = widget.NewList(
		func() int {
			return len(s.strategies)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("策略名称策略名称"),
				widget.NewLabel("基金代码"),
				widget.NewLabel("类型"),
				widget.NewLabel("基准"),
				widget.NewLabel("状态"),
				widget.NewButton("计算", nil),
				widget.NewButton("删除", nil),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(s.strategies) {
				return
			}
			st := s.strategies[id]
			box := obj.(*fyne.Container)

			box.Objects[0].(*widget.Label).SetText(st.Name)
			box.Objects[1].(*widget.Label).SetText(st.FundCode)

			typeText := "普通定投"
			switch st.StrategyType {
			case "ma_deviation":
				typeText = "均线偏离"
			case "valuation":
				typeText = "估值定投"
			case "target_value":
				typeText = "目标市值"
			}
			box.Objects[2].(*widget.Label).SetText(typeText)
			box.Objects[3].(*widget.Label).SetText(fmt.Sprintf("¥%.0f", st.BaseAmount))

			status := "停用"
			if st.Active {
				status = "启用"
			}
			box.Objects[4].(*widget.Label).SetText(status)

			// 计算按钮
			calcBtn := box.Objects[5].(*widget.Button)
			calcBtn.OnTapped = func() {
				s.calculateStrategy(st)
			}

			// 删除按钮
			delBtn := box.Objects[6].(*widget.Button)
			delBtn.OnTapped = func() {
				dialog.ShowConfirm("确认删除", "确定要删除策略 "+st.Name+" 吗？", func(ok bool) {
					if ok {
						service.GetStrategyService().DeleteStrategy(st.ID)
						s.Refresh()
					}
				}, fyne.CurrentApp().Driver().AllWindows()[0])
			}
		},
	)

	// 结果显示
	s.resultLabel = widget.NewLabel("选择策略并点击'计算'查看建议金额")
	resultCard := widget.NewCard("策略计算结果", "", s.resultLabel)

	// 策略说明
	helpText := `智能定投策略说明：

1. 均线偏离法
   根据当前净值与均线的偏离程度调整定投金额
   低于均线时加倍，高于均线时减少

2. 估值定投
   根据指数PE/PB估值调整定投金额
   低估时多投，高估时少投

3. 目标市值法
   设定目标持仓市值，每期补足差额
   适合有明确目标的投资者`

	helpCard := widget.NewCard("策略说明", "", widget.NewLabel(helpText))

	// 布局
	header := container.NewBorder(nil, nil, widget.NewLabel("我的策略"), addBtn)

	s.content = container.NewBorder(
		header,
		container.NewVBox(resultCard, helpCard),
		nil,
		nil,
		s.strategyList,
	)
}

// showAddDialog 显示添加策略对话框
func (s *StrategyUI) showAddDialog() {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("策略名称")

	codeEntry := widget.NewEntry()
	codeEntry.SetPlaceHolder("基金代码")

	amountEntry := widget.NewEntry()
	amountEntry.SetPlaceHolder("基准金额")
	amountEntry.SetText("1000")

	frequencySelect := widget.NewSelect([]string{"每日", "每周", "每月"}, nil)
	frequencySelect.SetSelected("每月")

	typeSelect := widget.NewSelect([]string{"普通定投", "均线偏离", "估值定投", "目标市值"}, nil)
	typeSelect.SetSelected("均线偏离")

	// 均线周期
	maPeriodEntry := widget.NewEntry()
	maPeriodEntry.SetPlaceHolder("均线周期(天)")
	maPeriodEntry.SetText("250")

	form := widget.NewForm(
		widget.NewFormItem("策略名称", nameEntry),
		widget.NewFormItem("基金代码", codeEntry),
		widget.NewFormItem("基准金额", amountEntry),
		widget.NewFormItem("定投频率", frequencySelect),
		widget.NewFormItem("策略类型", typeSelect),
		widget.NewFormItem("均线周期", maPeriodEntry),
	)

	win := fyne.CurrentApp().Driver().AllWindows()[0]

	dialog.ShowCustomConfirm("新建策略", "创建", "取消", form, func(ok bool) {
		if !ok {
			return
		}

		name := nameEntry.Text
		code := codeEntry.Text
		if name == "" || code == "" {
			dialog.ShowError(fmt.Errorf("请填写策略名称和基金代码"), win)
			return
		}

		amount, err := strconv.ParseFloat(amountEntry.Text, 64)
		if err != nil || amount <= 0 {
			dialog.ShowError(fmt.Errorf("请输入有效的基准金额"), win)
			return
		}

		// 频率
		frequency := "monthly"
		switch frequencySelect.Selected {
		case "每日":
			frequency = "daily"
		case "每周":
			frequency = "weekly"
		}

		// 类型
		strategyType := "normal"
		switch typeSelect.Selected {
		case "均线偏离":
			strategyType = "ma_deviation"
		case "估值定投":
			strategyType = "valuation"
		case "目标市值":
			strategyType = "target_value"
		}

		// 参数
		maPeriod, _ := strconv.Atoi(maPeriodEntry.Text)
		if maPeriod <= 0 {
			maPeriod = 250
		}

		params := service.MADeviationParams{
			MAPeriod:      maPeriod,
			MaxMultiplier: 2.0,
			MinMultiplier: 0.5,
		}

		// 获取基金名称
		fundName := code
		results, _ := service.GetFundAPI().SearchFund(code)
		for _, r := range results {
			if r.Code == code {
				fundName = r.Name
				break
			}
		}

		_, err = service.GetStrategyService().CreateStrategy(name, code, fundName, frequency, strategyType, amount, params)
		if err != nil {
			dialog.ShowError(err, win)
			return
		}

		dialog.ShowInformation("成功", "策略创建成功", win)
		s.Refresh()
	}, win)
}

// calculateStrategy 计算策略建议
func (s *StrategyUI) calculateStrategy(st model.Strategy) {
	s.resultLabel.SetText("计算中...")

	go func() {
		var result *service.StrategyResult
		var err error

		switch st.StrategyType {
		case "ma_deviation":
			var params service.MADeviationParams
			json.Unmarshal([]byte(st.Params), &params)
			if params.MAPeriod == 0 {
				params.MAPeriod = 250
				params.MaxMultiplier = 2.0
				params.MinMultiplier = 0.5
			}
			result, err = service.GetStrategyService().CalculateMADeviation(st.FundCode, st.BaseAmount, params)

		case "valuation":
			var params service.ValuationParams
			json.Unmarshal([]byte(st.Params), &params)
			if params.LowPE == 0 {
				params.LowPE = 10
				params.HighPE = 20
				params.MaxMultiplier = 2.0
				params.MinMultiplier = 0.5
			}
			// 这里需要获取当前PE，暂时用默认值
			result, err = service.GetStrategyService().CalculateValuation(st.FundCode, st.BaseAmount, 15, params)

		default:
			result = &service.StrategyResult{
				BaseAmount:    st.BaseAmount,
				SuggestAmount: st.BaseAmount,
				Multiplier:    1.0,
				Reason:        "普通定投，按基准金额投入",
			}
		}

		if err != nil {
			s.resultLabel.SetText("计算失败: " + err.Error())
			return
		}

		text := fmt.Sprintf(`策略: %s
基金: %s
基准金额: ¥%.2f
建议金额: ¥%.2f
倍数: %.2fx
原因: %s`,
			st.Name,
			st.FundName,
			result.BaseAmount,
			result.SuggestAmount,
			result.Multiplier,
			result.Reason,
		)

		s.resultLabel.SetText(text)
	}()
}

// Content 获取内容
func (s *StrategyUI) Content() fyne.CanvasObject {
	return s.content
}

// Refresh 刷新数据
func (s *StrategyUI) Refresh() {
	s.strategies, _ = service.GetStrategyService().GetAllStrategies()
	s.strategyList.Refresh()
}
