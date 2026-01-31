package service

import (
	"jijin/internal/model"
	"jijin/internal/repository"
)

// QDIIService QDII基金服务
type QDIIService struct{}

var qdiiService = &QDIIService{}

// GetQDIIService 获取QDII服务实例
func GetQDIIService() *QDIIService {
	return qdiiService
}

// GetQDIIFund 获取QDII基金信息
func (q *QDIIService) GetQDIIFund(fundCode string) (*model.QDIIFund, error) {
	return repository.GetQDIIFund(fundCode)
}
