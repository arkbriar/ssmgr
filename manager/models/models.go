package models

import "time"

// Server represents a row in table `Servers`
type Server struct {
	Hostname  string    `gorm:"priamry_key;size:255;not null"`
	PublicIP  string    `gorm:"type:char(15);unique;not null"`
	PrivateIP string    `gorm:"type:char(15);unique;not null"`
	SlavePort uint32    `gorm:"not null"`
	Bandwidth int64     `gorm:"not null"`
	Transfer  int64     `gorm:"not null"`
	Provider  string    `gorm:"not null"`
	Extra     string    `gorm:"type:varchar(4095)"`
	Services  []Service `gorm:"foreign_key:Hostname"`
}

// User represents a row in table `Users`
type User struct {
	UserId   string    `gorm:"primary_key;size:255;not null"`
	Alias    string    `gorm:"size:255"`
	Phone    string    `gorm:"size:255;unique;not null"`
	Email    string    `gorm:"size:255;unique;not null"`
	Password string    `gorm:"size:255;not null"`
	Services []Service `gorm:"foreign_key:UserId"`
}

// Service represents a row in table `Services`
type Service struct {
	Hostname   string    `gorm:"size:255;not null"`
	Port       uint32    `gorm:"not null"`
	Traffic    int64     `gorm:"not null"`
	CreateTime time.Time `gorm:"not null"`
	Status     string    `gorm:"size:255;not null"`
	UserId     string    `gorm:"size:255;not null"`
}

// Product represents a row in table `Products`
type Product struct {
	ProductId   string  `gorm:"primary_key;size:255;not null"`
	Price       uint    `gorm:"not null"`
	Description string  `gorm:"type:varchar(1023);not null"`
	Status      string  `gorm:"size:255;not null"`
	Extra       string  `gorm:"type:varchar(4095)"`
	Orders      []Order `gorm:"foreign_key:ProductId"`
}

// Order represents a row in table `Orders`
type Order struct {
	OrderId    string    `gorm:"primary_key;size:255;not null"`
	Channel    string    `gorm:"size:255;not null"`
	UserId     string    `gorm:"size:255;not null"`
	User       User      `gorm:"foreign_key:UserId"`
	CreateTime time.Time `gorm:"not null"`
	Amount     uint      `gorm:"not null"`
	ProductId  string    `gorm:"size:255;not null"`
}
