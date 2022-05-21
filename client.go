package main

import (
	"context"
	"crypto/tls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"time"
	investapi "tinkoff-invest-contest/investAPI"
)

const ServiceAddress string = "invest-public-api.tinkoff.ru:443"

type Client struct {
	token            string
	appname          string
	marketDataStream investapi.MarketDataStreamService_MarketDataStreamClient

	InstrumentsService      investapi.InstrumentsServiceClient
	OperationsService       investapi.OperationsServiceClient
	OrdersService           investapi.OrdersServiceClient
	MarketDataService       investapi.MarketDataServiceClient
	SandboxService          investapi.SandboxServiceClient
	UsersService            investapi.UsersServiceClient
	StopOrdersService       investapi.StopOrdersServiceClient
	MarketDataStreamService investapi.MarketDataStreamServiceClient
	OrdersStreamService     investapi.OrdersStreamServiceClient
}

func NewClient(token string) *Client {
	var err error
	clientConn, err := grpc.Dial(ServiceAddress, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	if err != nil {
		log.Fatalln(err)
	}
	client := Client{
		token:                   token,
		InstrumentsService:      investapi.NewInstrumentsServiceClient(clientConn),
		OperationsService:       investapi.NewOperationsServiceClient(clientConn),
		OrdersService:           investapi.NewOrdersServiceClient(clientConn),
		MarketDataService:       investapi.NewMarketDataServiceClient(clientConn),
		SandboxService:          investapi.NewSandboxServiceClient(clientConn),
		UsersService:            investapi.NewUsersServiceClient(clientConn),
		StopOrdersService:       investapi.NewStopOrdersServiceClient(clientConn),
		MarketDataStreamService: investapi.NewMarketDataStreamServiceClient(clientConn),
		OrdersStreamService:     investapi.NewOrdersStreamServiceClient(clientConn),
	}
	client.marketDataStream, err = client.MarketDataStreamService.MarketDataStream(
		newContextWithBearerToken(token),
	)
	if err != nil {
		log.Fatalln(err)
	}
	return &client
}

func newContextWithBearerToken(token string) context.Context {
	md := metadata.New(map[string]string{
		"x-app-name":    AppName,
		"Authorization": "Bearer " + token,
	})
	return metadata.NewOutgoingContext(context.Background(), md)
}

func (c *Client) GetAccounts() ([]*investapi.Account, error) {
	accountsResp, err := c.UsersService.GetAccounts(
		newContextWithBearerToken(c.token),
		&investapi.GetAccountsRequest{},
	)
	if err != nil {
		return nil, err
	}
	return accountsResp.Accounts, nil
}

func (c *Client) GetMarginAttributes(accountId string) (*investapi.GetMarginAttributesResponse, error) {
	marginAttributesResp, err := c.UsersService.GetMarginAttributes(
		newContextWithBearerToken(c.token),
		&investapi.GetMarginAttributesRequest{
			AccountId: accountId,
		},
	)
	if err != nil {
		return nil, err
	}
	return marginAttributesResp, nil
}

func (c *Client) GetInfo() (*investapi.GetInfoResponse, error) {
	infoResp, err := c.UsersService.GetInfo(
		newContextWithBearerToken(c.token),
		&investapi.GetInfoRequest{},
	)
	if err != nil {
		return nil, err
	}
	return infoResp, nil
}

func (c *Client) GetPositions(accountId string) (*investapi.PositionsResponse, error) {
	positionsResp, err := c.OperationsService.GetPositions(
		newContextWithBearerToken(c.token),
		&investapi.PositionsRequest{
			AccountId: accountId,
		},
	)
	if err != nil {
		return nil, err
	}
	return positionsResp, nil
}

func (c *Client) GetSandboxPositions(accountId string) (*investapi.PositionsResponse, error) {
	positionsResp, err := c.SandboxService.GetSandboxPositions(
		newContextWithBearerToken(c.token),
		&investapi.PositionsRequest{
			AccountId: accountId,
		},
	)
	if err != nil {
		return nil, err
	}
	return positionsResp, nil
}

func (c *Client) ShareBy(idType investapi.InstrumentIdType, classCode string, id string) (*investapi.Share, error) {
	shareResp, err := c.InstrumentsService.ShareBy(
		newContextWithBearerToken(c.token),
		&investapi.InstrumentRequest{
			IdType:    idType,
			ClassCode: classCode,
			Id:        id,
		},
	)
	if err != nil {
		return nil, err
	}
	return shareResp.Instrument, nil
}

func (c *Client) OpenSandboxAccount() (*investapi.OpenSandboxAccountResponse, error) {
	openSandboxAccountResp, err := c.SandboxService.OpenSandboxAccount(
		newContextWithBearerToken(c.token),
		&investapi.OpenSandboxAccountRequest{},
	)
	if err != nil {
		return nil, err
	}
	return openSandboxAccountResp, nil
}

func (c *Client) SandboxPayIn(accountId string, currency string, amount float64) (*investapi.SandboxPayInResponse, error) {
	sandboxPayInResp, err := c.SandboxService.SandboxPayIn(
		newContextWithBearerToken(c.token),
		&investapi.SandboxPayInRequest{
			AccountId: accountId,
			Amount:    MoneyValueFromFloat(currency, amount),
		},
	)
	if err != nil {
		return nil, err
	}
	return sandboxPayInResp, nil
}

func (c *Client) CloseSandboxAccount(accountId string) (*investapi.CloseSandboxAccountResponse, error) {
	closeSandboxAccountResp, err := c.SandboxService.CloseSandboxAccount(
		newContextWithBearerToken(c.token),
		&investapi.CloseSandboxAccountRequest{
			AccountId: accountId,
		},
	)
	if err != nil {
		return nil, err
	}
	return closeSandboxAccountResp, nil
}

func (c *Client) GetCandles(figi string, from time.Time, to time.Time, interval investapi.CandleInterval) ([]*investapi.HistoricCandle, error) {
	candlesResp, err := c.MarketDataService.GetCandles(
		newContextWithBearerToken(c.token),
		&investapi.GetCandlesRequest{
			Figi:     figi,
			From:     timestamppb.New(from),
			To:       timestamppb.New(to),
			Interval: interval,
		},
	)
	if err != nil {
		return nil, err
	}
	return candlesResp.Candles, nil
}

func (c *Client) GetSandboxAccounts() ([]*investapi.Account, error) {
	sandboxAccountsResp, err := c.SandboxService.GetSandboxAccounts(
		newContextWithBearerToken(c.token),
		&investapi.GetAccountsRequest{},
	)
	if err != nil {
		return nil, err
	}
	return sandboxAccountsResp.Accounts, nil
}

func (c *Client) PostOrder(figi string, quantity int64, price float64, direction investapi.OrderDirection,
	accountId string, orderType investapi.OrderType, orderId string) (*investapi.PostOrderResponse, error) {
	postOrderResp, err := c.OrdersService.PostOrder(
		newContextWithBearerToken(c.token),
		&investapi.PostOrderRequest{
			Figi:      figi,
			Quantity:  quantity,
			Price:     QuotationFromFloat(price),
			Direction: direction,
			AccountId: accountId,
			OrderType: orderType,
			OrderId:   orderId,
		},
	)
	if err != nil {
		return nil, err
	}
	return postOrderResp, nil
}

func (c *Client) PostSandboxOrder(figi string, quantity int64, price float64, direction investapi.OrderDirection,
	accountId string, orderType investapi.OrderType, orderId string) (*investapi.PostOrderResponse, error) {
	postOrderResp, err := c.SandboxService.PostSandboxOrder(
		newContextWithBearerToken(c.token),
		&investapi.PostOrderRequest{
			Figi:      figi,
			Quantity:  quantity,
			Price:     QuotationFromFloat(price),
			Direction: direction,
			AccountId: accountId,
			OrderType: orderType,
			OrderId:   orderId,
		},
	)
	if err != nil {
		return nil, err
	}
	return postOrderResp, nil
}

func (c *Client) GetOrderState(accountId string, orderId string) (*investapi.OrderState, error) {
	orderState, err := c.OrdersService.GetOrderState(
		newContextWithBearerToken(c.token),
		&investapi.GetOrderStateRequest{
			AccountId: accountId,
			OrderId:   orderId,
		},
	)
	if err != nil {
		return nil, err
	}
	return orderState, nil
}

func (c *Client) GetSandboxOrderState(accountId string, orderId string) (*investapi.OrderState, error) {
	orderState, err := c.SandboxService.GetSandboxOrderState(
		newContextWithBearerToken(c.token),
		&investapi.GetOrderStateRequest{
			AccountId: accountId,
			OrderId:   orderId,
		},
	)
	if err != nil {
		return nil, err
	}
	return orderState, nil
}

func (c *Client) RunMarketDataStreamLoop(handler func(marketDataResp *investapi.MarketDataResponse)) error {
	for {
		marketDataResp, err := c.marketDataStream.Recv()
		if err != nil {
			return err
		}
		handler(marketDataResp)
	}
}

func (c *Client) SubscribeInfo(figi string) error {
	instruments := []*investapi.InfoInstrument{
		{Figi: figi},
	}
	err := c.marketDataStream.Send(&investapi.MarketDataRequest{Payload: &investapi.MarketDataRequest_SubscribeInfoRequest{
		SubscribeInfoRequest: &investapi.SubscribeInfoRequest{
			SubscriptionAction: investapi.SubscriptionAction_SUBSCRIPTION_ACTION_SUBSCRIBE,
			Instruments:        instruments,
		},
	}})
	return err
}

func (c *Client) SubscribeCandles(figi string, interval investapi.SubscriptionInterval) error {
	instruments := []*investapi.CandleInstrument{
		{
			Figi:     figi,
			Interval: interval,
		},
	}
	err := c.marketDataStream.Send(&investapi.MarketDataRequest{Payload: &investapi.MarketDataRequest_SubscribeCandlesRequest{
		SubscribeCandlesRequest: &investapi.SubscribeCandlesRequest{
			SubscriptionAction: investapi.SubscriptionAction_SUBSCRIPTION_ACTION_SUBSCRIBE,
			Instruments:        instruments,
		},
	}})
	return err
}

func (c *Client) UnsubscribeInfo(figi string) error {
	instruments := []*investapi.InfoInstrument{
		{Figi: figi},
	}
	err := c.marketDataStream.Send(&investapi.MarketDataRequest{Payload: &investapi.MarketDataRequest_SubscribeInfoRequest{
		SubscribeInfoRequest: &investapi.SubscribeInfoRequest{
			SubscriptionAction: investapi.SubscriptionAction_SUBSCRIPTION_ACTION_UNSUBSCRIBE,
			Instruments:        instruments,
		},
	}})
	return err
}

func (c *Client) UnsubscribeCandles(figi string, interval investapi.SubscriptionInterval) error {
	instruments := []*investapi.CandleInstrument{
		{
			Figi:     figi,
			Interval: interval,
		},
	}
	err := c.marketDataStream.Send(&investapi.MarketDataRequest{Payload: &investapi.MarketDataRequest_SubscribeCandlesRequest{
		SubscribeCandlesRequest: &investapi.SubscribeCandlesRequest{
			SubscriptionAction: investapi.SubscriptionAction_SUBSCRIPTION_ACTION_UNSUBSCRIBE,
			Instruments:        instruments,
		},
	}})
	return err
}
