package models

import (
	"github.com/jinzhu/gorm"
)

type Server struct {
	gorm.Model
	Url   string `gorm:"size:255"`
	Token string `gorm:"size:255"`
}

type Service struct {
	gorm.Model

}
