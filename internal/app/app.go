package app

import (
	"sync"
	"time"

	"jijin/internal/repository"
	"jijin/internal/service"
	apptheme "jijin/internal/theme"
	"jijin/internal/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// App 应用程序
type App struct {
	fyneApp    fyne.App
	mainWindow fyne.Window
	content    *fyne.Container

	// UI组件
	homeUI       *ui.HomeUI
	searchUI     *ui.SearchUI
	portfolioUI  *ui.PortfolioUI
	calculatorUI *ui.CalculatorUI
	strategyUI   *ui.StrategyUI
	analysisUI   *ui.AnalysisUI
	toolsUI      *ui.ToolsUI
	alertUI      *ui.AlertUI

	// 自动刷新
	refreshTicker *time.Ticker
	stopRefresh   chan bool
	refreshMu     sync.Mutex
	isRefreshing  bool

	// 状态栏
	statusLabel *widget.Label
	lastUpdate  time.Time
}

// NewApp 创建应用
func NewApp() *App {
	return &App{
		stopRefresh: make(chan bool),
	}
}

// Run 运行应用
func (a *App) Run() error {
	// 初始化数据库
	if err := repository.InitDB(); err != nil {
		return err
	}

	// 创建Fyne应用
	a.fyneApp = app.New()
	a.fyneApp.Settings().SetTheme(&apptheme.AppTheme{})

	a.mainWindow = a.fyneApp.NewWindow("基金助手")
	a.mainWindow.Resize(fyne.NewSize(1100, 750))
	a.mainWindow.CenterOnScreen()

	// 创建UI组件
	a.createUI()

	// 启动自动刷新(每5分钟)
	a.startAutoRefresh(5 * time.Minute)

	// 设置关闭处理
	a.mainWindow.SetOnClosed(func() {
		a.stopAutoRefresh()
	})

	a.mainWindow.ShowAndRun()
	return nil
}

// createUI 创建UI
func (a *App) createUI() {
	// 创建各个页面
	a.homeUI = ui.NewHomeUI(a.onRefresh)
	a.searchUI = ui.NewSearchUI(a.onAddHolding)
	a.portfolioUI = ui.NewPortfolioUI()
	a.calculatorUI = ui.NewCalculatorUI()
	a.strategyUI = ui.NewStrategyUI()
	a.analysisUI = ui.NewAnalysisUI()
	a.toolsUI = ui.NewToolsUI()
	a.alertUI = ui.NewAlertUI()

	// 设置窗口引用（用于显示对话框）
	a.toolsUI.SetWindow(a.mainWindow)
	a.alertUI.SetWindow(a.mainWindow)

	// 创建标签页
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("首页", theme.HomeIcon(), a.homeUI.Content()),
		container.NewTabItemWithIcon("搜索", theme.SearchIcon(), a.searchUI.Content()),
		container.NewTabItemWithIcon("持仓", theme.ListIcon(), a.portfolioUI.Content()),
		container.NewTabItemWithIcon("工具箱", theme.SettingsIcon(), a.toolsUI.Content()),
		container.NewTabItemWithIcon("提醒", theme.WarningIcon(), a.alertUI.Content()),
		container.NewTabItemWithIcon("分析", theme.DocumentIcon(), a.analysisUI.Content()),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	// 当切换标签时刷新数据
	tabs.OnSelected = func(ti *container.TabItem) {
		switch ti.Text {
		case "首页":
			a.homeUI.Refresh()
		case "持仓":
			a.portfolioUI.Refresh()
		case "工具箱":
			a.toolsUI.Refresh()
		case "提醒":
			a.alertUI.Refresh()
		case "分析":
			a.analysisUI.Refresh()
		}
	}

	// 状态栏
	a.statusLabel = widget.NewLabel("就绪")
	a.statusLabel.TextStyle = fyne.TextStyle{Italic: true}

	refreshBtn := widget.NewButtonWithIcon("刷新数据", theme.ViewRefreshIcon(), a.onRefresh)
	refreshBtn.Importance = widget.HighImportance

	// 状态栏背景
	statusBg := canvas.NewRectangle(apptheme.GetCardBgColor())
	statusContent := container.NewBorder(
		nil, nil,
		container.NewHBox(
			widget.NewIcon(theme.InfoIcon()),
			a.statusLabel,
		),
		refreshBtn,
	)
	statusBar := container.NewStack(statusBg, container.NewPadded(statusContent))

	// 主布局
	mainContent := container.NewBorder(
		nil,
		statusBar,
		nil,
		nil,
		tabs,
	)

	a.mainWindow.SetContent(mainContent)

	// 初始加载
	a.homeUI.Refresh()
}

// onRefresh 手动刷新
func (a *App) onRefresh() {
	a.refreshMu.Lock()
	if a.isRefreshing {
		a.refreshMu.Unlock()
		return
	}
	a.isRefreshing = true
	a.refreshMu.Unlock()

	a.statusLabel.SetText("正在刷新数据...")

	go func() {
		defer func() {
			a.refreshMu.Lock()
			a.isRefreshing = false
			a.refreshMu.Unlock()
		}()

		err := service.GetFundAPI().RefreshAllHoldings()
		if err != nil {
			a.statusLabel.SetText("刷新失败: " + err.Error())
			return
		}

		a.lastUpdate = time.Now()
		a.statusLabel.SetText("上次更新: " + a.lastUpdate.Format("15:04:05"))

		// 刷新UI
		a.homeUI.Refresh()
		a.portfolioUI.Refresh()
		a.analysisUI.Refresh()
	}()
}

// onAddHolding 添加持仓回调
func (a *App) onAddHolding(code, name string, amount, nav, fee float64) {
	// 先添加持仓记录
	service.GetPortfolioService().AddHolding(code, name)
	// 再执行买入
	service.GetPortfolioService().Buy(code, amount, nav, fee, time.Now())
	// 刷新UI(在主线程)
	a.portfolioUI.Refresh()
	a.homeUI.Refresh()
}

// startAutoRefresh 启动自动刷新
func (a *App) startAutoRefresh(interval time.Duration) {
	a.refreshTicker = time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-a.refreshTicker.C:
				// 只在交易时间刷新(9:30-15:00)
				now := time.Now()
				hour := now.Hour()
				minute := now.Minute()
				weekday := now.Weekday()

				// 周末不刷新
				if weekday == time.Saturday || weekday == time.Sunday {
					continue
				}

				// 非交易时间不刷新
				if hour < 9 || (hour == 9 && minute < 30) || hour >= 15 {
					continue
				}

				a.onRefresh()

			case <-a.stopRefresh:
				return
			}
		}
	}()
}

// stopAutoRefresh 停止自动刷新
func (a *App) stopAutoRefresh() {
	if a.refreshTicker != nil {
		a.refreshTicker.Stop()
	}
	close(a.stopRefresh)
}
