package client

import (
	"context"
	"crypto/tls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/utils"
)

const ServiceAddress string = "invest-public-api.tinkoff.ru:443"
const AppName = "m8u"

type Client struct {
	token   string
	appname string

	marketDataStream investapi.MarketDataStreamService_MarketDataStreamClient
	tradesStream     investapi.OrdersStreamService_TradesStreamClient

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

// NewClient creates a new Tinkoff Invest API gRPC client
func NewClient(token string) *Client {
	utils.WaitForInternetConnection()
	var err error
	clientConn, err := grpc.Dial(ServiceAddress, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	utils.MaybeCrash(err)
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

	return &client
}

// InitMarketDataStream initializes a market data stream.
// Call it after creating a Client if you need any data from Invest API market data streams
func (c *Client) InitMarketDataStream() {
	var err error
	c.marketDataStream, err = c.MarketDataStreamService.MarketDataStream(
		newContextWithBearerToken(c.token),
	)
	utils.MaybeCrash(err)
}

func (c *Client) InitTradesStream(accountIds []string) {
	var err error
	c.tradesStream, err = c.OrdersStreamService.TradesStream(
		newContextWithBearerToken(c.token),
		&investapi.TradesStreamRequest{Accounts: accountIds},
	)
	utils.MaybeCrash(err)
}

func newContextWithBearerToken(token string) context.Context {
	md := metadata.New(map[string]string{
		"x-app-name":    AppName,
		"Authorization": "Bearer " + token,
	})
	return metadata.NewOutgoingContext(context.Background(), md)
}

func (c *Client) BondBy(idType investapi.InstrumentIdType, classCode string, id string) (*investapi.Bond, error) {
	utils.WaitForInternetConnection()
	bondResp, err := c.InstrumentsService.BondBy(
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
	return bondResp.Instrument, nil
}

func (c *Client) CloseSandboxAccount(accountId string) (*investapi.CloseSandboxAccountResponse, error) {
	utils.WaitForInternetConnection()
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

func (c *Client) CurrencyBy(idType investapi.InstrumentIdType, classCode string, id string) (*investapi.Currency, error) {
	utils.WaitForInternetConnection()
	currencyResp, err := c.InstrumentsService.CurrencyBy(
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
	return currencyResp.Instrument, nil
}

func (c *Client) EtfBy(idType investapi.InstrumentIdType, classCode string, id string) (*investapi.Etf, error) {
	utils.WaitForInternetConnection()
	etfResp, err := c.InstrumentsService.EtfBy(
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
	return etfResp.Instrument, nil
}

func (c *Client) FutureBy(idType investapi.InstrumentIdType, classCode string, id string) (*investapi.Future, error) {
	utils.WaitForInternetConnection()
	futureResp, err := c.InstrumentsService.FutureBy(
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
	return futureResp.Instrument, nil
}

func (c *Client) GetAccounts() ([]*investapi.Account, error) {
	utils.WaitForInternetConnection()
	accountsResp, err := c.UsersService.GetAccounts(
		newContextWithBearerToken(c.token),
		&investapi.GetAccountsRequest{},
	)
	if err != nil {
		return nil, err
	}
	return accountsResp.Accounts, nil
}

func (c *Client) GetCandles(figi string, from time.Time, to time.Time, interval investapi.CandleInterval) ([]*investapi.HistoricCandle, error) {
	utils.WaitForInternetConnection()
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

func (c *Client) GetInfo() (*investapi.GetInfoResponse, error) {
	utils.WaitForInternetConnection()
	infoResp, err := c.UsersService.GetInfo(
		newContextWithBearerToken(c.token),
		&investapi.GetInfoRequest{},
	)
	if err != nil {
		return nil, err
	}
	return infoResp, nil
}

func (c *Client) GetMarginAttributes(accountId string) (*investapi.GetMarginAttributesResponse, error) {
	utils.WaitForInternetConnection()
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

func (c *Client) GetOrderState(accountId string, orderId string) (*investapi.OrderState, error) {
	utils.WaitForInternetConnection()
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

func (c *Client) GetPortfolio(accountId string) (*investapi.PortfolioResponse, error) {
	utils.WaitForInternetConnection()
	portfolioResp, err := c.OperationsService.GetPortfolio(
		newContextWithBearerToken(c.token),
		&investapi.PortfolioRequest{
			AccountId: accountId,
		},
	)
	if err != nil {
		return nil, err
	}
	return portfolioResp, nil
}

func (c *Client) GetPositions(accountId string) (*investapi.PositionsResponse, error) {
	utils.WaitForInternetConnection()
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

func (c *Client) GetSandboxAccounts() ([]*investapi.Account, error) {
	utils.WaitForInternetConnection()
	sandboxAccountsResp, err := c.SandboxService.GetSandboxAccounts(
		newContextWithBearerToken(c.token),
		&investapi.GetAccountsRequest{},
	)
	if err != nil {
		return nil, err
	}
	return sandboxAccountsResp.Accounts, nil
}

func (c *Client) GetSandboxOrderState(accountId string, orderId string) (*investapi.OrderState, error) {
	utils.WaitForInternetConnection()
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

func (c *Client) GetSandboxPortfolio(accountId string) (*investapi.PortfolioResponse, error) {
	utils.WaitForInternetConnection()
	portfolioResp, err := c.SandboxService.GetSandboxPortfolio(
		newContextWithBearerToken(c.token),
		&investapi.PortfolioRequest{
			AccountId: accountId,
		},
	)
	if err != nil {
		return nil, err
	}
	return portfolioResp, nil
}

func (c *Client) GetSandboxPositions(accountId string) (*investapi.PositionsResponse, error) {
	utils.WaitForInternetConnection()
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

func (c *Client) OpenSandboxAccount() (*investapi.OpenSandboxAccountResponse, error) {
	utils.WaitForInternetConnection()
	openSandboxAccountResp, err := c.SandboxService.OpenSandboxAccount(
		newContextWithBearerToken(c.token),
		&investapi.OpenSandboxAccountRequest{},
	)
	if err != nil {
		return nil, err
	}
	return openSandboxAccountResp, nil
}

func (c *Client) PostOrder(figi string, quantity int64, price float64, direction investapi.OrderDirection,
	accountId string, orderType investapi.OrderType, orderId string) (*investapi.PostOrderResponse, error) {
	utils.WaitForInternetConnection()
	postOrderResp, err := c.OrdersService.PostOrder(
		newContextWithBearerToken(c.token),
		&investapi.PostOrderRequest{
			Figi:      figi,
			Quantity:  quantity,
			Price:     utils.FloatToQuotation(price),
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
	utils.WaitForInternetConnection()
	postOrderResp, err := c.SandboxService.PostSandboxOrder(
		newContextWithBearerToken(c.token),
		&investapi.PostOrderRequest{
			Figi:      figi,
			Quantity:  quantity,
			Price:     utils.FloatToQuotation(price),
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

func (c *Client) RunMarketDataStreamLoop(handleResponse func(marketDataResp *investapi.MarketDataResponse),
	resubscribe func()) {
	var err error
	var resp *investapi.MarketDataResponse
	for {
		utils.WaitForInternetConnection()
		if err != nil {
			resubscribe()
		}
		resp, err = c.marketDataStream.Recv()
		handleResponse(resp)
	}
}

func (c *Client) RunTradesStreamLoop(handleResponse func(tradesResp *investapi.TradesStreamResponse)) {
	var err error
	var resp *investapi.TradesStreamResponse
	for {
		utils.WaitForInternetConnection()
		resp, err = c.tradesStream.Recv()
		utils.MaybeCrash(err)
		handleResponse(resp)
	}
}

func (c *Client) SandboxPayIn(accountId string, currency string, amount float64) (*investapi.SandboxPayInResponse, error) {
	utils.WaitForInternetConnection()
	sandboxPayInResp, err := c.SandboxService.SandboxPayIn(
		newContextWithBearerToken(c.token),
		&investapi.SandboxPayInRequest{
			AccountId: accountId,
			Amount:    utils.FloatToMoneyValue(currency, amount),
		},
	)
	if err != nil {
		return nil, err
	}
	return sandboxPayInResp, nil
}

func (c *Client) ShareBy(idType investapi.InstrumentIdType, classCode string, id string) (*investapi.Share, error) {
	utils.WaitForInternetConnection()
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

func (c *Client) SubscribeCandles(figi string, interval investapi.SubscriptionInterval) error {
	utils.WaitForInternetConnection()
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

func (c *Client) SubscribeInfo(figi string) error {
	utils.WaitForInternetConnection()
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

func (c *Client) SubscribeOrderBook(figi string, depth int32) error {
	utils.WaitForInternetConnection()
	instruments := []*investapi.OrderBookInstrument{
		{
			Figi:  figi,
			Depth: depth,
		},
	}
	err := c.marketDataStream.Send(&investapi.MarketDataRequest{Payload: &investapi.MarketDataRequest_SubscribeOrderBookRequest{
		SubscribeOrderBookRequest: &investapi.SubscribeOrderBookRequest{
			SubscriptionAction: investapi.SubscriptionAction_SUBSCRIPTION_ACTION_SUBSCRIBE,
			Instruments:        instruments,
		},
	}})
	return err
}

func (c *Client) UnsubscribeCandles(figi string, interval investapi.SubscriptionInterval) error {
	utils.WaitForInternetConnection()
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

func (c *Client) UnsubscribeInfo(figi string) error {
	utils.WaitForInternetConnection()
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

func (c *Client) UnsubscribeOrderBook(figi string, depth int32) error {
	utils.WaitForInternetConnection()
	instruments := []*investapi.OrderBookInstrument{
		{
			Figi:  figi,
			Depth: depth,
		},
	}
	err := c.marketDataStream.Send(&investapi.MarketDataRequest{Payload: &investapi.MarketDataRequest_SubscribeOrderBookRequest{
		SubscribeOrderBookRequest: &investapi.SubscribeOrderBookRequest{
			SubscriptionAction: investapi.SubscriptionAction_SUBSCRIPTION_ACTION_UNSUBSCRIBE,
			Instruments:        instruments,
		},
	}})
	return err
}
