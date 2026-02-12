package main

import (
	"flag"
	"fmt"
	"go-noah/internal/model"
	"go-noah/internal/repository"
	"go-noah/pkg/config"
	"go-noah/pkg/log"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	var envConf = flag.String("conf", "config/local.yml", "config path, eg: -conf ./config/local.yml")
	var user = flag.String("user", "admin", "username to reset")
	var password = flag.String("password", "123456", "new password")
	flag.Parse()

	conf := config.NewConfig(*envConf)
	logger := log.NewLog(conf)

	db := repository.NewDB(conf, logger)

	hash, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	res := db.Model(&model.AdminUser{}).Where("username = ?", *user).Update("password", string(hash))
	if res.Error != nil {
		panic(res.Error)
	}
	fmt.Printf("reset password for user '%s' affected=%d\n", *user, res.RowsAffected)
}
