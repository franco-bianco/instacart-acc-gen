package gen

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"

	"github.com/bxcodec/faker/v4"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

type Session struct {
	Cancel      context.CancelFunc
	Ctx         context.Context
	Log         *logrus.Entry
	Client      *resty.Client
	UserConfig  *Config
	ProxyList   []string
	CouponCodes []string

	state state
}

type state struct {
	Email       string
	Password    string
	FirstName   string
	PostalCode  string
	CouponValue string
	CapTaskID   string
	ReCapToken  string
}

func NewSession(log *logrus.Logger, ctx context.Context, cancel context.CancelFunc, userConfig *Config, taskID int) (*Session, error) {

	firstName := faker.FirstName()
	email := fmt.Sprintf("%s+%d@gmail.com", userConfig.GmailAddress, rand.Intn(8999)+1000)
	postalCode := fmt.Sprintf("%05d", rand.Intn(99950-501)+501)

	return &Session{
		Cancel:     cancel,
		Ctx:        ctx,
		Log:        log.WithField("task_id", taskID),
		UserConfig: userConfig,
		Client:     resty.New(),
		state: state{
			Email:      email,
			Password:   generatePassword(),
			FirstName:  firstName,
			PostalCode: postalCode,
		},
	}, nil
}

func (s *Session) setProxy() error {
	if len(s.ProxyList) == 0 {
		return nil
	}
	proxy := s.ProxyList[rand.Intn(len(s.ProxyList))]
	parts := strings.Split(proxy, ":")
	if len(parts) != 4 {
		return fmt.Errorf("proxy must be in the ip:port:user:pass format")
	}
	uri, _ := url.Parse(fmt.Sprintf("http://%s:%s@%s:%s", parts[2], parts[3], parts[0], parts[1]))
	s.Client.SetTransport(&http.Transport{
		Proxy: http.ProxyURL(uri),
	})
	s.Client.SetProxy(fmt.Sprintf("http://%s:%s@%s:%s", parts[2], parts[3], parts[0], parts[1]))
	return nil
}

func generatePassword() string {
	lowercase := "abcdefghijklmnopqrstuvwxyz"
	uppercase := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers := "0123456789"
	special := "!@$*?"
	password := make([]byte, 16)
	for i := 0; i < 16; i++ {
		if i < 4 {
			password[i] = lowercase[rand.Intn(len(lowercase))]
		} else if i < 8 {
			password[i] = uppercase[rand.Intn(len(uppercase))]
		} else if i < 12 {
			password[i] = numbers[rand.Intn(len(numbers))]
		} else {
			password[i] = special[rand.Intn(len(special))]
		}
	}
	rand.Shuffle(len(password), func(i, j int) {
		password[i], password[j] = password[j], password[i]
	})
	return string(password)
}
