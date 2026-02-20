package repository

import (
	"context"
	"errors"
	"time"

	"goal-bite-api/internal/domain/user"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("record not found")

type UserRepository struct {
	db *gorm.DB
}

type UserUpdate struct {
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

func NewUserRepository(database *gorm.DB) *UserRepository {
	return &UserRepository{db: database}
}

func (r *UserRepository) GetByID(ctx context.Context, id uint) (user.User, error) {
	var u user.User
	err := r.db.WithContext(ctx).First(&u, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return user.User{}, ErrNotFound
	}
	if err != nil {
		return user.User{}, err
	}

	return u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (user.User, error) {
	var u user.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return user.User{}, ErrNotFound
	}
	if err != nil {
		return user.User{}, err
	}
	return u, nil
}

func (r *UserRepository) Create(ctx context.Context, value user.User) (user.User, error) {
	if err := r.db.WithContext(ctx).Create(&value).Error; err != nil {
		return user.User{}, err
	}
	return value, nil
}

func (r *UserRepository) CreateWithSession(ctx context.Context, value user.User, session CreateAuthSessionInput) (user.User, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&value).Error; err != nil {
			return err
		}
		record := AuthSession{
			UserID:    value.ID,
			TokenHash: session.TokenHash,
			ExpiresAt: session.ExpiresAt.UTC(),
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return user.User{}, err
	}
	return value, nil
}

func (r *UserRepository) Update(ctx context.Context, id uint, updates UserUpdate) (user.User, error) {
	values := map[string]any{}
	if updates.Name != nil {
		values["name"] = *updates.Name
	}
	if updates.SexSet {
		if updates.Sex != nil {
			values["sex"] = *updates.Sex
		} else {
			values["sex"] = nil
		}
	}
	if updates.BirthDateSet {
		if updates.BirthDate != nil {
			values["birth_date"] = *updates.BirthDate
		} else {
			values["birth_date"] = nil
		}
	}
	if updates.HeightCMSet {
		if updates.HeightCM != nil {
			values["height_cm"] = *updates.HeightCM
		} else {
			values["height_cm"] = nil
		}
	}
	if updates.ActivityLevelSet {
		if updates.ActivityLevel != nil {
			values["activity_level"] = *updates.ActivityLevel
		} else {
			values["activity_level"] = nil
		}
	}

	result := r.db.WithContext(ctx).Model(&user.User{}).Where("id = ?", id).Updates(values)
	if result.Error != nil {
		return user.User{}, result.Error
	}
	if result.RowsAffected == 0 {
		return user.User{}, ErrNotFound
	}
	return r.GetByID(ctx, id)
}
