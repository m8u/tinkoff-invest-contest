package api

import (
	"encoding/json"
	"tinkoff-invest-contest/internal/utils"
)

type response struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Payload any    `json:"payload"`
}

func marshalResponse(status int, message string, payload ...any) string {
	bytes, err := json.Marshal(response{
		Status:  status,
		Message: message,
		Payload: payload,
	})
	utils.MaybeCrash(err)
	return string(bytes)
}
