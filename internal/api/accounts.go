package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"tinkoff-invest-contest/internal/app"
)

func CreateSandboxAccount(c *gin.Context) {
	args := struct {
		RUB float64 `form:"rub"`
		USD float64 `form:"usd"`
	}{}
	err := c.Bind(&args)
	if err != nil {
		_, _ = c.Writer.WriteString(marshalResponse(
			http.StatusBadRequest,
			"One or more arguments are invalid ("+err.Error()+")",
		))
		return
	}
	accountId := app.SandboxEnv.CreateSandboxAccount(map[string]float64{
		"rub": args.RUB,
		"usd": args.USD,
	})
	_, _ = c.Writer.WriteString(marshalResponse(
		http.StatusOK,
		"",
		struct {
			AccountId string `json:"accountId"`
		}{accountId},
	))
}

func RemoveSandboxAccount(c *gin.Context) {
	id := c.Query("id")
	app.SandboxEnv.RemoveSandboxAccount(id)
}

func GetCombatAccounts(c *gin.Context) {
	_, _ = c.Writer.WriteString(marshalResponse(
		http.StatusOK,
		"",
		app.CombatEnv.GetAccountsPayload(),
	))
}

func GetSandboxAccounts(c *gin.Context) {
	_, _ = c.Writer.WriteString(marshalResponse(
		http.StatusOK,
		"",
		app.SandboxEnv.GetAccountsPayload(),
	))
}
