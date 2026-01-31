package ui

import (
	"fmt"
	"image/color"
	"strconv"

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

// SearchUI 搜索页面UI
type SearchUI struct {
	content *fyne.Container
	onAdd   func(code, name string, amount, nav, fee float64)

	searchEntry *widget.Entry
	resultList  *fyne.Container
	results     []model.FundSearchResult

	// 基金详情
	detailCard  fyne.CanvasObject
	codeLabel   *widget.Label
	nameLabel   *widget.Label
	typeLabel   *widget.Label
	navLabel    *widget.Label
	estLabel    *widget.Label
	growthLabel *widget.Label
	addBtn      *widget.Button

	// 走势图
	chartContainer *fyne.Container
	currentCode    string
}

// NewSearchUI 创建搜索页面UI
func NewSearchUI(onAdd func(code, name string, amount, nav, fee float64)) *SearchUI {
	s := &SearchUI{
		onAdd: onAdd,
	}
	s.build()
	return s
}

// build 构建UI
func (s *SearchUI) build() {
	// 搜索框
	s.searchEntry = widget.NewEntry()
	s.searchEntry.SetPlaceHolder("输入基金代码或名称搜索...")
	s.searchEntry.OnChanged = func(text string) {
		if len(text) >= 2 {
			s.doSearch(text)
		}
	}

	searchBtn := widget.NewButtonWithIcon("搜索", theme.SearchIcon(), func() {
		s.doSearch(s.searchEntry.Text)
	})
	searchBtn.Importance = widget.HighImportance

	searchBar := container.NewBorder(nil, nil, nil, searchBtn, s.searchEntry)

	// 搜索结果列表
	s.resultList = container.NewVBox()

	// 初始提示
	tipLabel := widget.NewLabel("请输入基金代码或名称进行搜索")
	tipLabel.Alignment = fyne.TextAlignCenter
	s.resultList.Add(container.NewCenter(tipLabel))

	resultScroll := container.NewVScroll(s.resultList)
	resultScroll.SetMinSize(fyne.NewSize(350, 300))

	// 基金详情卡片
	s.codeLabel = widget.NewLabel("-")
	s.nameLabel = widget.NewLabel("-")
	s.typeLabel = widget.NewLabel("-")
	s.navLabel = widget.NewLabel("-")
	s.estLabel = widget.NewLabel("-")
	s.growthLabel = widget.NewLabel("-")

	s.addBtn = widget.NewButtonWithIcon("添加到持仓", theme.ContentAddIcon(), func() {
		if s.codeLabel.Text != "-" && s.codeLabel.Text != "" {
			s.showAddHoldingDialog()
		}
	})
	s.addBtn.Importance = widget.HighImportance

	detailContent := container.NewVBox(
		s.createDetailRow("基金代码", s.codeLabel),
		widget.NewSeparator(),
		s.createDetailRow("基金名称", s.nameLabel),
		widget.NewSeparator(),
		s.createDetailRow("基金类型", s.typeLabel),
		widget.NewSeparator(),
		s.createDetailRow("单位净值", s.navLabel),
		widget.NewSeparator(),
		s.createDetailRow("估算净值", s.estLabel),
		widget.NewSeparator(),
		s.createDetailRow("估算涨跌", s.growthLabel),
		layout.NewSpacer(),
		s.addBtn,
	)

	detailCardWidget := widget.NewCard("基金详情", "", container.NewPadded(detailContent))

	// 走势图容器
	s.chartContainer = container.NewVBox()
	chartCard := widget.NewCard("近期走势", "最近30天净值走势", s.chartContainer)

	// 右侧面板
	rightPanel := container.NewVBox(
		detailCardWidget,
		chartCard,
	)

	// 左侧面板
	leftPanel := widget.NewCard("搜索基金", "", container.NewBorder(
		container.NewVBox(searchBar, widget.NewSeparator()),
		nil, nil, nil,
		resultScroll,
	))

	// 布局
	split := container.NewHSplit(leftPanel, container.NewVScroll(rightPanel))
	split.SetOffset(0.4)

	s.content = container.NewPadded(split)
}

// showAddHoldingDialog 显示添加持仓对话框
func (s *SearchUI) showAddHoldingDialog() {
	win := fyne.CurrentApp().Driver().AllWindows()[0]

	amountEntry := widget.NewEntry()
	amountEntry.SetPlaceHolder("输入买入金额")

	// 获取当前净值
	currentNav := s.navLabel.Text
	if s.estLabel.Text != "-" && s.estLabel.Text != "" && s.estLabel.Text != "0.0000" {
		currentNav = s.estLabel.Text
	}

	navEntry := widget.NewEntry()
	navEntry.SetPlaceHolder("成交净值")
	navEntry.SetText(currentNav)

	feeEntry := widget.NewEntry()
	feeEntry.SetPlaceHolder("手续费(可选)")
	feeEntry.SetText("0")

	form := widget.NewForm(
		widget.NewFormItem("买入金额(元)", amountEntry),
		widget.NewFormItem("成交净值", navEntry),
		widget.NewFormItem("手续费(元)", feeEntry),
	)

	dialog.ShowCustomConfirm(
		"添加持仓 - "+s.nameLabel.Text,
		"确认添加",
		"取消",
		form,
		func(ok bool) {
			if !ok {
				return
			}

			amount, err := strconv.ParseFloat(amountEntry.Text, 64)
			if err != nil || amount <= 0 {
				dialog.ShowError(fmt.Errorf("请输入有效的买入金额"), win)
				return
			}

			nav, err := strconv.ParseFloat(navEntry.Text, 64)
			if err != nil || nav <= 0 {
				dialog.ShowError(fmt.Errorf("请输入有效的净值"), win)
				return
			}

			fee, _ := strconv.ParseFloat(feeEntry.Text, 64)

			// 调用回调添加持仓
			s.onAdd(s.codeLabel.Text, s.nameLabel.Text, amount, nav, fee)

			dialog.ShowInformation("成功", fmt.Sprintf("已添加持仓\n买入金额: ¥%.2f\n成交净值: %.4f\n份额: %.2f", amount, nav, (amount-fee)/nav), win)
		},
		win,
	)
}

// createDetailRow 创建详情行
func (s *SearchUI) createDetailRow(label string, value *widget.Label) fyne.CanvasObject {
	titleLabel := widget.NewLabel(label)
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	value.Alignment = fyne.TextAlignTrailing
	return container.NewBorder(nil, nil, titleLabel, value)
}

// createResultItem 创建搜索结果项
func (s *SearchUI) createResultItem(item model.FundSearchResult) fyne.CanvasObject {
	bg := canvas.NewRectangle(apptheme.GetCardBgColor())
	bg.CornerRadius = 4

	codeLabel := widget.NewLabelWithStyle(item.Code, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	nameLabel := widget.NewLabel(item.Name)
	nameLabel.Truncation = fyne.TextTruncateEllipsis

	typeLabel := widget.NewLabel(item.Type)
	typeLabel.Importance = widget.LowImportance

	content := container.NewBorder(
		nil, nil,
		codeLabel,
		typeLabel,
		nameLabel,
	)

	btn := widget.NewButton("", func() {
		s.showFundDetail(item.Code)
	})
	btn.Importance = widget.LowImportance

	return container.NewStack(bg, container.NewPadded(content), btn)
}

// doSearch 执行搜索
func (s *SearchUI) doSearch(keyword string) {
	if keyword == "" {
		return
	}

	s.resultList.RemoveAll()
	loadingLabel := widget.NewLabel("搜索中...")
	loadingLabel.Alignment = fyne.TextAlignCenter
	s.resultList.Add(container.NewCenter(loadingLabel))
	s.resultList.Refresh()

	go func() {
		results, err := service.GetFundAPI().SearchFund(keyword)

		// 使用fyne的方式在主线程更新UI
		fyne.CurrentApp().Driver().CanvasForObject(s.resultList).Refresh(s.resultList)

		if err != nil {
			s.resultList.RemoveAll()
			errorLabel := widget.NewLabel("搜索失败: " + err.Error())
			s.resultList.Add(container.NewCenter(errorLabel))
			s.resultList.Refresh()
			return
		}

		s.results = results
		s.resultList.RemoveAll()

		if len(results) == 0 {
			noResultLabel := widget.NewLabel("未找到相关基金")
			noResultLabel.Alignment = fyne.TextAlignCenter
			s.resultList.Add(container.NewCenter(noResultLabel))
		} else {
			for _, item := range results {
				s.resultList.Add(s.createResultItem(item))
			}
		}
		s.resultList.Refresh()
	}()
}

// showFundDetail 显示基金详情
func (s *SearchUI) showFundDetail(code string) {
	s.currentCode = code
	s.codeLabel.SetText(code)
	s.nameLabel.SetText("加载中...")
	s.typeLabel.SetText("-")
	s.navLabel.SetText("-")
	s.estLabel.SetText("-")
	s.growthLabel.SetText("-")

	// 清空走势图
	s.chartContainer.RemoveAll()
	loadingLabel := widget.NewLabel("加载走势图...")
	s.chartContainer.Add(container.NewCenter(loadingLabel))
	s.chartContainer.Refresh()

	go func() {
		fund, err := service.GetFundAPI().GetFundDetail(code)
		if err != nil {
			s.nameLabel.SetText("获取失败")
			return
		}

		// 获取类型
		for _, r := range s.results {
			if r.Code == code {
				fund.Type = r.Type
				fund.Name = r.Name
				break
			}
		}

		s.nameLabel.SetText(fund.Name)
		s.typeLabel.SetText(fund.Type)
		s.navLabel.SetText(fmt.Sprintf("%.4f", fund.NetValue))
		s.estLabel.SetText(fmt.Sprintf("%.4f", fund.EstValue))

		if fund.EstGrowth >= 0 {
			s.growthLabel.SetText(fmt.Sprintf("+%.2f%%", fund.EstGrowth))
			s.growthLabel.Importance = widget.SuccessImportance
		} else {
			s.growthLabel.SetText(fmt.Sprintf("%.2f%%", fund.EstGrowth))
			s.growthLabel.Importance = widget.DangerImportance
		}

		// 加载走势图
		s.loadChart(code)
	}()
}

// loadChart 加载走势图
func (s *SearchUI) loadChart(code string) {
	histories, err := service.GetFundAPI().GetFundHistory(code, 30)

	s.chartContainer.RemoveAll()

	if err != nil || len(histories) == 0 {
		errorLabel := widget.NewLabel("无法加载走势数据")
		s.chartContainer.Add(container.NewCenter(errorLabel))
		s.chartContainer.Refresh()
		return
	}

	// 创建简易走势图
	chart := s.createLineChart(histories)
	s.chartContainer.Add(chart)
	s.chartContainer.Refresh()
}

// createLineChart 创建折线图
func (s *SearchUI) createLineChart(histories []model.NetValueHistory) fyne.CanvasObject {
	if len(histories) == 0 {
		return widget.NewLabel("无数据")
	}

	// 找出最大最小值
	var minVal, maxVal float64 = histories[0].NetValue, histories[0].NetValue
	for _, h := range histories {
		if h.NetValue < minVal {
			minVal = h.NetValue
		}
		if h.NetValue > maxVal {
			maxVal = h.NetValue
		}
	}

	// 添加一些边距
	valRange := maxVal - minVal
	if valRange < 0.01 {
		valRange = 0.01
	}
	minVal -= valRange * 0.1
	maxVal += valRange * 0.1
	valRange = maxVal - minVal

	chartWidth := float32(500)
	chartHeight := float32(150)

	// 创建背景
	bg := canvas.NewRectangle(color.RGBA{R: 248, G: 250, B: 252, A: 255})
	bg.SetMinSize(fyne.NewSize(chartWidth, chartHeight))

	// 创建折线
	points := make([]fyne.CanvasObject, 0)

	// 倒序处理(从旧到新)
	n := len(histories)
	for i := n - 1; i >= 0; i-- {
		h := histories[i]
		idx := n - 1 - i

		x := float32(idx) / float32(n-1) * (chartWidth - 20) + 10
		y := chartHeight - 10 - float32((h.NetValue-minVal)/valRange)*(chartHeight-20)

		// 绘制点
		circle := canvas.NewCircle(color.RGBA{R: 64, G: 128, B: 255, A: 255})
		circle.Resize(fyne.NewSize(6, 6))
		circle.Move(fyne.NewPos(x-3, y-3))
		points = append(points, circle)

		// 绘制连线
		if i < n-1 {
			nextIdx := n - 2 - i
			nextX := float32(nextIdx+1) / float32(n-1) * (chartWidth - 20) + 10
			nextY := chartHeight - 10 - float32((histories[i+1].NetValue-minVal)/valRange)*(chartHeight-20)

			line := canvas.NewLine(color.RGBA{R: 64, G: 128, B: 255, A: 200})
			line.StrokeWidth = 2
			line.Position1 = fyne.NewPos(x, y)
			line.Position2 = fyne.NewPos(nextX, nextY)
			points = append(points, line)
		}
	}

	// 添加Y轴标签
	maxLabel := widget.NewLabel(fmt.Sprintf("%.4f", maxVal))
	maxLabel.TextStyle = fyne.TextStyle{}
	minLabel := widget.NewLabel(fmt.Sprintf("%.4f", minVal))
	minLabel.TextStyle = fyne.TextStyle{}

	// 日期标签
	startDate := histories[len(histories)-1].Date.Format("01-02")
	endDate := histories[0].Date.Format("01-02")
	dateLabel := widget.NewLabel(fmt.Sprintf("%s ~ %s", startDate, endDate))
	dateLabel.Alignment = fyne.TextAlignCenter

	// 组合图表
	chartContent := container.NewWithoutLayout(append([]fyne.CanvasObject{bg}, points...)...)
	chartContent.Resize(fyne.NewSize(chartWidth, chartHeight))

	return container.NewVBox(
		container.NewHBox(maxLabel, layout.NewSpacer()),
		chartContent,
		container.NewHBox(minLabel, layout.NewSpacer()),
		dateLabel,
	)
}

// Content 获取内容
func (s *SearchUI) Content() fyne.CanvasObject {
	return s.content
}
