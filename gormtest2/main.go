package main

import (
  "gorm.io/gorm"
  "gorm.io/driver/sqlite"
  "time"
)

type Product struct {
  gorm.Model
  Code  string
  Price uint
}

type Meta struct {
	ID        string  `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Model struct {
	ID        string `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Application struct {
	Model
	Name string `gorm:"not null; index"`
}

type PeerDependencies {
	Model
	Services []Service `gorm:"many2many:peer_membership"
}

type ApplicationGroup struct {
	Model
	Name     string `gorm:"not null"`
	Metadata string
	Applications []Application `gorm:"many2many:application_group_memberships";`
	//Applications []Application `gorm:many2many:application_group_memberships;foreignKey:application_group_id;joinForeignKey:application_id`
}

//type ApplicationGroupMembership struct {
//	ApplicationID      string `gorm:"not null"`
//	ApplicationGroupID string `gorm:"not null"`
//}

type Language struct {
  Model // gorm.Model
  Name string
}

type User struct {
  Model //gorm.Model
  Languages []Language `gorm:"many2many:user_languages;"`
}

func main() {
  db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
  if err != nil {
    panic("failed to connect database")
  }

  // Migrate the schema
  db.AutoMigrate(&Product{})

  // Create
  db.Create(&Product{Code: "D42", Price: 100})

  // Read
  var product Product
  db.First(&product, 1) // find product with integer primary key
  db.First(&product, "code = ?", "D42") // find product with code D42

  // Update - update product's price to 200
  db.Model(&product).Update("Price", 200)
  // Update - update multiple fields
  db.Model(&product).Updates(Product{Price: 200, Code: "F42"}) // non-zero fields
  db.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})

  // Delete - delete product
  db.Delete(&product, 1)

  db.AutoMigrate(&User{}, &Language{})
  db.AutoMigrate(&Application{}, &ApplicationGroup{})
}
