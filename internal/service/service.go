package service

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"net/http"
	"wh-hard01.kol.wb.ru/wh-tech/wh-tech-back/wh_tech_discord_api/clients"
	"wh-hard01.kol.wb.ru/wh-tech/wh-tech-back/wh_tech_discord_api/internal/common"
	"wh-hard01.kol.wb.ru/wh-tech/wh-tech-back/wh_tech_discord_api/internal/constant"
	"wh-hard01.kol.wb.ru/wh-tech/wh-tech-back/wh_tech_discord_api/internal/jwt"
	"wh-hard01.kol.wb.ru/wh-tech/wh-tech-back/wh_tech_discord_api/internal/models"
	"wh-hard01.kol.wb.ru/wh-tech/wh-tech-back/wh_tech_discord_api/utils/validation"
	"wh-hard01.kol.wb.ru/wh_core/gocore_auth"
	"wh-hard01.kol.wb.ru/wh_core/gocore_rest_auto_api/validators"
	"wh-hard01.kol.wb.ru/wh_core/gocore_service_configs/configs"
	utils "wh-hard01.kol.wb.ru/wh_core/gocore_utils"
)

type DiscordApiService struct {
	whTechWebApiClient *clients.WhTechWebApiClient
	discordApiClient   *clients.DiscordApiClient
	params             models.Params
	jwtAuth            *jwt.JwtAuth
}

func NewDiscordApiService(jwtAuth *jwt.JwtAuth, discordApiClient *clients.DiscordApiClient, whTechWebApiClient *clients.WhTechWebApiClient) *DiscordApiService {
	return &DiscordApiService{
		jwtAuth:            jwtAuth,
		discordApiClient:   discordApiClient,
		whTechWebApiClient: whTechWebApiClient,
	}
}

func (g *DiscordApiService) Configure(ctx context.Context, config configs.Config) {
	if err := jsoniter.Unmarshal(config.GetByServiceKeyRequired("service_param_key"), &g.params); err != nil {
		logrus.Panicf("Error while parsing service_param_key configs - %s", err.Error())
	}
}

func (g *DiscordApiService) AddUserToServer(c *gin.Context, _ map[string]interface{}) {
	// получаем текущий employee_id
	employeeID := gocore_auth.MustGetCurrentEmployeeID(c)

	// Генерируем токен и зашиваем в него employee_id.
	token, err := g.jwtAuth.GenerateToken(employeeID)
	if err != nil {
		utils.BindServiceErrorWithAbort(c, "generate jwt token error", err)
		return
	}

	discordAuthUrl := g.discordApiClient.GetRedirectURI(token)
	utils.BindObjectToRestData(c, discordAuthUrl)
}

func (g *DiscordApiService) ApiCallBack(c *gin.Context, _ map[string]interface{}) {
	var callBackInfo models.CallBackInfo
	if err := c.ShouldBind(&callBackInfo); err != nil {
		utils.BindValidationErrorWithAbort(c, err.Error())
		return
	}

	if callBackInfo.Error != nil && callBackInfo.ErrorDescription != nil {
		if *callBackInfo.Error == constant.AccessDeniedError && *callBackInfo.ErrorDescription == constant.UserRejectedAuthRequest { // Пользователь отклонил запрос аутентификации по OAUTH2.0
			c.Redirect(http.StatusSeeOther, g.params.RedirectURLIfRejectAuth) // перекидываем пользователя на страницу от которой поступил запрос
			utils.BindNoContent(c)
			return
		}
		utils.BindServiceErrorWithAbort(c, "callBack function returned an error", fmt.Errorf("ошибка при выполнении callBack-функции: %s. Описание ошибки: %s", *callBackInfo.Error, *callBackInfo.ErrorDescription))
		return
	}

	if callBackInfo.Code == nil {
		utils.BindValidationErrorWithAbort(c, "code parameter was not passed")
		return
	}

	// Парсим токен, возвращённый в параметре state, получаем employee_id
	employeeID, err := g.jwtAuth.ParseToken(callBackInfo.State)
	if err != nil {
		utils.BindValidationErrorWithAbort(c, err.Error())
		return
	}

	// Получаем access token пользователя по коду авторизации
	accessToken, err := g.discordApiClient.GetAccessTokenByCode(*callBackInfo.Code)
	if err != nil {
		utils.BindServiceErrorWithAbort(c, "get access token by code error", fmt.Errorf("Не удалось получить access token по коду авторизации пользователя: %w. EmployeeID - %s", err, employeeID))
		return
	}

	// Получаем инфо о пользователе по access token
	user, err := g.discordApiClient.GetUserInfoByAccessToken(accessToken)
	if err != nil {
		utils.BindServiceErrorWithAbort(c, "get user info by access token error", fmt.Errorf("Не удалось получить инфо о пользователе по access token: %w. EmployeeID - %s", err, employeeID))
		return
	}

	// получаем employee_name по employee_id
	employeeName, err := g.whTechWebApiClient.GetUserDataByEmployeeID(employeeID)
	if err != nil {
		utils.BindServiceErrorWithAbort(c, err.Error(), fmt.Errorf("ошибка при получении имени сотрудника по employee_id: %w. EmployeeID - %s", err, employeeID))
		return
	}

	// получив employeeName, оставляем от него только ФИ
	employeeName, err = validation.ValidateEmployeeNameToDiscord(employeeName)
	if err != nil {
		utils.BindServiceErrorWithAbort(c, err.Error(), fmt.Errorf("ошибка при валидировании имени сотрудника по employee_id: %w. EmployeeID - %s", err, employeeID))
		return
	}

	_, err = g.discordApiClient.AddUserToServer(user.ID, accessToken, employeeName, common.DefaultRolesIds)
	if err != nil {
		utils.BindServiceErrorWithAbort(c, "add user to server error", fmt.Errorf("Не удалось добавить пользователя на сервер: %w. EmployeeID - %s", err, employeeID))
		return
	}

	c.Redirect(http.StatusSeeOther, g.params.DiscordServerURL) // Перекидываем пользователя на наш сервер (в общий чат)
	utils.BindNoContent(c)
}

func (g *DiscordApiService) BanMember(c *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		UserID string  `json:"user_id" binding:"required,max=50"`
		Reason *string `json:"reason" binding:"omitempty,max=512"`
	})
	if !ok {
		utils.BindValidationErrorWithAbort(c, "BODY validation error")
		return
	}

	err := g.discordApiClient.BanMember(body.UserID, body.Reason)
	if err != nil {
		utils.BindServiceErrorWithAbort(c, "ban member error", fmt.Errorf("Не удалось забанить пользователя сервера: %w", err))
		return
	}

	utils.BindNoContent(c)
}

func (g *DiscordApiService) KickMember(c *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		UserID string  `json:"user_id" binding:"required,max=50"`
		Reason *string `json:"reason" binding:"omitempty,max=512"`
	})
	if !ok {
		utils.BindValidationErrorWithAbort(c, "BODY validation error")
		return
	}

	err := g.discordApiClient.KickMember(body.UserID, body.Reason)
	if err != nil {
		utils.BindServiceErrorWithAbort(c, "kick member error", fmt.Errorf("Не удалось удалить пользователя с сервера: %w", err))
		return
	}

	utils.BindNoContent(c)
}

func (g *DiscordApiService) CheckValidToken(c *gin.Context, _ map[string]interface{}) {
	err := g.discordApiClient.CheckValidToken()
	if err != nil {
		utils.BindServiceErrorWithAbort(c, "check valid bot token error", fmt.Errorf("Ошибка при проверке валидности токена бота: %w", err))
		return
	}

	utils.BindNoContent(c)
}
