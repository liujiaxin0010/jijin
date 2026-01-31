package service

import (
	"math"
	"time"

	"jijin/internal/model"
	"jijin/internal/repository"
)

// CalculatorService 定投计算服务
type CalculatorService struct{}

var calculatorService = &CalculatorService{}

// GetCalculatorService 获取计算服务实例
func GetCalculatorService() *CalculatorService {
	return calculatorService
}

// InvestmentResult 定投计算结果
type InvestmentResult struct {
	TotalInvest    float64   // 总投入
	TotalShares    float64   // 总份额
	CurrentValue   float64   // 当前市值
	TotalProfit    float64   // 总收益
	ProfitRate     float64   // 收益率(%)
	AnnualReturn   float64   // 年化收益率(%)
	InvestCount    int       // 投资次数
	StartDate      time.Time // 开始日期
	EndDate        time.Time // 结束日期
	AvgCost        float64   // 平均成本
}

// CalculateInvestment 计算定投收益(使用历史数据)
func (c *CalculatorService) CalculateInvestment(fundCode string, amount float64, frequency string, startDate, endDate time.Time) (*InvestmentResult, error) {
	// 获取历史净值数据
	days := int(endDate.Sub(startDate).Hours()/24) + 30
	histories, err := GetFundAPI().GetFundHistory(fundCode, days)
	if err != nil {
		return nil, err
	}

	if len(histories) == 0 {
		return nil, nil
	}

	// 构建日期到净值的映射
	navMap := make(map[string]float64)
	for _, h := range histories {
		dateStr := h.Date.Format("2006-01-02")
		navMap[dateStr] = h.NetValue
	}

	result := &InvestmentResult{
		StartDate: startDate,
		EndDate:   endDate,
	}

	// 模拟定投
	current := startDate
	for !current.After(endDate) {
		dateStr := current.Format("2006-01-02")

		// 查找最近的交易日净值
		nav := c.findNearestNav(navMap, current, histories)
		if nav > 0 {
			shares := amount / nav
			result.TotalInvest += amount
			result.TotalShares += shares
			result.InvestCount++
		}

		// 计算下一次定投日期
		switch frequency {
		case "daily":
			current = current.AddDate(0, 0, 1)
		case "weekly":
			current = current.AddDate(0, 0, 7)
		case "monthly":
			current = current.AddDate(0, 1, 0)
		default:
			current = current.AddDate(0, 1, 0)
		}

		_ = dateStr // 避免未使用警告
	}

	// 计算最终收益
	if len(histories) > 0 && result.TotalShares > 0 {
		latestNav := histories[0].NetValue // 历史数据按日期降序排列
		result.CurrentValue = result.TotalShares * latestNav
		result.TotalProfit = result.CurrentValue - result.TotalInvest
		result.AvgCost = result.TotalInvest / result.TotalShares

		if result.TotalInvest > 0 {
			result.ProfitRate = result.TotalProfit / result.TotalInvest * 100
		}

		// 计算年化收益率
		years := endDate.Sub(startDate).Hours() / 24 / 365
		if years > 0 && result.TotalInvest > 0 {
			result.AnnualReturn = (math.Pow(result.CurrentValue/result.TotalInvest, 1/years) - 1) * 100
		}
	}

	return result, nil
}

// findNearestNav 查找最近交易日的净值
func (c *CalculatorService) findNearestNav(navMap map[string]float64, date time.Time, histories []model.NetValueHistory) float64 {
	// 先尝试当天
	dateStr := date.Format("2006-01-02")
	if nav, ok := navMap[dateStr]; ok {
		return nav
	}

	// 向前查找最近的交易日(最多7天)
	for i := 1; i <= 7; i++ {
		prevDate := date.AddDate(0, 0, -i)
		dateStr := prevDate.Format("2006-01-02")
		if nav, ok := navMap[dateStr]; ok {
			return nav
		}
	}

	return 0
}

// CalculateCurrentProfit 计算当前持仓收益
func (c *CalculatorService) CalculateCurrentProfit(holding *model.Holding) (profit, profitRate float64) {
	if holding.Cost == 0 {
		return 0, 0
	}

	currentValue := holding.Shares * holding.CurrentNav
	profit = currentValue - holding.Cost
	profitRate = profit / holding.Cost * 100

	return
}

// CalculateAnnualReturn 计算年化收益率
func (c *CalculatorService) CalculateAnnualReturn(holding *model.Holding) float64 {
	if holding.Cost == 0 || holding.CreatedAt.IsZero() {
		return 0
	}

	years := time.Since(holding.CreatedAt).Hours() / 24 / 365
	if years < 0.01 {
		return 0
	}

	currentValue := holding.Shares * holding.CurrentNav
	return (math.Pow(currentValue/holding.Cost, 1/years) - 1) * 100
}

// GetHistoryData 获取历史净值数据(用于图表)
func (c *CalculatorService) GetHistoryData(fundCode string, days int) (dates []string, values []float64, err error) {
	histories, err := repository.GetNetValueHistory(fundCode, days)
	if err != nil || len(histories) == 0 {
		// 尝试从网络获取
		histories, err = GetFundAPI().GetFundHistory(fundCode, days)
		if err != nil {
			return nil, nil, err
		}
		// 保存到数据库
		repository.SaveNetValueHistories(histories)
	}

	// 倒序(从旧到新)
	for i := len(histories) - 1; i >= 0; i-- {
		h := histories[i]
		dates = append(dates, h.Date.Format("01-02"))
		values = append(values, h.NetValue)
	}

	return
}
