package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/asaskevich/govalidator"
	"github.com/kataras/go-mailer"
	"github.com/kataras/iris"

	"github.com/arkbriar/ssmgr/master/orm"
)

const verifyCodeExpire = 300

var mail mailer.Service

func NewApp(webroot string) *iris.Framework {

	mail = mailer.New(mailer.Config{
		Host:      config.Email.Host,
		Port:      config.Email.Port,
		Username:  config.Email.Username,
		Password:  config.Email.Password,
		FromAddr:  config.Email.FromAddr,
		FromAlias: config.Email.FromAlias,
	})

	app := iris.New()

	app.Post("/email", handleEmail)
	app.Post("/code", handleCode)
	app.Post("/account", handleAccount)
	app.Post("/config", handleConfig)
	app.Post("/password", handlePassword)
	app.Put("/config", handleConfigPut)
	app.Post("/logout", handleLogout)
	app.Post("/user", handleUser)
	app.Post("/flow", handleFlow)
	app.Post("/group", handleGroup)
	app.Put("/user", handleUserPut)

	app.Get("/*path", func(ctx *iris.Context) {
		path := ctx.Param("path")
		switch {
		case strings.HasPrefix(path, "/libs"), strings.HasPrefix(path, "/public"):
			ctx.ServeFile(webroot+path, true)
		default:
			ctx.ServeFile(webroot+"/views/index.html", true)
		}
	})

	return app
}

func isAdmin(ctx *iris.Context) bool {
	admin, err := ctx.Session().GetBoolean("is_admin")
	return err == nil && admin
}

func handleEmail(ctx *iris.Context) {
	var request struct {
		Email string `json:"email",valid:"email"`
	}
	if err := ctx.ReadJSON(&request); err != nil {
		panic(err.Error())
	}

	if _, err := govalidator.ValidateStruct(&request); err != nil {
		ctx.WriteString(err.Error())
		return
	}

	// Prevent send to one email addr for too many times
	var sentCount int
	timeFrom := time.Now().Add(-verifyCodeExpire * time.Second).Unix()
	db.Model(&orm.VerifyCode{}).Where("email = ? AND time > ?", request.Email, timeFrom).Count(&sentCount)
	if sentCount >= 3 {
		ctx.SetStatusCode(iris.StatusForbidden)
		ctx.WriteString("sent too many times")
		return
	}

	vcode := fmt.Sprintf("%06d", rand.Int31n(1000000))
	logrus.Infof("Send verify code to %s: %s", request.Email, vcode)

	content := fmt.Sprintf("Your verify code is %s.\n", vcode)
	go func() {
		err := mail.Send("Free Shadowsocks", content, request.Email)
		if err != nil {
			logrus.Errorf("Failed to send email: %s", err)
		}
	}()

	db.Save(&orm.VerifyCode{
		Email: request.Email,
		Code:  vcode,
		Time:  time.Now().Unix(),
	})

	ctx.WriteString("success")
}

func handleCode(ctx *iris.Context) {
	var request struct {
		Email string `json:"email",valid:"email"`
		Code  string `json:"code",valid:"length(6|6)"`
	}
	if err := ctx.ReadJSON(&request); err != nil {
		panic(err.Error())
	}

	if _, err := govalidator.ValidateStruct(&request); err != nil {
		ctx.WriteString(err.Error())
		return
	}

	var records []orm.VerifyCode
	timeFrom := time.Now().Add(-verifyCodeExpire * time.Second).Unix()
	db.Where("email = ? AND code = ? AND time > ?", request.Email, request.Code, timeFrom).Find(&records)
	if len(records) == 0 {
		ctx.SetStatusCode(iris.StatusForbidden)
		ctx.WriteString("login failed")
		return
	}

	var user orm.User
	db.Where("email = ? AND disabled = 0", request.Email).First(&user)

	if user.Email == "" {
		// User is not created yet
		user = *CreateUser(request.Email)
	}

	ctx.Session().Set("user_id", user.ID)
	ctx.WriteString(user.ID)
}

func handleAccount(ctx *iris.Context) {
	var request struct {
		UserID string `json:"address",valid:"length(32|32)"`
	}
	if err := ctx.ReadJSON(&request); err != nil {
		panic(err.Error())
	}
	if _, err := govalidator.ValidateStruct(&request); err != nil {
		ctx.WriteString(err.Error())
		return
	}

	isLogin := ctx.Session().GetString("user_id") == request.UserID
	isAdmin, _ := ctx.Session().GetBoolean("is_admin")

	if !isLogin && !isAdmin {
		ctx.SetStatusCode(iris.StatusForbidden)
		ctx.WriteString("please login first")
		return
	}

	var (
		user    orm.User
		allocs  []orm.Allocation
		flowSum []struct{ Flow int64 }
	)
	db.Where("id = ?", request.UserID).First(&user)
	db.Where("user_id = ?", request.UserID).Find(&allocs)
	db.Raw("SELECT sum(flow) AS flow FROM flow_record WHERE user_id = ?", request.UserID).Scan(&flowSum)

	type serverInfo struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	servers := make([]*serverInfo, 0, len(allocs))

	for _, alloc := range allocs {
		slave := slaves[alloc.ServerID]
		if slave == nil {
			logrus.Warnf("Server '%s' does not exist", alloc.ServerID)
			continue
		}

		servers = append(servers, &serverInfo{
			Host:     slave.Config.Host,
			Port:     alloc.Port,
			Password: alloc.Password,
			Name:     slave.Config.Name,
		})
	}

	type response struct {
		Address     string        `json:"address"`
		Email       string        `json:"email"`
		Flow        int64         `json:"flow"`
		CurrentFlow int64         `json:"currentFlow"`
		Time        int64         `json:"time"`
		Expired     int64         `json:"expired"`
		Disabled    bool          `json:"isDisabled"`
		Servers     []*serverInfo `json:"servers"`
		Method      string        `json:"method"`
	}
	ctx.JSON(iris.StatusOK, &response{
		Address:     request.UserID,
		Email:       user.Email,
		Flow:        user.QuotaFlow,
		CurrentFlow: flowSum[0].Flow,
		Time:        user.Time * 1000, // convert to milliseconds
		Expired:     user.Expired * 1000,
		Disabled:    user.Disabled,
		Servers:     servers,
		Method:      "aes-256-cfb",
	})
}

func handlePassword(ctx *iris.Context) {
	var request struct {
		Password string `json:"password",valid:"-"`
	}
	if err := ctx.ReadJSON(&request); err != nil {
		panic(err.Error())
	}
	if _, err := govalidator.ValidateStruct(&request); err != nil {
		ctx.WriteString(err.Error())
		return
	}

	if request.Password == config.Password {
		ctx.Session().Set("is_admin", true)
		ctx.WriteString("success")
	} else {
		ctx.SetStatusCode(http.StatusForbidden)
		ctx.WriteString("login failed")
	}
}

type systemConfig struct {
	Shadowsocks struct {
		Flow int64 `json:"flow"`
		Time int64 `json:"time"`
	} `json:"shadowsocks"`
	Limit struct {
		User struct {
			Day   int64 `json:"day"`
			Week  int64 `json:"week"`
			Month int64 `json:"month"`
		} `json:"user"`
		Global struct {
			Day   int64 `json:"day"`
			Week  int64 `json:"week"`
			Month int64 `json:"month"`
		} `json:"global"`
	}
}

// TODO: Currently front-end does not support config for groups
func handleConfig(ctx *iris.Context) {
	if !isAdmin(ctx) {
		ctx.SetStatusCode(iris.StatusUnauthorized)
		ctx.WriteString("please login first")
		return
	}

	var ret systemConfig
	ret.Shadowsocks.Flow = defaultGroup.Config.Limit.Flow
	ret.Shadowsocks.Time = defaultGroup.Config.Limit.Time
	ctx.JSON(iris.StatusOK, &ret)
}

func handleConfigPut(ctx *iris.Context) {
	if !isAdmin(ctx) {
		ctx.SetStatusCode(iris.StatusUnauthorized)
		ctx.WriteString("please login first")
		return
	}

	var req systemConfig
	if err := ctx.ReadJSON(&req); err != nil {
		panic(err.Error())
	}

	defaultGroup.Config.Limit.Flow = req.Shadowsocks.Flow
	defaultGroup.Config.Limit.Time = req.Shadowsocks.Time

	// Save into config file
	go func() {
		data, _ := json.MarshalIndent(config, "", "  ")
		err := ioutil.WriteFile(*configPath, data, 0644)
		if err != nil {
			logrus.Warn("Failed to save config: %s", err.Error())
		}
	}()
}

func handleLogout(ctx *iris.Context) {
	ctx.Session().Clear()
	ctx.WriteString("success")
}

func handleUser(ctx *iris.Context) {
	if !isAdmin(ctx) {
		ctx.SetStatusCode(iris.StatusUnauthorized)
		ctx.WriteString("please login first")
		return
	}

	const SQL = `SELECT user_id, email, quota_flow, sum(flow) AS current_flow, time, expired, disabled
FROM users JOIN flow_record ON users.id = flow_record.user_id
GROUP BY user_id`
	rows, err := db.Raw(SQL).Rows()
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	type response struct {
		UserID      string `json:"address"`
		Email       string `json:"email"`
		Flow        int64  `json:"flow"`
		CurrentFlow int64  `json:"currentFlow"`
		Time        int64  `json:"time"`
		Expired     int64  `json:"expired"`
		Disabled    bool   `json:"isDisabled"`
	}

	users := make([]*response, 0)

	for rows.Next() {
		var u response
		rows.Scan(&u.UserID, &u.Email, &u.Flow, &u.CurrentFlow, &u.Time, &u.Expired, &u.Disabled)

		// convert to milliseconds
		u.Time *= 1000
		u.Expired *= 1000

		users = append(users, &u)
	}

	ctx.JSON(iris.StatusOK, users)
}

func handleFlow(ctx *iris.Context) {
	if !isAdmin(ctx) {
		ctx.SetStatusCode(iris.StatusUnauthorized)
		ctx.WriteString("please login first")
		return
	}

	const SQL = `SELECT sum(flow) AS total_flow FROM flow_record`

	var result struct {
		TotalFlow int64 `json:"flow"`
	}
	db.Raw(SQL).Scan(&result)

	ctx.JSON(iris.StatusOK, &result)
}

func handleGroup(ctx *iris.Context) {
	if !isAdmin(ctx) {
		ctx.SetStatusCode(iris.StatusUnauthorized)
		ctx.WriteString("please login first")
		return
	}

	ctx.JSON(iris.StatusOK, GetGroupIDs())
}

type userConfig struct {
	UserID  string `json:"user_id",valid:"length(32|32)"`
	GroupID string `json:"group_id"`
}

func handleUserPut(ctx *iris.Context) {
	if !isAdmin(ctx) {
		ctx.SetStatusCode(iris.StatusUnauthorized)
		ctx.WriteString("please login first")
		return
	}

	var conf userConfig
	if err := ctx.ReadJSON(&conf); err != nil {
		panic(err.Error())
	}
	if _, err := govalidator.ValidateStruct(&conf); err != nil {
		ctx.WriteString(err.Error())
		return
	}

	if !HasGroup(conf.GroupID) {
		ctx.WriteString("group " + conf.GroupID + " not found")
		return
	}

	if err := ChangeUserGroup(conf.UserID, conf.GroupID); err != nil {
		ctx.WriteString(err.Error())
		return
	}
}
