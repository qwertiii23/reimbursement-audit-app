package smartrule

import "time"

type CityTier struct {
	ID        string    `json:"id"`
	CityName  string    `json:"city_name"`
	Tier      int       `json:"tier"`
	Province  string    `json:"province"`
	CreatedAt time.Time `json:"created_at"`
}

type AccommodationStandard struct {
	ID            string    `json:"id"`
	CityTier      int       `json:"city_tier"`
	EmployeeLevel string    `json:"employee_level"`
	DailyLimit    float64   `json:"daily_limit"`
	CreatedAt     time.Time `json:"created_at"`
}

type CityEvent struct {
	ID                      string    `json:"id"`
	CityName                string    `json:"city_name"`
	EventName               string    `json:"event_name"`
	EventType               string    `json:"event_type"`
	StartDate               time.Time `json:"start_date"`
	EndDate                 time.Time `json:"end_date"`
	AccommodationAdjustment float64   `json:"accommodation_adjustment"`
	CreatedAt               time.Time `json:"created_at"`
}

type Holiday struct {
	ID                      string    `json:"id"`
	HolidayName             string    `json:"holiday_name"`
	HolidayDate             time.Time `json:"holiday_date"`
	IsAdjusted              bool      `json:"is_adjusted"`
	AccommodationAdjustment float64   `json:"accommodation_adjustment"`
	CreatedAt               time.Time `json:"created_at"`
}

type PolicyKnowledge struct {
	ID            string    `json:"id"`
	PolicyType    string    `json:"policy_type"`
	PolicySection string    `json:"policy_section"`
	PolicyContent string    `json:"policy_content"`
	Keywords      []string  `json:"keywords"`
	CreatedAt     time.Time `json:"created_at"`
}

type TransportationStandard struct {
	ID              string    `json:"id" gorm:"primaryKey"`
	TransportType   string    `json:"transport_type" gorm:"column:transport_type"`
	EmployeeLevel   string    `json:"employee_level" gorm:"column:employee_level"`
	DailyLimit      float64   `json:"daily_limit" gorm:"column:daily_limit"`
	SingleTripLimit float64   `json:"single_trip_limit" gorm:"column:single_trip_limit"`
	Description     string    `json:"description" gorm:"column:description"`
	CreatedAt       time.Time `json:"created_at" gorm:"column:created_at"`
}

func (TransportationStandard) TableName() string {
	return "transportation_standard"
}

type MealStandard struct {
	ID            string    `json:"id"`
	CityTier      int       `json:"city_tier"`
	EmployeeLevel string    `json:"employee_level"`
	MealType      string    `json:"meal_type"`
	DailyLimit    float64   `json:"daily_limit"`
	PerMealLimit  float64   `json:"per_meal_limit"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
}

type EntertainmentStandard struct {
	ID             string    `json:"id"`
	GuestType      string    `json:"guest_type"`
	EmployeeLevel  string    `json:"employee_level"`
	PerPersonLimit float64   `json:"per_person_limit"`
	DailyLimit     float64   `json:"daily_limit"`
	Description    string    `json:"description"`
	CreatedAt      time.Time `json:"created_at"`
}

type OvertimeStandard struct {
	ID              string    `json:"id"`
	EmployeeLevel   string    `json:"employee_level"`
	HourlyRate      float64   `json:"hourly_rate"`
	DailyMaxHours   float64   `json:"daily_max_hours"`
	MonthlyMaxHours float64   `json:"monthly_max_hours"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
}
