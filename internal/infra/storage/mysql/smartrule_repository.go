package mysql

import (
	"context"
	"reimbursement-audit/internal/domain/smartrule"
	"reimbursement-audit/internal/pkg/logger"
	"time"
)

type CityTierRepository struct {
	client *Client
	logger logger.Logger
}

func NewCityTierRepository(client *Client, logger logger.Logger) *CityTierRepository {
	return &CityTierRepository{
		client: client,
		logger: logger,
	}
}

func (r *CityTierRepository) GetCityTier(ctx context.Context, cityName string) (*smartrule.CityTier, error) {
	var cityTier smartrule.CityTier
	err := r.client.GetDB().WithContext(ctx).Where("city_name = ?", cityName).First(&cityTier).Error
	if err != nil {
		return nil, err
	}
	return &cityTier, nil
}

type AccommodationStandardRepository struct {
	client *Client
	logger logger.Logger
}

func NewAccommodationStandardRepository(client *Client, logger logger.Logger) *AccommodationStandardRepository {
	return &AccommodationStandardRepository{
		client: client,
		logger: logger,
	}
}

func (r *AccommodationStandardRepository) GetStandard(ctx context.Context, cityTier int, employeeLevel string) (*smartrule.AccommodationStandard, error) {
	var standard smartrule.AccommodationStandard
	err := r.client.GetDB().WithContext(ctx).
		Where("city_tier = ? AND employee_level = ?", cityTier, employeeLevel).
		First(&standard).Error
	if err != nil {
		return nil, err
	}
	return &standard, nil
}

type MealStandardRepository struct {
	client *Client
	logger logger.Logger
}

func NewMealStandardRepository(client *Client, logger logger.Logger) *MealStandardRepository {
	return &MealStandardRepository{
		client: client,
		logger: logger,
	}
}

func (r *MealStandardRepository) GetStandard(ctx context.Context, cityTier int, employeeLevel string, mealType string) (*smartrule.MealStandard, error) {
	var standard smartrule.MealStandard
	err := r.client.GetDB().WithContext(ctx).
		Where("city_tier = ? AND employee_level = ? AND meal_type = ?", cityTier, employeeLevel, mealType).
		First(&standard).Error
	if err != nil {
		return nil, err
	}
	return &standard, nil
}

type EntertainmentStandardRepository struct {
	client *Client
	logger logger.Logger
}

func NewEntertainmentStandardRepository(client *Client, logger logger.Logger) *EntertainmentStandardRepository {
	return &EntertainmentStandardRepository{
		client: client,
		logger: logger,
	}
}

func (r *EntertainmentStandardRepository) GetStandard(ctx context.Context, guestType string, employeeLevel string) (*smartrule.EntertainmentStandard, error) {
	var standard smartrule.EntertainmentStandard
	err := r.client.GetDB().WithContext(ctx).
		Where("guest_type = ? AND employee_level = ?", guestType, employeeLevel).
		First(&standard).Error
	if err != nil {
		return nil, err
	}
	return &standard, nil
}

type OvertimeStandardRepository struct {
	client *Client
	logger logger.Logger
}

func NewOvertimeStandardRepository(client *Client, logger logger.Logger) *OvertimeStandardRepository {
	return &OvertimeStandardRepository{
		client: client,
		logger: logger,
	}
}

func (r *OvertimeStandardRepository) GetStandard(ctx context.Context, employeeLevel string) (*smartrule.OvertimeStandard, error) {
	var standard smartrule.OvertimeStandard
	err := r.client.GetDB().WithContext(ctx).
		Where("employee_level = ?", employeeLevel).
		First(&standard).Error
	if err != nil {
		return nil, err
	}
	return &standard, nil
}

type CityEventRepository struct {
	client *Client
	logger logger.Logger
}

func NewCityEventRepository(client *Client, logger logger.Logger) *CityEventRepository {
	return &CityEventRepository{
		client: client,
		logger: logger,
	}
}

func (r *CityEventRepository) GetEventsByCityAndDateRange(ctx context.Context, cityName string, startDate, endDate time.Time) ([]*smartrule.CityEvent, error) {
	var events []*smartrule.CityEvent
	err := r.client.GetDB().WithContext(ctx).
		Where("city_name = ? AND start_date <= ? AND end_date >= ?", cityName, endDate, startDate).
		Find(&events).Error
	if err != nil {
		return nil, err
	}
	return events, nil
}

type HolidayRepository struct {
	client *Client
	logger logger.Logger
}

func NewHolidayRepository(client *Client, logger logger.Logger) *HolidayRepository {
	return &HolidayRepository{
		client: client,
		logger: logger,
	}
}

func (r *HolidayRepository) GetHolidaysByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*smartrule.Holiday, error) {
	var holidays []*smartrule.Holiday
	err := r.client.GetDB().WithContext(ctx).
		Where("holiday_date >= ? AND holiday_date <= ?", startDate, endDate).
		Find(&holidays).Error
	if err != nil {
		return nil, err
	}
	return holidays, nil
}

type PolicyKnowledgeRepository struct {
	client *Client
	logger logger.Logger
}

func NewPolicyKnowledgeRepository(client *Client, logger logger.Logger) *PolicyKnowledgeRepository {
	return &PolicyKnowledgeRepository{
		client: client,
		logger: logger,
	}
}

func (r *PolicyKnowledgeRepository) GetPoliciesByType(ctx context.Context, policyType string) ([]*smartrule.PolicyKnowledge, error) {
	var policies []*smartrule.PolicyKnowledge
	err := r.client.GetDB().WithContext(ctx).
		Where("policy_type = ?", policyType).
		Find(&policies).Error
	if err != nil {
		return nil, err
	}
	return policies, nil
}

func (r *PolicyKnowledgeRepository) SearchPolicies(ctx context.Context, keywords []string) ([]*smartrule.PolicyKnowledge, error) {
	var policies []*smartrule.PolicyKnowledge

	query := r.client.GetDB().WithContext(ctx)
	for _, keyword := range keywords {
		query = query.Or("JSON_CONTAINS(keywords, ?)", `"`+keyword+`"`)
	}

	err := query.Find(&policies).Error
	if err != nil {
		return nil, err
	}
	return policies, nil
}

type TransportationStandardRepository struct {
	client *Client
	logger logger.Logger
}

func NewTransportationStandardRepository(client *Client, logger logger.Logger) *TransportationStandardRepository {
	return &TransportationStandardRepository{
		client: client,
		logger: logger,
	}
}

func (r *TransportationStandardRepository) GetStandard(ctx context.Context, transportType string, employeeLevel string) (*smartrule.TransportationStandard, error) {
	var standard smartrule.TransportationStandard
	err := r.client.GetDB().WithContext(ctx).
		Where("transport_type = ? AND employee_level = ?", transportType, employeeLevel).
		First(&standard).Error
	if err != nil {
		return nil, err
	}
	return &standard, nil
}
