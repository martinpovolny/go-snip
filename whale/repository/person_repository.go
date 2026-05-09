package repository

import (
	"whale/models"

	"gorm.io/gorm"
)

type PersonRepository interface {
	Save(p *models.Person) error
	GetByExternalID(id string) (*models.Person, error)
}
type PersonDBRepository struct {
	db *gorm.DB
}

func NewPersonDBRepository(db *gorm.DB) *PersonDBRepository {
	return &PersonDBRepository{db: db}
}

func (r *PersonDBRepository) GetByExternalID(externalID string) (*models.Person, error) {
	var person models.Person
	result := r.db.Where("external_id = ?", externalID).First(&person)
	if result.Error != nil {
		return nil, result.Error
	}
	return &person, nil
}

func (r *PersonDBRepository) Save(person *models.Person) error {
	return r.db.Create(person).Error
}
