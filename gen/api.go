package gen

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

var (
	ErrInvalidCoupon = errors.New("invalid coupon")
)

type createAccountBody struct {
	OperationName string                  `json:"operationName"`
	Variables     createAccountVariables  `json:"variables"`
	Extensions    createAccountExtensions `json:"extensions"`
}

type createAccountVariables struct {
	Email         string        `json:"email"`
	Password      string        `json:"password"`
	FirstName     string        `json:"firstName"`
	UtmParameters utmParameters `json:"utmParameters"`
	PostalCode    string        `json:"postalCode"`
	Recaptcha     string        `json:"recaptcha"`
	StreetAddress string        `json:"streetAddress"`
}

type utmParameters struct {
	UtmSource string `json:"utmSource"`
}

type createAccountExtensions struct {
	PersistedQuery persistedQuery `json:"persistedQuery"`
}

type persistedQuery struct {
	Version    int    `json:"version"`
	Sha256Hash string `json:"sha256Hash"`
}

type createAccountRes struct {
	Data struct {
		CreateUser struct {
			Token    string `json:"token"`
			Typename string `json:"__typename"`
		} `json:"createUser"`
	} `json:"data"`
}

type applyCouponRes struct {
	PromotionCodeRedemption struct {
		Label    string      `json:"label"`
		TermURL  interface{} `json:"term_url"`
		SubLabel string      `json:"sub_label"`
	} `json:"promotion_code_redemption"`
}

func (s *Session) createAccount() error {
	s.Log.Info("creating account...")

	res, err := s.Client.R().
		SetHeaders(map[string]string{
			"sec-ch-ua":           "\"Google Chrome\";v=\"119",
			"accept":              "*/*",
			"content-type":        "application/json",
			"x-client-identifier": "web",
			"sec-ch-ua-mobile":    "?0",
			"user-agent":          "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
			"sec-ch-ua-platform":  "\"macOS\"",
			"origin":              "https://www.instacart.com",
			"sec-fetch-site":      "same-origin",
			"sec-fetch-mode":      "cors",
			"sec-fetch-dest":      "empty",
			"referer":             "https://www.instacart.com/",
			"accept-language":     "en-US,en;q=0.9",
		}).
		SetBody(createAccountBody{
			OperationName: "CreateUser",
			Variables: createAccountVariables{
				Email:     s.state.Email,
				Password:  s.state.Password,
				FirstName: s.state.FirstName,
				UtmParameters: utmParameters{
					UtmSource: "auth_v4_modal",
				},
				PostalCode:    s.state.PostalCode,
				Recaptcha:     s.state.ReCapToken,
				StreetAddress: "",
			},
			Extensions: createAccountExtensions{
				PersistedQuery: persistedQuery{
					Version:    1,
					Sha256Hash: "812b97e9e3007d2800cacb35a7492f94838308be300be8009d3d4532c0cd9b29",
				},
			},
		}).
		Post("https://www.instacart.com/graphql")
	if err != nil {
		return err
	}
	if res.StatusCode() != 200 {
		fmt.Println(res.String())
		return fmt.Errorf("invalid status code: %d", res.StatusCode())
	}

	var createAccountRes createAccountRes
	if err := json.Unmarshal(res.Body(), &createAccountRes); err != nil {
		return err
	}

	s.Client.SetCookie(&http.Cookie{
		Name:  "__Host-instacart_sid",
		Value: createAccountRes.Data.CreateUser.Token,
	})

	return nil
}

func (s *Session) applyCoupon(coupon string) (*applyCouponRes, error) {
	s.Log.Info("applying coupon...")

	res, err := s.Client.R().
		SetHeaders(map[string]string{
			"sec-ch-ua":           "\"Google Chrome\";v=\"119",
			"x-csrf-token":        "",
			"sec-ch-ua-mobile":    "?0",
			"user-agent":          "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
			"content-type":        "application/json",
			"accept":              "application/json",
			"x-client-identifier": "web",
			"x-requested-with":    "XMLHttpRequest",
			"sec-ch-ua-platform":  "\"macOS\"",
			"origin":              "https://www.instacart.com",
			"sec-fetch-site":      "same-origin",
			"sec-fetch-mode":      "cors",
			"sec-fetch-dest":      "empty",
			"referer":             "https://www.instacart.com/store/account/manage_promos",
			"accept-language":     "en-US,en;q=0.9",
		}).
		SetBody(fmt.Sprintf(`{"code":"%s"}`, coupon)).
		Post("https://www.instacart.com/v3/promotion_codes/redemptions")
	if err != nil {
		return nil, err
	}
	if strings.Contains(res.String(), "This promotion isn't available right now.") || strings.Contains(res.String(), "reached its redemption limit") {
		return nil, ErrInvalidCoupon
	}
	if res.StatusCode() != 200 {
		fmt.Println(res.String())
		return nil, fmt.Errorf("invalid status code: %d", res.StatusCode())
	}

	var applyCouponRes applyCouponRes
	if err := json.Unmarshal(res.Body(), &applyCouponRes); err != nil {
		return nil, err
	}

	return &applyCouponRes, nil
}

func (s *Session) applyValidCoupon() error {
	s.Log.Info("applying valid coupon...")
	for i := 0; i < len(s.CouponCodes); i++ {
		coupon := s.CouponCodes[i]
		res, err := s.applyCoupon(coupon)
		if err != nil {
			if err == ErrInvalidCoupon {
				continue
			}
			return err
		}
		if strings.Contains(res.PromotionCodeRedemption.SubLabel, "You got $10 off") {
			s.Log.Infof("CODE %s only gave $10 off, trying next code", coupon)
			continue
		}

		// get coupon value
		re := regexp.MustCompile(`\$(\d+) off`)
		matches := re.FindStringSubmatch(res.PromotionCodeRedemption.SubLabel)
		if len(matches) != 2 {
			return fmt.Errorf("failed to get coupon value")
		}
		s.state.CouponValue = matches[1]

		return nil
	}
	return fmt.Errorf("no valid coupons found")
}
