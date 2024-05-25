package dingding

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/wsqigo/basic-go/webook/internal/domain"
	"github.com/wsqigo/basic-go/webook/pkg/logger"
	"net/http"
	"net/url"
	"time"
)

type Service interface {
	AuthURL(ctx context.Context, state string) (string, error)
	VerifyCode(ctx context.Context, code string) (domain.DDingInfo, error)
}

var redirectURL = url.PathEscape("https://8509v633q4.yicp.fun/oauth2/dingding/callback")

type service struct {
	appID     string
	appSecret string
	client    *http.Client
	l         logger.LoggerV1
}

func NewService(appID string, appSecret string, l logger.LoggerV1) Service {
	return &service{
		appID:     appID,
		appSecret: appSecret,
		client:    http.DefaultClient,
		l:         l,
	}
}

func (s *service) VerifyCode(ctx context.Context, code string) (domain.DDingInfo, error) {
	h := hmac.New(sha256.New, []byte(s.appSecret))
	timestamp := time.Now().UnixMilli()
	strTimeStamp := fmt.Sprintf("%d", timestamp)
	h.Write([]byte(strTimeStamp))
	sha := h.Sum(nil)
	sig := base64.StdEncoding.EncodeToString(sha)
	mysig := url.QueryEscape(sig)

	accessTokenUrl := fmt.Sprintf(`https://oapi.dingtalk.com/sns/getuserinfo_bycode?signature=%s&timestamp=%d&accessKey=%s`,
		mysig, timestamp, s.appID)
	m, err := json.Marshal(map[string]string{
		"tmp_auth_code": code,
	})
	if err != nil {
		return domain.DDingInfo{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, accessTokenUrl, bytes.NewReader(m))
	if err != nil {
		return domain.DDingInfo{}, err
	}
	httpResp, err := s.client.Do(req)
	if err != nil {
		return domain.DDingInfo{}, nil
	}

	var res Result
	err = json.NewDecoder(httpResp.Body).Decode(&res)
	if err != nil {
		// 转 JSON 为结构体出错
		return domain.DDingInfo{}, err
	}
	if res.ErrCode != 0 {
		return domain.DDingInfo{},
			fmt.Errorf("调用钉钉接口失败 errcode %d, errmsg %s", res.ErrCode, res.ErrMsg)
	}
	return domain.DDingInfo{
		UnionId: res.UserInfo.UnionId,
		OpenId:  res.UserInfo.OpenId,
	}, nil
}

func (s *service) AuthURL(ctx context.Context, state string) (string, error) {
	const authURLPattern = `https://oapi.dingtalk.com/connect/qrconnect?appid=%s&&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s`
	return fmt.Sprintf(authURLPattern, s.appID, redirectURL, state), nil
}

type UserInfo struct {
	Nick                 string `json:"nick"`
	UnionId              string `json:"unionid"`
	OpenId               string `json:"openid"`
	MainOrgAuthHighLevel bool   `json:"main_org_auth_high_level"`
}

type Result struct {
	ErrCode  int      `json:"errcode"`
	ErrMsg   string   `json:"errmsg"`
	UserInfo UserInfo `json:"user_info"`
}
