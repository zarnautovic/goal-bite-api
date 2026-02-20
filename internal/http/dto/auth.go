package dto

import (
	"errors"
	"strings"
	"time"

	"nutrition/internal/auth"
	"nutrition/internal/service"
)

var (
	ErrInvalidEmail        = errors.New("invalid email")
	ErrInvalidPassword     = errors.New("invalid password")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

type RegisterRequest struct {
	Name          string   `json:"name" example:"John Doe"`
	Email         string   `json:"email" example:"john@gmail.com"`
	Password      string   `json:"password" example:"Pass1234!"`
	Sex           *string  `json:"sex,omitempty" example:"male"`
	BirthDate     *string  `json:"birth_date,omitempty" example:"1994-05-18"`
	HeightCM      *float64 `json:"height_cm,omitempty" example:"178"`
	ActivityLevel *string  `json:"activity_level,omitempty" example:"moderate"`
}

func (r *RegisterRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return ErrInvalidName
	}
	if strings.TrimSpace(r.Email) == "" {
		return ErrInvalidEmail
	}
	if !auth.ValidatePasswordPolicy(r.Password) {
		return ErrInvalidPassword
	}
	if r.Sex != nil {
		switch *r.Sex {
		case "male", "female":
		default:
			return ErrInvalidName
		}
	}
	if r.BirthDate != nil {
		if _, err := time.Parse("2006-01-02", *r.BirthDate); err != nil {
			return ErrInvalidName
		}
	}
	if r.HeightCM != nil && *r.HeightCM <= 0 {
		return ErrInvalidName
	}
	if r.ActivityLevel != nil {
		switch *r.ActivityLevel {
		case "sedentary", "light", "moderate", "active", "very_active":
		default:
			return ErrInvalidName
		}
	}
	return nil
}

func (r *RegisterRequest) ToServiceInput() (service.RegisterInput, error) {
	var birthDate *time.Time
	if r.BirthDate != nil {
		v, err := time.Parse("2006-01-02", *r.BirthDate)
		if err != nil {
			return service.RegisterInput{}, err
		}
		v = v.UTC()
		birthDate = &v
	}
	return service.RegisterInput{
		Name:          r.Name,
		Email:         r.Email,
		Password:      r.Password,
		Sex:           r.Sex,
		BirthDate:     birthDate,
		HeightCM:      r.HeightCM,
		ActivityLevel: r.ActivityLevel,
	}, nil
}

type LoginRequest struct {
	Email    string `json:"email" example:"john@gmail.com"`
	Password string `json:"password" example:"Pass1234!"`
}

func (r *LoginRequest) Validate() error {
	if strings.TrimSpace(r.Email) == "" {
		return ErrInvalidEmail
	}
	if strings.TrimSpace(r.Password) == "" {
		return ErrInvalidPassword
	}
	return nil
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" example:"9f5v..."`
}

func (r *RefreshTokenRequest) Validate() error {
	if strings.TrimSpace(r.RefreshToken) == "" {
		return ErrInvalidRefreshToken
	}
	return nil
}
