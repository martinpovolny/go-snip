package main

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"time"
)

type Model struct {
	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type Service struct {
	Model
	Owner            string
	Name             string           `gorm:"not null; index"`
	PeerDependencies []PeerDependency `gorm:"many2many:peer_memberships"`
}

type PeerDependency struct {
	Model
	Name     string
	Owner    string
	Services []Service `gorm:"many2many:peer_memberships"`
}

func main() {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&Service{}, &PeerDependency{})

	pd1 := PeerDependency{
		Name:     "databases",
		Owner:    "mpovolny@redhat.com",
		Services: []Service{{Name: "mysql", Owner: "mpovolny@redhat.com"}, {Name: "postgres", Owner: "mpovolny@redhat.com"}},
	}

	db.Create(&pd1)

	var pd PeerDependency
	db.Debug().Preload("Services").First(&pd)
	fmt.Printf("PD:\t %v\n\n", pd)
	fmt.Printf("SERVICES:\t %v\n\n", pd.Services)

	var service Service
	db.Preload("PeerDependencies").First(&service)
	fmt.Printf("SERVICE 1:\t %v\n\n", service)

	/* This works: 
	var pd2 PeerDependency
	db.First(&pd2)
	fmt.Printf("PD2:\t %v\n\n", pd2)
	s3 := Service{Name: "mssql", Owner: "foobar@redhat.com", PeerDependencies: []PeerDependency{pd2}}
	fmt.Printf("SERVICE 3:\t %v\n\n", s3)
	db.Create(&s3)
	*/

	/* This works too: */
	s3 := Service{Name: "mssql", Owner: "foobar@redhat.com", PeerDependencies: []PeerDependency{pd}}
	db.Create(&s3)
	fmt.Printf("SERVICE 3:\t %v\n\n", s3)

	db.Preload("PeerDependencies").First(&s3)
	fmt.Printf("SERVICE 3:\t %v\n\n", s3)


	/* Adding a service */
	var pd3 PeerDependency
	db.Preload("Services").First(&pd3)
	fmt.Printf("PD (pre): \t %v\n\n", pd3)
	pd3.Services = append(pd3.Services, Service{Name: "Oracle", Owner: "foo@bar.com"})
	db.Save(&pd3)

	db.Preload("Services").First(&pd3)
	fmt.Printf("PD (post): \t %v\n\n", pd3)
}
