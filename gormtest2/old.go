package main

import (
  "gorm.io/gorm"
  "fmt"
  "github.com/jinzhu/gorm/dialects/sqlite"
)

type User struct {
  gorm.Model
  Languages []Language `gorm:"many2many:user_languages;"`
}

type Language struct {
  gorm.Model
  Name string
}

func main() {
  fmt.Println("Hello, world.")

  //db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{
  db, err = gorm.Open("sqlite", "database.db", &gorm.Config{
  	DisableForeignKeyConstraintWhenMigrating: true,
  })

  db.AutoMigrate(&User{}, &Language)
}
