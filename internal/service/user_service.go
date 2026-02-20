package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"nutrition/internal/domain/user"
	"nutrition/internal/repository"
)

var ErrUserNotFound = errors.New("user not found")
var ErrInvalidUserProfile = errors.New("invalid user profile")

type UserReader interface {
	GetByID(ctx context.Context, id uint) (user.User, error)
	Update(ctx context.Context, id uint, updates repository.UserUpdate) (user.User, error)
}

type UserService struct {
	repo UserReader
}

func NewUserService(repo UserReader) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetByID(ctx context.Context, id uint) (user.User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return user.User{}, ErrUserNotFound
	}
	if err != nil {
		return user.User{}, err
	}
	u.PasswordHash = ""

	return u, nil
}

type UpdateUserInput struct {
	Name             *string
	SexSet           bool
	Sex              *string
	BirthDateSet     bool
	BirthDate        *time.Time
	HeightCMSet      bool
	HeightCM         *float64
	ActivityLevelSet bool
	ActivityLevel    *string
}

func (s *UserService) Update(ctx context.Context, id uint, in UpdateUserInput) (user.User, error) {
	if in.Name == nil && !in.SexSet && !in.BirthDateSet && !in.HeightCMSet && !in.ActivityLevelSet {
		return user.User{}, ErrNoFieldsToUpdate
	}

	updates := repository.UserUpdate{}
	if in.Name != nil {
		trimmed := strings.TrimSpace(*in.Name)
		if trimmed == "" {
			return user.User{}, ErrInvalidUserProfile
		}
		updates.Name = &trimmed
	}
	if in.SexSet && in.Sex != nil {
		switch *in.Sex {
		case "male", "female":
		default:
			return user.User{}, ErrInvalidUserProfile
		}
	}
	if in.BirthDateSet && in.BirthDate != nil {
		v := in.BirthDate.UTC()
		in.BirthDate = &v
	}
	if in.HeightCMSet && in.HeightCM != nil {
		if *in.HeightCM <= 0 {
			return user.User{}, ErrInvalidUserProfile
		}
	}
	if in.ActivityLevelSet && in.ActivityLevel != nil {
		switch *in.ActivityLevel {
		case "sedentary", "light", "moderate", "active", "very_active":
		default:
			return user.User{}, ErrInvalidUserProfile
		}
	}
	updates.SexSet = in.SexSet
	updates.Sex = in.Sex
	updates.BirthDateSet = in.BirthDateSet
	updates.BirthDate = in.BirthDate
	updates.HeightCMSet = in.HeightCMSet
	updates.HeightCM = in.HeightCM
	updates.ActivityLevelSet = in.ActivityLevelSet
	updates.ActivityLevel = in.ActivityLevel

	u, err := s.repo.Update(ctx, id, updates)
	if errors.Is(err, repository.ErrNotFound) {
		return user.User{}, ErrUserNotFound
	}
	if err != nil {
		return user.User{}, err
	}
	u.PasswordHash = ""
	return u, nil
}
