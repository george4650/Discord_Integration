package clients

import (
	"context"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"golang.org/x/oauth2"
	"net/http"
	"wh-hard01.kol.wb.ru/wh-tech/wh-tech-back/wh_tech_discord_api/internal/common"
	"wh-hard01.kol.wb.ru/wh-tech/wh-tech-back/wh_tech_discord_api/internal/models"
	client "wh-hard01.kol.wb.ru/wh_core/gocore_http"
	"wh-hard01.kol.wb.ru/wh_core/gocore_service_configs/configs"
)

type DiscordApiClient struct {
	fastHttpClient  *client.HttpClient
	oauthConfig     oauth2.Config
	discordBotToken string
	discordServerID int
}

const (
	endpointAPI    = "/api/v9/"
	endpointUsers  = endpointAPI + "users"
	endpointGuilds = endpointAPI + "guilds"
)

func NewDiscordApiClient() *DiscordApiClient {
	return &DiscordApiClient{}
}

func (r *DiscordApiClient) Configure(ctx context.Context, config configs.Config) {
	var botConfig models.BotConfig
	if err := jsoniter.Unmarshal(config.GetByServiceKeyRequired("discord_bot"), &botConfig); err != nil {
		logrus.Panicf("Error while parsing discord_bot configs - %s", err.Error())
	}
	r.oauthConfig = oauth2.Config{
		Endpoint: oauth2.Endpoint{
			AuthURL:  common.AuthURL,
			TokenURL: common.TokenURL,
		},
		ClientID:     botConfig.DiscordBotClientID,
		ClientSecret: botConfig.DiscordBotClientSecret,
		RedirectURL:  botConfig.DiscordOAUTH2ApiCallback,
		Scopes:       common.Scopes,
	}
	r.discordServerID = botConfig.DiscordServerID
	r.discordBotToken = botConfig.DiscordBotToken
	r.fastHttpClient = client.NewHttpClient()
	r.fastHttpClient.InitHttpClient("discord_api_client")(ctx, config)
	logrus.Debugf("bot id - %v", botConfig.DiscordBotClientID)
}

func (r *DiscordApiClient) GetRedirectURI(state string) string {
	return r.oauthConfig.AuthCodeURL(state)
}

func (r *DiscordApiClient) GetAccessTokenByCode(code string) (string, error) {
	token, err := r.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return "", fmt.Errorf("[GetAccessTokenByCode] Ошибка при попытке получить access token пользователя - %w", err)
	}
	return token.AccessToken, nil
}

func (r *DiscordApiClient) GetUserInfoByAccessToken(accessToken string) (*models.User, error) {
	api := fmt.Sprintf("%s/@me", endpointUsers)

	resp, err := r.fastHttpClient.HTTPRequestWithOpts(nil, http.MethodGet, nil, api, nil, func(request *fasthttp.Request) error {
		request.Header.Set("Authorization", "Bearer "+accessToken)
		return nil
	})
	if err != nil {
		return nil, r.fastHttpClient.GenerateError(resp, err, api)
	}

	if resp.StatusCode() == http.StatusOK {
		user := models.User{}
		err = jsoniter.Unmarshal(resp.Body(), &user)
		if err != nil {
			return nil, fmt.Errorf("[GetUserInfoByAccessToken] Unmarshal body error - %w", err)
		}
		return &user, nil
	}

	return nil, r.fastHttpClient.GenerateError(resp, err, api)
}

func (r *DiscordApiClient) AddUserToServer(userID, accessToken, employeeFullName string, roles []string) (*models.Member, error) {
	api := fmt.Sprintf("%s/%d/members/%s", endpointGuilds, r.discordServerID, userID)

	data := models.AddMemberParams{
		AccessToken: accessToken,
		Roles:       roles,
		Nick:        employeeFullName,
	}

	resp, err := r.fastHttpClient.HTTPRequestWithOpts(nil, http.MethodPut, &data, api, nil, func(request *fasthttp.Request) error {
		request.Header.Set("Authorization", r.discordBotToken)
		return nil
	})
	if err != nil {
		return nil, r.fastHttpClient.GenerateError(resp, err, api)
	}

	switch resp.StatusCode() {
	case http.StatusCreated:
		member := models.Member{}
		err = jsoniter.Unmarshal(resp.Body(), &member)
		if err != nil {
			return nil, fmt.Errorf("[AddUserToServer] Unmarshal body error - %w", err)
		}
		return &member, nil
	case http.StatusNoContent:
		return nil, nil
	}

	return nil, r.fastHttpClient.GenerateError(resp, err, api)
}

func (r *DiscordApiClient) BanMember(userID string, reason *string) error {
	api := fmt.Sprintf("%s/%d/bans/%s", endpointGuilds, r.discordServerID, userID)

	data := models.BanReason{Reason: reason}

	resp, err := r.fastHttpClient.HTTPRequestWithOpts(nil, http.MethodPut, &data, api, nil, func(request *fasthttp.Request) error {
		request.Header.Set("Authorization", r.discordBotToken)
		return nil
	})
	if err != nil {
		return r.fastHttpClient.GenerateError(resp, err, api)
	}

	if resp.StatusCode() == http.StatusNoContent {
		return nil
	}
	return r.fastHttpClient.GenerateError(resp, err, api)
}

func (r *DiscordApiClient) KickMember(userID string, reason *string) error {
	api := fmt.Sprintf("%s/%d/members/%s", endpointGuilds, r.discordServerID, userID)

	data := models.BanReason{Reason: reason}

	resp, err := r.fastHttpClient.HTTPRequestWithOpts(nil, http.MethodDelete, &data, api, nil, func(request *fasthttp.Request) error {
		request.Header.Set("Authorization", r.discordBotToken)
		return nil
	})
	if err != nil {
		return r.fastHttpClient.GenerateError(resp, err, api)
	}

	if resp.StatusCode() == http.StatusNoContent {
		return nil
	}
	return r.fastHttpClient.GenerateError(resp, err, api)
}

func (r *DiscordApiClient) CheckValidToken() error {
	api := fmt.Sprintf("%s/@me", endpointUsers)

	resp, err := r.fastHttpClient.HTTPRequestWithOpts(nil, http.MethodGet, nil, api, nil, func(request *fasthttp.Request) error {
		request.Header.Set("Authorization", r.discordBotToken)
		return nil
	})
	if err != nil {
		return r.fastHttpClient.GenerateError(resp, err, api)
	}

	if resp.StatusCode() == http.StatusOK {
		return nil
	}

	return r.fastHttpClient.GenerateError(resp, err, api)
}
