package service

import (
	"errors"
	"time"

	"jijin/internal/model"
	"jijin/internal/repository"
)

// PortfolioService 持仓服务
type PortfolioService struct{}

var portfolioService = &PortfolioService{}

// GetPortfolioService 获取持仓服务实例
func GetPortfolioService() *PortfolioService {
	return portfolioService
}

// AddHolding 添加持仓
func (p *PortfolioService) AddHolding(fundCode, fundName string) (*model.Holding, error) {
	// 检查是否已存在
	existing, _ := repository.GetHoldingByFundCode(fundCode)
	if existing != nil {
		return existing, nil
	}

	holding := &model.Holding{
		FundCode: fundCode,
		FundName: fundName,
		Shares:   0,
		Cost:     0,
	}

	if err := repository.SaveHolding(holding); err != nil {
		return nil, err
	}

	return holding, nil
}

// Buy 买入
func (p *PortfolioService) Buy(fundCode string, amount, netValue, fee float64, tradeDate time.Time) error {
	holding, err := repository.GetHoldingByFundCode(fundCode)
	if err != nil {
		return errors.New("持仓不存在，请先添加持仓")
	}

	// 计算份额
	actualAmount := amount - fee
	shares := actualAmount / netValue

	// 更新持仓
	holding.Shares += shares
	holding.Cost += amount
	if holding.Shares > 0 {
		holding.CostPrice = holding.Cost / holding.Shares
	}
	holding.CurrentNav = netValue

	if err := repository.SaveHolding(holding); err != nil {
		return err
	}

	// 记录交易
	tx := &model.Transaction{
		FundCode:  fundCode,
		FundName:  holding.FundName,
		Type:      "buy",
		Amount:    amount,
		NetValue:  netValue,
		Shares:    shares,
		Fee:       fee,
		TradeDate: tradeDate,
	}

	return repository.SaveTransaction(tx)
}

// Sell 卖出
func (p *PortfolioService) Sell(fundCode string, shares, netValue, fee float64, tradeDate time.Time) error {
	holding, err := repository.GetHoldingByFundCode(fundCode)
	if err != nil {
		return errors.New("持仓不存在")
	}

	if holding.Shares < shares {
		return errors.New("卖出份额超过持有份额")
	}

	// 计算金额
	amount := shares * netValue
	actualAmount := amount - fee

	// 更新持仓
	costRatio := shares / holding.Shares
	holding.Cost -= holding.Cost * costRatio
	holding.Shares -= shares
	if holding.Shares > 0 {
		holding.CostPrice = holding.Cost / holding.Shares
	} else {
		holding.CostPrice = 0
	}
	holding.CurrentNav = netValue

	if err := repository.SaveHolding(holding); err != nil {
		return err
	}

	// 记录交易
	tx := &model.Transaction{
		FundCode:  fundCode,
		FundName:  holding.FundName,
		Type:      "sell",
		Amount:    actualAmount,
		NetValue:  netValue,
		Shares:    shares,
		Fee:       fee,
		TradeDate: tradeDate,
	}

	return repository.SaveTransaction(tx)
}

// GetAllHoldings 获取所有持仓
func (p *PortfolioService) GetAllHoldings() ([]model.Holding, error) {
	return repository.GetAllHoldings()
}

// GetHolding 获取持仓详情
func (p *PortfolioService) GetHolding(id uint) (*model.Holding, error) {
	return repository.GetHolding(id)
}

// DeleteHolding 删除持仓
func (p *PortfolioService) DeleteHolding(id uint) error {
	return repository.DeleteHolding(id)
}

// GetTransactions 获取交易记录
func (p *PortfolioService) GetTransactions(fundCode string) ([]model.Transaction, error) {
	if fundCode == "" {
		return repository.GetAllTransactions()
	}
	return repository.GetTransactionsByFundCode(fundCode)
}

// GetPortfolioSummary 获取持仓汇总
func (p *PortfolioService) GetPortfolioSummary() (totalCost, totalValue, totalProfit float64, profitRate float64) {
	holdings, err := repository.GetAllHoldings()
	if err != nil {
		return
	}

	for _, h := range holdings {
		totalCost += h.Cost
		totalValue += h.MarketValue()
	}

	totalProfit = totalValue - totalCost
	if totalCost > 0 {
		profitRate = totalProfit / totalCost * 100
	}

	return
}
