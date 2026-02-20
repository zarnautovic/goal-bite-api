package dto

import (
	"encoding/json"
	"time"

	"nutrition/internal/service"
)

type UpdateMeRequest struct {
	Name          *string  `json:"name,omitempty" example:"John Doe"`
	Sex           *string  `json:"sex,omitempty" example:"male"`
	BirthDate     *string  `json:"birth_date,omitempty" example:"1994-05-18"`
	HeightCM      *float64 `json:"height_cm,omitempty" example:"178"`
	ActivityLevel *string  `json:"activity_level,omitempty" example:"moderate"`

	nameSet          bool
	sexSet           bool
	birthDateSet     bool
	heightCMSet      bool
	activityLevelSet bool
}

func (r *UpdateMeRequest) UnmarshalJSON(data []byte) error {
	type alias UpdateMeRequest
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*r = UpdateMeRequest{}

	if v, ok := raw["name"]; ok {
		r.nameSet = true
		if string(v) != "null" {
			var parsed string
			if err := json.Unmarshal(v, &parsed); err != nil {
				return err
			}
			r.Name = &parsed
		}
	}
	if v, ok := raw["sex"]; ok {
		r.sexSet = true
		if string(v) != "null" {
			var parsed string
			if err := json.Unmarshal(v, &parsed); err != nil {
				return err
			}
			r.Sex = &parsed
		}
	}
	if v, ok := raw["birth_date"]; ok {
		r.birthDateSet = true
		if string(v) != "null" {
			var parsed string
			if err := json.Unmarshal(v, &parsed); err != nil {
				return err
			}
			r.BirthDate = &parsed
		}
	}
	if v, ok := raw["height_cm"]; ok {
		r.heightCMSet = true
		if string(v) != "null" {
			var parsed float64
			if err := json.Unmarshal(v, &parsed); err != nil {
				return err
			}
			r.HeightCM = &parsed
		}
	}
	if v, ok := raw["activity_level"]; ok {
		r.activityLevelSet = true
		if string(v) != "null" {
			var parsed string
			if err := json.Unmarshal(v, &parsed); err != nil {
				return err
			}
			r.ActivityLevel = &parsed
		}
	}
	return nil
}

func (r *UpdateMeRequest) Validate() error {
	if !r.nameSet && !r.sexSet && !r.birthDateSet && !r.heightCMSet && !r.activityLevelSet {
		return service.ErrNoFieldsToUpdate
	}
	if r.nameSet && r.Name == nil {
		return ErrInvalidName
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
			return ErrInvalidDate
		}
	}
	if r.HeightCM != nil && *r.HeightCM <= 0 {
		return ErrInvalidWeightKG
	}
	if r.ActivityLevel != nil {
		switch *r.ActivityLevel {
		case "sedentary", "light", "moderate", "active", "very_active":
		default:
			return ErrInvalidActivityLevel
		}
	}
	return nil
}

func (r *UpdateMeRequest) ToServiceInput() (service.UpdateUserInput, error) {
	out := service.UpdateUserInput{
		Name:             r.Name,
		SexSet:           r.sexSet,
		Sex:              r.Sex,
		HeightCMSet:      r.heightCMSet,
		HeightCM:         r.HeightCM,
		ActivityLevelSet: r.activityLevelSet,
		ActivityLevel:    r.ActivityLevel,
	}
	if r.birthDateSet {
		out.BirthDateSet = true
	}
	if r.BirthDate != nil {
		v, err := time.Parse("2006-01-02", *r.BirthDate)
		if err != nil {
			return service.UpdateUserInput{}, err
		}
		v = v.UTC()
		out.BirthDate = &v
	}
	return out, nil
}
