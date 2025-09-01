package config

import (
	"fmt"
	"jogodecartasonline/server/game/models"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)



var user = os.Getenv("DB_USER")
var password = os.Getenv("DB_PASSWORD")
var dbname = os.Getenv("DB_NAME")
var host = os.Getenv("DB_HOST")
var port = os.Getenv("DB_PORT")

var dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
    user, password, host, port, dbname)



func CretaeTable() gorm.DB{
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Falha ao conectar no banco de dados")
	}

	// Cria as tabelas automaticamente
	err = db.AutoMigrate(&models.Player{}, &models.Card{})
	if err != nil {
		panic("Falha ao criar tabelas")
	}
	fmt.Print("mIGRAÇÃO FEITA")
	return  *db
}



