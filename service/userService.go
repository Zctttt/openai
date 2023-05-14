package services

import (
	"errors"
	"fmt"
	"go-chatgpt-api/initialize/mysql"
	"go-chatgpt-api/models"
)

func CreateUser(user, pass string) error {
	var (
		num      int64
		err      error
		userInfo = models.NewUser()
	)
	userInfo.SetName(user)
	userInfo.SetPassword(pass)
	userInfo.InitTime()

	if err = mysql.GetMarketDb().Model(models.MtUser{}).Where("NAME = ?", user).Count(&num).Error; err != nil {
		return err
	}
	if num != 0 {
		return errors.New("当前用户已存在")
	}
	if err = mysql.GetMarketDb().Create(userInfo).Error; err != nil {
		return err
	}
	return nil
}

func FindUser(user, pass string) (*models.MtUser, error) {
	var (
		err      error
		num      int64
		userInfo = models.NewUser()
	)

	if err = mysql.GetMarketDb().Model(models.MtUser{}).Where("NAME = ?", user).Count(&num).Error; err != nil {
		fmt.Println(err)
		return nil, err
	}
	if num == 0 {
		return nil, errors.New("用户不存在")
	}

	if err = mysql.GetMarketDb().Model(models.MtUser{}).Where("NAME = ?", user).Find(userInfo).Error; err != nil {
		fmt.Println(err)
		return nil, err
	}

	if !userInfo.ValidatePassword(pass) {
		return nil, errors.New("密码错误")
	}
	return userInfo, nil
}
