/*
	DISCLAIMER:
	Методы этого REST клиента API версии 2 частично используют стурктуры,
	представленные в OpenAPI Go SDK (github.com/Tinkoff/invest-openapi-go-sdk),
	написанном под API версии 1. В связи с этим, читатель может встретить здесь
	большое колчество крайне сомнительных решений.
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"time"
)

const apiUrl string = "https://invest-public-api.tinkoff.ru/rest/tinkoff.public.invest.api.contract.v1"

type RestClientV2 struct {
	token   string
	appname string
}

type apiError struct {
	Code        int    `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description"`
}

func (c *RestClientV2) request(path string, payload any) ([]byte, error) {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	bb, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, path, bytes.NewReader(bb))
	if err != nil {
		return nil, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.token)
	req.Header.Add("x-app-name", c.appname)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			var apiError apiError
			body, err := ioutil.ReadAll(resp.Body)
			MaybeCrash(err)
			err = json.Unmarshal(body, &apiError)
			MaybeCrash(err)
			return nil, fmt.Errorf("%v", apiError.Message)
		}
		return nil, fmt.Errorf("bad response code '%v' from %v", resp.Status, path)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

func (c *RestClientV2) CloseSandboxAccount(accountId string) error {
	path := apiUrl + ".SandboxService/CloseSandboxAccount"

	payload := struct {
		AccountId string `json:"accountId"`
	}{AccountId: accountId}

	_, err := c.request(path, payload)
	if err != nil {
		return err
	}

	return nil
}

func (c *RestClientV2) GetAccounts() ([]sdk.Account, error) {
	path := apiUrl + ".UsersService/GetAccounts"

	payload := struct{}{}
	respBody, err := c.request(path, payload)
	if err != nil {
		return nil, err
	}

	accountsResp := struct {
		Accounts []struct {
			Id          string    `json:"id"`
			Type        string    `json:"type"`
			Name        string    `json:"name"`
			Status      string    `json:"status"`
			OpenedDate  time.Time `json:"openedDate"`
			ClosedDate  time.Time `json:"closedDate"`
			AccessLevel string    `json:"accessLevel"`
		} `json:"accounts"`
	}{}
	err = json.Unmarshal(respBody, &accountsResp)
	if err != nil {
		return nil, err
	}

	accounts := make([]sdk.Account, 0)
	for _, account := range accountsResp.Accounts {
		accounts = append(accounts, sdk.Account{ID: account.Id})
	}

	return accounts, nil
}

func (c *RestClientV2) GetCandles(from time.Time, to time.Time, interval sdk.CandleInterval, figi string) ([]sdk.Candle, error) {
	path := apiUrl + ".MarketDataService/GetCandles"

	candleIntervalsV1ToV2 := map[sdk.CandleInterval]string{
		sdk.CandleInterval1Min: "CANDLE_INTERVAL_1_MIN",
		//		sdk.CandleInterval2Min:   "",
		//		sdk.CandleInterval3Min:   "",
		sdk.CandleInterval5Min: "CANDLE_INTERVAL_5_MIN",
		//		sdk.CandleInterval10Min:  "",
		sdk.CandleInterval15Min: "CANDLE_INTERVAL_15_MIN",
		//		sdk.CandleInterval30Min:  "",
		sdk.CandleInterval1Hour: "CANDLE_INTERVAL_HOUR",
		//		sdk.CandleInterval2Hour:  "",
		//		sdk.CandleInterval4Hour:  "",
		sdk.CandleInterval1Day: "CANDLE_INTERVAL_DAY",
		//		sdk.CandleInterval1Week:  "",
		//		sdk.CandleInterval1Month: "",
	}

	payload := struct {
		FIGI     string    `json:"figi"`
		From     time.Time `json:"from"`
		To       time.Time `json:"to"`
		Interval string    `json:"interval"`
	}{figi, from, to, candleIntervalsV1ToV2[interval]}

	respBody, err := c.request(path, payload)
	if err != nil {
		return nil, err
	}

	candlesResp := struct {
		Candles []struct {
			Volume     string    `json:"volume"`
			High       Quotation `json:"high"`
			Low        Quotation `json:"low"`
			Close      Quotation `json:"close"`
			Open       Quotation `json:"open"`
			Time       time.Time `json:"time"`
			IsComplete bool      `json:"isComplete"`
		} `json:"candles"`
	}{}
	err = json.Unmarshal(respBody, &candlesResp)
	if err != nil {
		return nil, err
	}

	candles := make([]sdk.Candle, 0)
	for _, candle := range candlesResp.Candles {
		volume, _ := strconv.Atoi(candle.Volume)
		candles = append(candles, sdk.Candle{
			FIGI:       figi,
			Interval:   interval,
			OpenPrice:  candle.Open.ToFloat(),
			ClosePrice: candle.Close.ToFloat(),
			HighPrice:  candle.High.ToFloat(),
			LowPrice:   candle.Low.ToFloat(),
			Volume:     float64(volume),
			TS:         candle.Time,
		})
	}

	return candles, nil
}

func (c *RestClientV2) GetInfo() (UserInfo, error) {
	path := apiUrl + ".UsersService/GetInfo"

	respBody, err := c.request(path, struct{}{})
	if err != nil {
		return UserInfo{}, err
	}

	var userInfo UserInfo
	err = json.Unmarshal(respBody, &userInfo)
	if err != nil {
		return UserInfo{}, err
	}

	return userInfo, nil
}

func (c *RestClientV2) GetMarginAttributes(accountId string) (MarginAttributes, error) {
	path := apiUrl + ".UsersService/GetMarginAttributes"

	payload := struct {
		AccountId string `json:"accountId"`
	}{accountId}

	respBody, err := c.request(path, payload)
	if err != nil {
		return MarginAttributes{}, err
	}

	var marginAttributes MarginAttributes
	err = json.Unmarshal(respBody, &marginAttributes)
	if err != nil {
		return MarginAttributes{}, err
	}

	return marginAttributes, nil
}

// GetPortfolio обращается к методу GetPositions (GetSandboxPositions) API версии 2 и формирует sdk.Portfolio
func (c *RestClientV2) GetPortfolio(accountId string, isSandbox bool) (sdk.Portfolio, error) {
	path := apiUrl
	if isSandbox {
		path += ".SandboxService/GetSandboxPositions"
	} else {
		path += ".OperationsService/GetPositions"
	}

	payload := struct {
		AccountId string `json:"accountId"`
	}{AccountId: accountId}

	respBody, err := c.request(path, payload)
	if err != nil {
		return sdk.Portfolio{}, err
	}

	type position struct {
		Blocked string `json:"blocked"`
		Balance string `json:"balance"`
		FIGI    string `json:"figi"`
	}
	positionsResp := struct {
		LimitsLoadingInProgress bool         `json:"limitsLoadingInProgress"`
		Money                   []MoneyValue `json:"money"`
		Blocked                 []MoneyValue `json:"blocked"`
		Futures                 []position   `json:"futures"`
		Securities              []position   `json:"securities"`
	}{}
	err = json.Unmarshal(respBody, &positionsResp)
	if err != nil {
		return sdk.Portfolio{}, err
	}

	portfolio := sdk.Portfolio{
		Positions:  []sdk.PositionBalance{},
		Currencies: []sdk.CurrencyBalance{},
	}
	for _, money := range positionsResp.Money {
		portfolio.Currencies = append(portfolio.Currencies, sdk.CurrencyBalance{
			Currency: sdk.Currency(money.Currency),
			Balance:  money.ToFloat(),
			Blocked:  0,
		})
	}
	for _, security := range positionsResp.Securities {
		balance, _ := strconv.Atoi(security.Balance)
		blocked, _ := strconv.Atoi(security.Blocked)
		portfolio.Positions = append(portfolio.Positions, sdk.PositionBalance{
			FIGI:    security.FIGI,
			Balance: float64(balance),
			Blocked: float64(blocked),
			Lots:    0, // в клиентском коде надо Balance делить на лотность
		})
	}

	return portfolio, nil
}

func (c *RestClientV2) OpenSandboxAccount() (sdk.Account, error) {
	path := apiUrl + ".SandboxService/OpenSandboxAccount"

	payload := struct{}{}
	respBody, err := c.request(path, payload)
	if err != nil {
		return sdk.Account{}, err
	}

	accountResp := struct {
		ID string `json:"accountId"`
	}{}
	err = json.Unmarshal(respBody, &accountResp)
	if err != nil {
		return sdk.Account{}, err
	}

	return sdk.Account{ID: accountResp.ID}, nil
}

func (c *RestClientV2) PostMarketOrder(figi string, lots int, direction sdk.OperationType, accountId string, isSandbox bool) (sdk.PlacedOrder, error) {
	path := apiUrl
	if isSandbox {
		path += ".SandboxService/PostSandboxOrder"
	} else {
		path += ".OrdersService/PostOrder"
	}

	orderDirectionsV1ToV2 := map[sdk.OperationType]string{
		sdk.BUY:  "ORDER_DIRECTION_BUY",
		sdk.SELL: "ORDER_DIRECTION_SELL",
	}

	payload := struct {
		FIGI     string `json:"figi"`
		Quantity string `json:"quantity"`
		Price    struct {
			Nano  int    `json:"nano"`
			Units string `json:"units"`
		} `json:"price"`
		Direction string `json:"direction"`
		AccountId string `json:"accountId"`
		OrderType string `json:"orderType"`
		OrderId   string `json:"orderId"`
	}{
		FIGI:      figi,
		Quantity:  fmt.Sprint(lots),
		Direction: orderDirectionsV1ToV2[direction],
		AccountId: accountId,
		OrderType: "ORDER_TYPE_MARKET",
		OrderId:   "",
	}

	respBody, err := c.request(path, payload)
	if err != nil {
		return sdk.PlacedOrder{}, err
	}

	orderResp := struct {
		OrderId               string     `json:"orderId"`
		FIGI                  string     `json:"figi"`
		ExecutionReportStatus string     `json:"executionReportStatus"`
		InitialOrderPrice     MoneyValue `json:"initialOrderPrice"`
		InitialCommission     MoneyValue `json:"initialCommission"`
		Message               string     `json:"message"`
		LotsExecuted          string     `json:"lotsExecuted"`
		TotalOrderAmount      MoneyValue `json:"totalOrderAmount"`
		LotsRequested         string     `json:"lotsRequested"`
		InitialOrderPricePt   Quotation  `json:"initialOrderPricePt"`
		ExecutedOrderPrice    MoneyValue `json:"executedOrderPrice"`
		ExecutedCommission    MoneyValue `json:"executedCommission"`
		InitialSecurityPrice  MoneyValue `json:"initialSecurityPrice"`
		AciValue              MoneyValue `json:"aciValue"`
	}{}
	err = json.Unmarshal(respBody, &orderResp)
	if err != nil {
		return sdk.PlacedOrder{}, err
	}

	statusV2ToV1 := map[string]sdk.OrderStatus{
		"EXECUTION_REPORT_STATUS_FILL":          sdk.OrderStatusFill,
		"EXECUTION_REPORT_STATUS_REJECTED":      sdk.OrderStatusRejected,
		"EXECUTION_REPORT_STATUS_CANCELLED":     sdk.OrderStatusCancelled,
		"EXECUTION_REPORT_STATUS_NEW":           sdk.OrderStatusNew,
		"EXECUTION_REPORT_STATUS_PARTIALLYFILL": sdk.OrderStatusPartiallyFill,
	}

	lotsRequested, _ := strconv.Atoi(orderResp.LotsRequested)
	lotsExecuted, _ := strconv.Atoi(orderResp.LotsExecuted)
	order := sdk.PlacedOrder{
		ID:            orderResp.OrderId,
		Operation:     direction,
		Status:        statusV2ToV1[orderResp.ExecutionReportStatus],
		RejectReason:  "",
		RequestedLots: lotsRequested,
		ExecutedLots:  lotsExecuted,
		Commission: sdk.MoneyAmount{
			Currency: sdk.RUB,
			Value:    orderResp.ExecutedCommission.ToFloat(),
		},
		Message: orderResp.Message,
	}

	return order, nil
}

func (c *RestClientV2) SandboxPayIn(accountId string, currency sdk.Currency, money float64) (MoneyValue, error) {
	path := apiUrl + ".SandboxService/SandboxPayIn"

	integer, fraction := math.Modf(money)
	units := fmt.Sprint(int(integer))
	payload := struct {
		AccountId string     `json:"accountId"`
		Amount    MoneyValue `json:"amount"`
	}{accountId,
		MoneyValue{
			Currency: string(currency),
			Units:    units,
			Nano:     int32(fraction * float64(len(units))),
		}}

	respBody, err := c.request(path, payload)
	if err != nil {
		return MoneyValue{}, err
	}

	balanceResp := struct {
		Balance MoneyValue `json:"balance"`
	}{}
	err = json.Unmarshal(respBody, &balanceResp)
	if err != nil {
		return MoneyValue{}, err
	}

	return balanceResp.Balance, nil
}

func (c *RestClientV2) ShareBy(idType InstrumentIdType, classCode string, id string) (Share, error) {
	path := apiUrl + ".InstrumentsService/ShareBy"

	payload := struct {
		IdType    InstrumentIdType `json:"idType"`
		ClassCode string           `json:"classCode"`
		Id        string           `json:"id"`
	}{idType, classCode, id}

	respBody, err := c.request(path, payload)
	if err != nil {
		return Share{}, err
	}

	var share Share
	err = json.Unmarshal(respBody, &share)
	if err != nil {
		return Share{}, err
	}

	return share, nil
}
