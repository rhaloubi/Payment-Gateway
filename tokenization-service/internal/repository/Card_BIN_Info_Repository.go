package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/rhaloubi/payment-gateway/tokenization-service/inits"
	model "github.com/rhaloubi/payment-gateway/tokenization-service/internal/models"
	"gorm.io/gorm"
)

type CardBINRepository struct{}

func NewCardBINRepository() *CardBINRepository {
	return &CardBINRepository{}
}

func (r *CardBINRepository) Create(binInfo *model.CardBINInfo) error {
	return inits.DB.Create(binInfo).Error
}

// FindByBIN finds card info by BIN (first 6 digits)
func (r *CardBINRepository) FindByBIN(bin string) (*model.CardBINInfo, error) {
	cacheKey := fmt.Sprintf("bin:%s", bin)
	cachedData, err := inits.RDB.Get(inits.Ctx, cacheKey).Result()

	if err == nil && cachedData != "" {
		var binInfo model.CardBINInfo
		if err = json.Unmarshal([]byte(cachedData), &binInfo); err == nil {
			return &binInfo, nil
		}
	}

	var binInfo model.CardBINInfo
	err = inits.DB.Where("bin = ?", bin).First(&binInfo).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	data, _ := json.Marshal(binInfo)
	inits.RDB.Set(inits.Ctx, cacheKey, data, 24*time.Hour)

	return &binInfo, nil
}

// Update updates BIN information
func (r *CardBINRepository) Update(binInfo *model.CardBINInfo) error {
	err := inits.DB.Save(binInfo).Error
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("bin:%s", binInfo.BIN)
	inits.RDB.Del(inits.Ctx, cacheKey)

	return nil
}

func (r *CardBINRepository) Delete(bin string) error {
	err := inits.DB.Where("bin = ?", bin).Delete(&model.CardBINInfo{}).Error
	if err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("bin:%s", bin)
	inits.RDB.Del(inits.Ctx, cacheKey)

	return nil
}

func (r *CardBINRepository) FindByCardBrand(cardBrand model.CardBrand) ([]model.CardBINInfo, error) {
	var bins []model.CardBINInfo
	err := inits.DB.Where("card_brand = ?", cardBrand).Find(&bins).Error
	return bins, err
}

// BulkCreate creates multiple BIN entries at once
func (r *CardBINRepository) BulkCreate(binInfos []model.CardBINInfo) error {
	return inits.DB.CreateInBatches(binInfos, 100).Error
}
