package repository

import (
	"whale/models"
)

type PersonRepositoryMock struct {
	SaveCalled int
	GetCalled  int
}

func NewPersonRepositoryMock() *PersonRepositoryMock {
	return &PersonRepositoryMock{}
}

func (r *PersonRepositoryMock) GetByExternalID(externalID string) (*models.Person, error) {
	r.GetCalled++
	return &models.Person{}, nil
}

func (r *PersonRepositoryMock) Save(person *models.Person) error {
	r.SaveCalled++
	return nil
}
