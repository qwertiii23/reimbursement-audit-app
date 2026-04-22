package smartrule

import (
	"context"
	"time"
)

type CityTierRepository interface {
	GetCityTier(ctx context.Context, cityName string) (*CityTier, error)
}

type AccommodationStandardRepository interface {
	GetStandard(ctx context.Context, cityTier int, employeeLevel string) (*AccommodationStandard, error)
}

type CityEventRepository interface {
	GetEventsByCityAndDateRange(ctx context.Context, cityName string, startDate, endDate time.Time) ([]*CityEvent, error)
}

type HolidayRepository interface {
	GetHolidaysByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*Holiday, error)
}

type PolicyKnowledgeRepository interface {
	GetPoliciesByType(ctx context.Context, policyType string) ([]*PolicyKnowledge, error)
	SearchPolicies(ctx context.Context, keywords []string) ([]*PolicyKnowledge, error)
}

type TransportationStandardRepository interface {
	GetStandard(ctx context.Context, transportType string, employeeLevel string) (*TransportationStandard, error)
}

type MealStandardRepository interface {
	GetStandard(ctx context.Context, cityTier int, employeeLevel string, mealType string) (*MealStandard, error)
}

type EntertainmentStandardRepository interface {
	GetStandard(ctx context.Context, guestType string, employeeLevel string) (*EntertainmentStandard, error)
}

type OvertimeStandardRepository interface {
	GetStandard(ctx context.Context, employeeLevel string) (*OvertimeStandard, error)
}
