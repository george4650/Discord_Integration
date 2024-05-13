package main

import (
	"wh-hard01.kol.wb.ru/wh-tech/wh-tech-back/wh_tech_discord_api/clients"
	"wh-hard01.kol.wb.ru/wh-tech/wh-tech-back/wh_tech_discord_api/internal/jwt"
	"wh-hard01.kol.wb.ru/wh-tech/wh-tech-back/wh_tech_discord_api/internal/service"
	"wh-hard01.kol.wb.ru/wh_core/gocore_rest_auto_api"
)

func main() {
	runner := rest_auto_api.NewService()
	runner.DisableJWTTokenAuth()
	whTechWebApiClient := clients.NewWhTechWebApiClient()
	discordApiClient := clients.NewDiscordApiClient()
	jwtAuth := jwt.NewJwtAuth()
	discordApi := service.NewDiscordApiService(jwtAuth, discordApiClient, whTechWebApiClient)
	runner.RegisterConfigurableEntity(jwtAuth.Configure)
	runner.RegisterConfigurableEntity(discordApiClient.Configure)
	runner.RegisterConfigurableEntity(whTechWebApiClient.Configure)
	runner.RegisterConfigurableEntity(discordApi.Configure)
	runner.RegisterCustomHandler("AddUserToServer", discordApi.AddUserToServer)
	runner.RegisterCustomHandler("ApiCallBack", discordApi.ApiCallBack)
	runner.RegisterCustomHandler("BanMember", discordApi.BanMember)
	runner.RegisterCustomHandler("KickMember", discordApi.KickMember)
	runner.RegisterCustomHandler("CheckValidToken", discordApi.CheckValidToken)
	runner.Run()
}
