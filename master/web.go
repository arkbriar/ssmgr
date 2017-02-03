package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/asaskevich/govalidator"
	"github.com/kataras/go-mailer"
	"github.com/kataras/iris"

	"github.com/arkbriar/ss-mgr/master/orm"
)

const verifyCodeExpire = 300

var mail mailer.Service

func NewApp() *iris.Framework {

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

	app.Get("/*path", func(ctx *iris.Context) {
		path := ctx.Param("path")
		switch {
		case strings.HasPrefix(path, "/libs"):
			ctx.ServeFile("./public"+path, true)
		case strings.HasPrefix(path, "/public"):
			ctx.ServeFile("."+path, true)
		default:
			ctx.ServeFile("./views/index.html", true)
		}
	})

	return app
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

	// TODO: Prevent send to one email for too many times (saying, 10 mail per day)

	vcode := fmt.Sprintf("%06d", rand.Int31n(1000000))
	iris.Logger.Printf("Send verify code to %s: %s", request.Email, vcode)

	content := fmt.Sprintf("Your verify code is %s.\n", vcode)
	go func() {
		err := mail.Send("Free Shadowsocks", content, request.Email)
		if err != nil {
			iris.Logger.Print("Failed to send email: ", err.Error())
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
	db.Where("email = ?", request.Email).First(&user)

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

	if ctx.Session().GetString("user_id") != request.UserID {
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

	var (
		hosts     = make([]string, 0, len(allocs))
		ports     = make([]string, 0, len(allocs))
		passwords = make([]string, 0, len(allocs))
	)
	for _, alloc := range allocs {
		slave := config.Slaves[alloc.ServerID]
		if slave == nil {
			logrus.Warnf("Server '%s' does not exist", alloc.ServerID)
			continue
		}

		hosts = append(hosts, slave.Host)
		ports = append(ports, strconv.Itoa(alloc.Port))
		passwords = append(passwords, alloc.Password)
	}

	type response struct {
		Address     string `json:"address"`
		Email       string `json:"email"`
		Flow        int64  `json:"flow"`
		CurrentFlow int64  `json:"currentFlow"`
		Time        int64  `json:"time"`
		Expired     int64  `json:"expired"`
		Disabled    bool   `json:"isDisabled"`
		Host        string `json:"host"`
		Port        string `json:"port"`
		Password    string `json:"password"`
		Method      string `json:"method"`
	}
	ctx.JSON(iris.StatusOK, &response{
		Address:     request.UserID,
		Email:       user.Email,
		Flow:        user.QuotaFlow,
		CurrentFlow: flowSum[0].Flow,
		Time:        user.Time * 1000,
		Expired:     user.Expired * 1000,
		Disabled:    user.Disabled,
		Host:        strings.Join(hosts, ", "),
		Port:        strings.Join(ports, ", "),
		Password:    strings.Join(passwords, ", "),
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

func handleConfig(ctx *iris.Context) {
	if admin, err := ctx.Session().GetBoolean("is_admin"); err != nil || !admin {
		ctx.SetStatusCode(iris.StatusUnauthorized)
		ctx.WriteString("please login first")
		return
	}

	var ret systemConfig
	ret.Shadowsocks.Flow = config.Limit.Flow
	ret.Shadowsocks.Time = config.Limit.Time
	ctx.JSON(iris.StatusOK, &ret)
}

func handleConfigPut(ctx *iris.Context) {
	if admin, err := ctx.Session().GetBoolean("is_admin"); err != nil || !admin {
		ctx.SetStatusCode(iris.StatusUnauthorized)
		ctx.WriteString("please login first")
		return
	}

	var req systemConfig
	if err := ctx.ReadJSON(&req); err != nil {
		panic(err.Error())
	}

	config.Limit.Flow = req.Shadowsocks.Flow
	config.Limit.Time = req.Shadowsocks.Time

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
