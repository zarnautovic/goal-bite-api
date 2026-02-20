package handlers

import "time"

type HealthResponse struct {
	// Service status.
	Status string `json:"status" example:"ok"`
	// Human-readable health message.
	Message string `json:"message" example:"nutrition api is running"`
}

type UserResponse struct {
	// User ID.
	ID uint `json:"id" example:"1"`
	// Display name.
	Name string `json:"name" example:"Test User"`
	// Unique normalized email address.
	Email string `json:"email" example:"john@gmail.com"`
	// Optional biological sex.
	Sex *string `json:"sex,omitempty" example:"male"`
	// Optional birth date.
	BirthDate *time.Time `json:"birth_date,omitempty" example:"1994-05-18T00:00:00Z"`
	// Optional height in centimeters.
	HeightCM *float64 `json:"height_cm,omitempty" example:"178"`
	// Optional activity level.
	ActivityLevel *string `json:"activity_level,omitempty" example:"moderate"`
	// Creation timestamp in RFC3339 UTC.
	CreatedAt time.Time `json:"created_at" example:"2026-02-17T12:00:00Z"`
	// Last update timestamp in RFC3339 UTC.
	UpdatedAt time.Time `json:"updated_at" example:"2026-02-17T12:00:00Z"`
}

type FoodResponse struct {
	// Food ID.
	ID uint `json:"id" example:"1"`
	// Food name.
	Name string `json:"name" example:"Rice"`
	// Optional brand name.
	BrandName *string `json:"brand_name,omitempty" example:"Fage"`
	// Optional product barcode.
	Barcode *string `json:"barcode,omitempty" example:"5901234123457"`
	// Energy in kcal per 100g.
	KcalPer100g float64 `json:"kcal_per_100g" example:"130"`
	// Protein grams per 100g.
	ProteinPer100g float64 `json:"protein_per_100g" example:"2.7"`
	// Carbohydrate grams per 100g.
	CarbsPer100g float64 `json:"carbs_per_100g" example:"28"`
	// Fat grams per 100g.
	FatPer100g float64 `json:"fat_per_100g" example:"0.3"`
	// Creation timestamp in RFC3339 UTC.
	CreatedAt time.Time `json:"created_at" example:"2026-02-17T12:00:00Z"`
	// Last update timestamp in RFC3339 UTC.
	UpdatedAt time.Time `json:"updated_at" example:"2026-02-17T12:00:00Z"`
}

type RecipeIngredientResponse struct {
	// Ingredient row ID.
	ID uint `json:"id" example:"1"`
	// Parent recipe ID.
	RecipeID uint `json:"recipe_id" example:"1"`
	// Referenced food ID.
	FoodID uint `json:"food_id" example:"1"`
	// Raw ingredient weight in grams.
	RawWeightG float64 `json:"raw_weight_g" example:"200"`
	// Optional ordering position.
	Position *int `json:"position,omitempty" example:"1"`
	// Creation timestamp in RFC3339 UTC.
	CreatedAt time.Time `json:"created_at" example:"2026-02-17T12:00:00Z"`
	// Last update timestamp in RFC3339 UTC.
	UpdatedAt time.Time `json:"updated_at" example:"2026-02-17T12:00:00Z"`
}

type RecipeResponse struct {
	// Recipe ID.
	ID uint `json:"id" example:"1"`
	// Recipe name.
	Name string `json:"name" example:"Rice Bowl"`
	// Final cooked yield weight in grams.
	YieldWeightG float64 `json:"yield_weight_g" example:"200"`
	// Energy in kcal per 100g.
	KcalPer100g float64 `json:"kcal_per_100g" example:"130"`
	// Protein grams per 100g.
	ProteinPer100g float64 `json:"protein_per_100g" example:"2.7"`
	// Carbohydrate grams per 100g.
	CarbsPer100g float64 `json:"carbs_per_100g" example:"28"`
	// Fat grams per 100g.
	FatPer100g float64 `json:"fat_per_100g" example:"0.3"`
	// Creation timestamp in RFC3339 UTC.
	CreatedAt time.Time `json:"created_at" example:"2026-02-17T12:00:00Z"`
	// Last update timestamp in RFC3339 UTC.
	UpdatedAt   time.Time                  `json:"updated_at" example:"2026-02-17T12:00:00Z"`
	Ingredients []RecipeIngredientResponse `json:"ingredients,omitempty"`
}

type MealItemResponse struct {
	// Meal item ID.
	ID uint `json:"id" example:"1"`
	// Parent meal ID.
	MealID uint `json:"meal_id" example:"1"`
	// Optional food source ID.
	FoodID *uint `json:"food_id,omitempty" example:"1"`
	// Optional recipe source ID.
	RecipeID *uint `json:"recipe_id,omitempty" example:"1"`
	// Consumed weight in grams.
	WeightG float64 `json:"weight_g" example:"150"`
	// Energy snapshot in kcal per 100g at log time.
	KcalPer100g float64 `json:"kcal_per_100g" example:"130"`
	// Protein snapshot in g per 100g at log time.
	ProteinPer100g float64 `json:"protein_per_100g" example:"2.7"`
	// Carbohydrate snapshot in g per 100g at log time.
	CarbsPer100g float64 `json:"carbs_per_100g" example:"28"`
	// Fat snapshot in g per 100g at log time.
	FatPer100g float64 `json:"fat_per_100g" example:"0.3"`
	// Creation timestamp in RFC3339 UTC.
	CreatedAt time.Time `json:"created_at" example:"2026-02-17T12:00:00Z"`
	// Last update timestamp in RFC3339 UTC.
	UpdatedAt time.Time `json:"updated_at" example:"2026-02-17T12:00:00Z"`
}

type MealResponse struct {
	// Meal ID.
	ID uint `json:"id" example:"1"`
	// Owner user ID.
	UserID uint `json:"user_id" example:"1"`
	// Meal type.
	MealType string `json:"meal_type" example:"lunch"`
	// Meal timestamp in RFC3339 UTC.
	EatenAt time.Time `json:"eaten_at" example:"2026-02-17T12:00:00Z"`
	// Creation timestamp in RFC3339 UTC.
	CreatedAt time.Time `json:"created_at" example:"2026-02-17T12:00:00Z"`
	// Last update timestamp in RFC3339 UTC.
	UpdatedAt time.Time          `json:"updated_at" example:"2026-02-17T12:00:00Z"`
	Items     []MealItemResponse `json:"items,omitempty"`
	// Aggregated kcal for this meal.
	TotalKcal float64 `json:"total_kcal" example:"350"`
	// Aggregated protein grams for this meal.
	TotalProteinG float64 `json:"total_protein_g" example:"10.5"`
	// Aggregated carbohydrate grams for this meal.
	TotalCarbsG float64 `json:"total_carbs_g" example:"75"`
	// Aggregated fat grams for this meal.
	TotalFatG float64 `json:"total_fat_g" example:"2.1"`
}

type DailyTotalsResponse struct {
	// Target date in YYYY-MM-DD.
	Date string `json:"date" example:"2026-02-17"`
	// Aggregated kcal for the day.
	TotalKcal float64 `json:"total_kcal" example:"2100"`
	// Aggregated protein grams for the day.
	TotalProteinG float64 `json:"total_protein_g" example:"140"`
	// Aggregated carbohydrate grams for the day.
	TotalCarbsG float64 `json:"total_carbs_g" example:"220"`
	// Aggregated fat grams for the day.
	TotalFatG float64 `json:"total_fat_g" example:"70"`
}

type BodyWeightLogResponse struct {
	// Body weight log ID.
	ID uint `json:"id" example:"1"`
	// Owner user ID.
	UserID uint `json:"user_id" example:"1"`
	// Body weight in kilograms.
	WeightKG float64 `json:"weight_kg" example:"85.2"`
	// Measurement timestamp in RFC3339 UTC.
	LoggedAt time.Time `json:"logged_at" example:"2026-02-17T08:00:00Z"`
	// Creation timestamp in RFC3339 UTC.
	CreatedAt time.Time `json:"created_at" example:"2026-02-17T08:00:00Z"`
	// Last update timestamp in RFC3339 UTC.
	UpdatedAt time.Time `json:"updated_at" example:"2026-02-17T08:00:00Z"`
}

type UserGoalResponse struct {
	// User goal ID.
	ID uint `json:"id" example:"1"`
	// Owner user ID.
	UserID uint `json:"user_id" example:"1"`
	// Target kcal per day.
	TargetKcal float64 `json:"target_kcal" example:"2200"`
	// Target protein grams per day.
	TargetProteinG float64 `json:"target_protein_g" example:"150"`
	// Target carbs grams per day.
	TargetCarbsG float64 `json:"target_carbs_g" example:"220"`
	// Target fat grams per day.
	TargetFatG float64 `json:"target_fat_g" example:"70"`
	// Optional weight goal in kilograms.
	WeightGoalKG *float64 `json:"weight_goal_kg,omitempty" example:"80"`
	// Optional activity level.
	ActivityLevel *string `json:"activity_level,omitempty" example:"moderate"`
	// Creation timestamp in RFC3339 UTC.
	CreatedAt time.Time `json:"created_at" example:"2026-02-17T08:00:00Z"`
	// Last update timestamp in RFC3339 UTC.
	UpdatedAt time.Time `json:"updated_at" example:"2026-02-17T08:00:00Z"`
}

type DailyProgressResponse struct {
	// Target date in YYYY-MM-DD.
	Date string `json:"date" example:"2026-02-17"`
	// Aggregated kcal for the day.
	TotalKcal float64 `json:"total_kcal" example:"1850"`
	// Aggregated protein grams for the day.
	TotalProteinG float64 `json:"total_protein_g" example:"120"`
	// Aggregated carbohydrate grams for the day.
	TotalCarbsG float64 `json:"total_carbs_g" example:"200"`
	// Aggregated fat grams for the day.
	TotalFatG float64 `json:"total_fat_g" example:"60"`
	// Target kcal for the day.
	TargetKcal float64 `json:"target_kcal" example:"2200"`
	// Target protein grams for the day.
	TargetProteinG float64 `json:"target_protein_g" example:"150"`
	// Target carbohydrate grams for the day.
	TargetCarbsG float64 `json:"target_carbs_g" example:"220"`
	// Target fat grams for the day.
	TargetFatG float64 `json:"target_fat_g" example:"70"`
	// Remaining kcal (target - total).
	RemainingKcal float64 `json:"remaining_kcal" example:"350"`
	// Remaining protein grams (target - total).
	RemainingProteinG float64 `json:"remaining_protein_g" example:"30"`
	// Remaining carbs grams (target - total).
	RemainingCarbsG float64 `json:"remaining_carbs_g" example:"20"`
	// Remaining fat grams (target - total).
	RemainingFatG float64 `json:"remaining_fat_g" example:"10"`
}

type EnergyProgressResponse struct {
	From                 string   `json:"from" example:"2026-01-22"`
	To                   string   `json:"to" example:"2026-02-18"`
	AvgIntakeKcal        float64  `json:"avg_intake_kcal" example:"2150"`
	WeightTrendKgPerWeek float64  `json:"weight_trend_kg_per_week" example:"-0.25"`
	ObservedTDEEKcal     float64  `json:"observed_tdee_kcal" example:"2425"`
	FormulaTDEEKcal      *float64 `json:"formula_tdee_kcal,omitempty" example:"2500"`
	RecommendedTDEEKcal  float64  `json:"recommended_tdee_kcal" example:"2460"`
	DataQualityScore     float64  `json:"data_quality_score" example:"0.78"`
}
