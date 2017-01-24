package persistence

import "time"

// Server represents a row in table `servers`
type Server struct {
	Hostname  string    `gorm:"priamry_key;size:255;not null"`
	PublicIP  string    `gorm:"type:char(15);unique;not null"`
	PrivateIP string    `gorm:"type:char(15);unique;not null"`
	SlavePort uint32    `gorm:"not null"`
	Bandwidth int64     `gorm:"not null"`
	Transfer  int64     `gorm:"not null"`
	Provider  string    `gorm:"not null"`
	Extra     string    `gorm:"type:varchar(4095)"`
	Services  []Service `gorm:"ForeignKey:Hostname"`
}

// TableName sets Server's table name to be `servers`
func (Server) TableName() string {
	return "servers"
}

// User represents a row in table `users`
type User struct {
	ID        string    `gorm:"primary_key;size:255;not null"`
	Role      string    `gorm:"-"`
	Alias     string    `gorm:"size:255"`
	Phone     string    `gorm:"size:255;unique;not null"`
	Email     string    `gorm:"size:255;unique;not null"`
	Password  string    `gorm:"size:255;not null"`
	CreatedAt time.Time `gorm:"not null"`
	Services  []Service `gorm:"ForeignKey:UserId;AssociationForeignKey:ID"`
}

// TableName sets User's table name to be `users` for normal users and
// "admin_users" for administrators.
func (u User) TableName() string {
	if u.Role == "admin" {
		return "admin_users"
	} else {
		return "users"
	}
}

// Service represents a row in table `services`
type Service struct {
	Hostname  string    `gorm:"size:255;not null"`
	Port      uint32    `gorm:"not null"`
	Traffic   int64     `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
	Status    string    `gorm:"size:255;not null"`
	UserId    string    `gorm:"size:255;not null"`
}

// TableName sets Service's table name to be `services`
func (Service) TableName() string {
	return "services"
}

// Product represents a row in table `products`
type Product struct {
	ID          string    `gorm:"primary_key;size:255;not null"`
	Price       uint      `gorm:"not null"`
	Description string    `gorm:"type:varchar(1023);not null"`
	Status      string    `gorm:"size:255;not null"`
	CreatedAt   time.Time `gorm:"not null"`
	Extra       string    `gorm:"type:varchar(4095)"`
	Orders      []Order   `gorm:"ForeignKey:ProductId;AssociationForeignKey:ID"`
}

// TableName sets Product's table name to be `products`
func (Product) TableName() string {
	return "products"
}

// Order represents a row in table `orders`
type Order struct {
	ID        string    `gorm:"primary_key;size:255;not null"`
	Channel   string    `gorm:"size:255;not null"`
	UserId    string    `gorm:"size:255;not null"`
	User      User      `gorm:"ForeignKey:UserId"`
	CreatedAt time.Time `gorm:"not null"`
	Amount    uint      `gorm:"not null"`
	ProductId string    `gorm:"size:255;not null"`
}

// TableName sets Order's table name to be `orders`
func (Order) TableName() string {
	return "orders"
}
