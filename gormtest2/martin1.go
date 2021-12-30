// main.go

package main

import (
    "time"
    "gorm.io/gorm"
    "gorm.io/driver/sqlite"
)

type Model struct {
     ID uint `gorm:"primaryKey"`
     CreatedAt time.Time `json:"created_at,omitempty"`
     UpdatedAt time.Time `json:"updated_at,omitempty"`
    // DeletedAt gorm.DeletedAt `json:"created_at,omitempty"`
}

type Service struct {
      Model
      Owner   string
      //id string
      Name string `gorm:"not null; index"`
      PeerDependencies []PeerDependency `gorm:"many2many:peer_memberships"` 
}

type PeerDependency struct {
     Model
     Name string
     Owner   string
    // id string
    Services []Service `gorm:"many2many:peer_memberships"`
}

type PeerMembership struct {
     Model
     PeerDependencyId uint
     ServiceId uint
    // id string  `peer_dependency_id`,`service_id`

}

//var db *gorm.DB
//var err error
func main() {
  db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
  // db, err = gorm.Open(" postgres", "host=db port=5432 user=status-board-user dbname=status-board sslmode=disable")
  if err != nil {
    panic("failed to connect database")
  }
// db.DropTable(&Service{})
//db.DropTable(&PeerMembership{}, &PeerDependency{}, &Service{})
  // db.AutoMigrate(&Service{}, &PeerDependency{}, &PeerMemberships{})
//  db.Droptable(&PeerMembership)
//All foreign keys need to define here
//db.Model(PeerMembership{}).AddForeignKey("service_id", "services(id)", "CASCADE", "CASCADE")
//db.Model(PeerMembership{}).AddForeignKey("peer_dependency_id", "peer_dependencies(id)", "CASCADE", "CASCADE")
   db.AutoMigrate(&Service{})
   db.AutoMigrate(&PeerDependency{})


   db.Create(&Service{Name: "fred", Owner: "wilma"})
   db.Create(&Service{Name: "fred2", Owner: "wilma2"})
   db.Create(&PeerDependency{Name: "barney", Owner: "betty"})

    db.AutoMigrate(&Service{})
   db.AutoMigrate(&PeerDependency{}) 
  //db.AutoMigrate(&Service{})

// var services []Service
peer_dependencies := PeerDependency{}

//db.First(&PeerDependency, "id = ?", 111)
db.Preload("Services").Find(&peer_dependencies)
// db.Model(&peer_dependency).Related(&services, "Service")
//// SELECT * FROM "users" INNER JOIN "user_languages" ON "user_languages"."user_id" = "users"."id" WHERE ("user_languages"."language_id" IN ('111'))
  db.AutoMigrate(&Service{}, &PeerDependency{})

}


// go build main.go

// see go.mod