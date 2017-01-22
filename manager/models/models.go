package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Server represents a row in table `Servers`
type Server struct {
	gorm.Model
	Hostname  string `gorm:"priamry_key;size:255;not null"`
	PublicIP  string `gorm:"type:char(15);unique;not null"`
	PrivateIP string `gorm:"type:char(15);unique;not null"`
	SlavePort uint32 `gorm:"not null"`
	Bandwidth int64  `gorm:"not null"`
	Transfer  int64  `gorm:"not null"`
	Provider  string `gorm:"not null"`
	Extra     string `gorm:"type:varchar(4095)"`
}

// User represents a row in table `Users`
type User struct {
	gorm.Model
	UserId   string `gorm:"primary_key;size:255;not null"`
	Alias    string `gorm:"size:255"`
	Phone    string `gorm:"size:255;unique;not null"`
	Email    string `gorm:"size:255;unique;not null"`
	Password string `gorm:"size:255;not null"`
}

// Service represents a row in table `Services`
type Service struct {
	gorm.Model
	Hostname   string    `gorm:"size:255;not null"`
	Server     Server    `gorm:"foreign_key:Hostname"`
	Port       uint32    `gorm:"not null"`
	Traffic    int64     `gorm:"not null"`
	CreateTime time.Time `gorm:"not null"`
	Status     string    `gorm:"size:255;not null"`
	User       User      `gorm:"foreign_key:UserId"`
	UserId     string    `gorm:"size:255;not null"`
}

// Product represents a row in table `Products`
type Product struct {
	gorm.Model
	ProductId   string `gorm:"primary_key;size:255;not null"`
	Price       uint   `gorm:"not null"`
	Description string `gorm:"type:varchar(1023);not null"`
	Status      string `gorm:"size:255;not null"`
	Extra       string `gorm:"type:varchar(4095)"`
}

// Order represents a row in table `Orders`
type Order struct {
	gorm.Model
	OrderId    string    `gorm:"primary_key;size:255;not null"`
	Channel    string    `gorm:"size:255;not null"`
	UserId     string    `gorm:"size:255;not null"`
	User       User      `gorm:"foreign_key:UserId"`
	CreateTime time.Time `gorm:"not null"`
	Amount     uint      `gorm:"not null"`
	ProductId  string    `gorm:"size:255;not null"`
	Product    Product   `gorm:"foreign_key:ProductId"`
}
