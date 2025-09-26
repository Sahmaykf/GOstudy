package main

import (
	"errors"
	"strings"

	"github.com/Sahmaykf/GOstudy/serverdir/data"
	"github.com/Sahmaykf/GOstudy/serverdir/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	Name          string
	Addr          string
	C             chan string
	conn          Conn
	server        *Server
	done          chan struct{}
	AccountID     *uint
	Authenticated bool //登录状态
}

func NewUser(conn Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		done:   make(chan struct{}),
		conn:   conn,
		server: server,
	}
	go user.ListenMessage()
	return user
}

func (now *User) Online() {

	now.server.maplock.Lock()
	now.server.OnlineMap[now.Name] = now
	now.server.maplock.Unlock()
	now.server.BroadCast(now, "上线")
}

func (now *User) Offline() {
	now.server.maplock.Lock()
	delete(now.server.OnlineMap, now.Name)
	now.server.maplock.Unlock()
	now.server.BroadCast(now, "下线")
	close(now.done)
	close(now.C)
	now.conn.Close()
}

func (now *User) sendMsg(msg string) {
	now.conn.Write([]byte(msg))
}

func (now *User) changeName(newName string) {
	now.server.maplock.Lock()
	delete(now.server.OnlineMap, now.Name)
	now.Name = newName
	now.server.OnlineMap[now.Name] = now
	now.server.maplock.Unlock()
	now.sendMsg("您的用户名更新为：" + now.Name + "\n")
}
func RegisterAccount(email, username, password string) error {
	email = strings.TrimSpace(email)
	username = strings.TrimSpace(username)
	if email == "" || username == "" || strings.TrimSpace(password) == "" {
		return errors.New("邮箱/用户名/密码不能为空")
	}
	var cnt int64
	err := data.DB.Model(&model.Account{}).
		Where("email = ? OR username = ?", email, username).
		Count(&cnt).Error
	if err != nil {
		return err
	}
	if cnt > 0 {
		return errors.New("邮箱或用户名已存在")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("生成密码哈希失败")
	}
	acc := &model.Account{
		Email:    email,
		Username: username,
		Password: string(hash),
	}
	return data.DB.Create(acc).Error
}
func Authenticate(username, password string) (*model.Account, error) {
	var acc model.Account
	if err := data.DB.Where("username = ?", username).First(&acc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("账号不存在")
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(acc.Password), []byte(password)); err != nil {
		return nil, errors.New("密码错误")
	}
	return &acc, nil
}
func (now *User) DoMessage(msg string) {
	if strings.HasPrefix(msg, "register|") {
		// register|email|username|password
		parts := strings.SplitN(msg, "|", 4)
		if len(parts) != 4 {
			now.sendMsg("注册格式错误:register|email|username|password\n")
			return
		}
		email := parts[1]
		username := parts[2]
		password := parts[3]

		if err := RegisterAccount(email, username, password); err != nil {
			now.sendMsg("注册失败: " + err.Error() + "\n")
			return
		}
		now.sendMsg("注册成功，请使用 login|username|password 登录\n")
		return
	}
	if strings.HasPrefix(msg, "login|") {
		// login|username|password
		parts := strings.SplitN(msg, "|", 3)
		if len(parts) != 3 {
			now.sendMsg("登录格式错误:login|username|password\n")
			return
		}
		username := parts[1]
		password := parts[2]
		acc, err := Authenticate(username, password)
		if err != nil {
			now.sendMsg("登录失败: " + err.Error() + "\n")
			return
		}
		var victim *User
		now.server.maplock.Lock()
		u, ok := now.server.OnlineMap[acc.Username]
		if ok && u != now { //移除旧会话
			victim = u
			delete(now.server.OnlineMap, acc.Username)
		}
		existing, ok := now.server.OnlineMap[now.Name]
		if ok && existing == now { //移除当前可能存在的临时会话
			delete(now.server.OnlineMap, now.Name)
		}
		// 绑定新身份
		now.Name = acc.Username
		now.AccountID = &acc.ID
		now.Authenticated = true
		now.server.OnlineMap[now.Name] = now
		now.server.maplock.Unlock()
		// 在锁外踢掉旧会话，避免死锁
		if victim != nil {
			victim.sendMsg("你的账号已在另一处登录，当前连接被下线\n")
			victim.Offline()
			return
		}

		now.sendMsg("登录成功，当前用户名: " + now.Name + "\n")
		return
	}

	if msg == "who" {
		now.server.maplock.Lock()
		for _, user := range now.server.OnlineMap {
			now.sendMsg("[" + user.Addr + "]" + user.Name + ":" + "在线\n")
		}
		now.server.maplock.Unlock()
		return
	}
	if strings.HasPrefix(msg, "rename|") {
		newName := strings.Split(msg, "|")[1]
		_, ok := now.server.OnlineMap[newName]
		if ok {
			now.sendMsg("用户名重复了\n")
		} else {
			now.changeName(newName)
		}
		return
	}
	if strings.HasPrefix(msg, "to|") {
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			now.sendMsg("格式不正确,正确的格式为to|张三|message...\n")
			return
		}
		remoteUser, ok := now.server.OnlineMap[remoteName]

		if !ok {
			now.sendMsg("查无此人\n")
			return
		}
		content := strings.Split(msg, "|")[2]
		if content == "" {
			now.sendMsg("消息为空,请重试\n")
			return
		}
		remoteUser.sendMsg(now.Name + " say: " + content + "\n")
		return
	}
	now.server.BroadCast(now, msg)

}

func (now *User) ListenMessage() {
	for {
		select {
		case msg, ok := <-now.C:
			if !ok {
				return
			}
			now.conn.Write([]byte(msg + "\n"))
		case <-now.done:
			return
		}
	}
}
