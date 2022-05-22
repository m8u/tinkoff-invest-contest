package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"time"
	investapi "tinkoff-invest-contest/investAPI"
)

// getHistoricalCandles подгружает из кэша или, если нет кэшированных, то загружает и кэширует исторические свечи
func getHistoricalCandles(client *Client, figi string, daysBeforeNow int, candleInterval investapi.CandleInterval) []*investapi.HistoricCandle {
	var candles []*investapi.HistoricCandle
	_, err := os.ReadDir("cache")
	if err != nil {
		err = os.Mkdir("cache", 0775)
		MaybeCrash(err)
	}
	filename := fmt.Sprintf("cache/%v_%v", figi, candleInterval)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Println("Downloading candles...")
		for day := daysBeforeNow; day >= 0; day-- {
			portion, err := client.GetCandles(
				figi,
				time.Now().AddDate(0, 0, -day-1),
				time.Now().AddDate(0, 0, -day),
				candleInterval,
			)
			MaybeCrash(err)
			candles = append(candles, portion...)
		}
	} else {
		err = json.Unmarshal(data, &candles)
		if err != nil {
			log.Fatalln(err)
		}

		// определяем начало нужного нам временного отрезка
		start := len(candles)
		for i := range candles {
			if candles[len(candles)-1-i].Time.AsTime().Unix() <= time.Now().AddDate(0, 0, -daysBeforeNow).Unix() {
				break
			}
			start--
		}
		candles = candles[start:]

		candleDuration := candles[1].Time.AsTime().Sub(candles[0].Time.AsTime())
		if err != nil {
			log.Fatalln(err)
		}
		// догружаем недостающие свечи
		if candles[0].Time.AsTime().Unix() > time.Now().AddDate(0, 0, -daysBeforeNow).Round(candleDuration).Unix() {
			var missingCandles []*investapi.HistoricCandle
			for day := daysBeforeNow; day >= int(time.Since(candles[0].Time.AsTime()).Hours()/24); day-- {
				portion, err := client.GetCandles(
					figi,
					time.Now().AddDate(0, 0, -day-1),
					time.Now().AddDate(0, 0, -day),
					candleInterval,
				)
				MaybeCrash(err)
				missingCandles = append(missingCandles, portion...)
			}
			missingEnd := 0
			for _, currentCandle := range missingCandles {
				if currentCandle.Time.AsTime() == candles[0].Time.AsTime() {
					break
				}
				missingEnd++
			}
			candles = append(missingCandles[:missingEnd], candles...)
		}
		if candles[len(candles)-1].Time.AsTime().Unix() < time.Now().Round(candleDuration).Unix() {
			var missingCandles []*investapi.HistoricCandle
			for day := int(time.Since(candles[len(candles)-1].Time.AsTime()).Hours() / 24); day >= 0; day-- {
				portion, err := client.GetCandles(
					figi,
					time.Now().AddDate(0, 0, -day-1),
					time.Now().AddDate(0, 0, -day),
					candleInterval,
				)
				MaybeCrash(err)
				missingCandles = append(missingCandles, portion...)
			}
			missingStart := 0
			for _, currentCandle := range missingCandles {
				missingStart++
				if currentCandle.Time.AsTime() == candles[len(candles)-1].Time.AsTime() {
					break
				}
			}
			candles = append(candles, missingCandles[missingStart:]...)
		}
	}

	data, err = json.Marshal(candles)
	if err != nil {
		log.Fatalln(err)
	}
	err = ioutil.WriteFile(filename, data, 0755)
	if err != nil {
		log.Fatalln(err)
	}

	return candles
}

type testAccount struct {
	freeMoney   float64
	lockedMoney float64
	lotsHave    int64
}

func testBuy(price float64, quantity int64, account *testAccount) {
	for i := 0; i < int(quantity); i++ {
		if account.lotsHave < 0 {
			account.lockedMoney -= price
		} else {
			account.freeMoney -= price
		}
		account.lotsHave++
	}
	account.freeMoney += account.lockedMoney
	account.lockedMoney = 0
}

func testSell(price float64, quantity int64, account *testAccount) {
	for i := 0; i < int(quantity); i++ {
		if account.lotsHave > 0 {
			account.freeMoney += price
		} else {
			account.lockedMoney += price
		}
		account.lotsHave--
	}
}

// TestOnHistoricalData тестирует стратегию на исторических данных по заданному инструменту
func TestOnHistoricalData(token string, figi string, daysBeforeNow int, candleInterval investapi.CandleInterval,
	strategyParams StrategyParams, startBalance float64, fee float64, allowMargin bool, charts *Charts) {
	account := &testAccount{
		freeMoney:   startBalance,
		lockedMoney: 0,
		lotsHave:    0,
	}

	*charts.StartBalance = startBalance

	client := NewClient(token)

	share, err := client.ShareBy(investapi.InstrumentIdType_INSTRUMENT_ID_TYPE_FIGI, "", figi)
	MaybeCrash(err)

	if allowMargin &&
		((share.Dshort.Units == 0 && share.Dshort.Nano == 0) ||
			(share.Dlong.Units == 0 && share.Dlong.Nano == 0)) {
		log.Fatalf("can't margin-trade %v (%v)", share.Ticker, figi)
	}

	log.Printf("\n"+
		"instrument: %v (%v)\n"+
		"start balance: %+v",
		share.Ticker, figi, account.freeMoney)

	candles := getHistoricalCandles(client, figi, daysBeforeNow, candleInterval)
	*charts.Candles = append(*charts.Candles, candles[:strategyParams.Window]...)

	for i := strategyParams.Window; i < len(candles); i++ {
		*charts.Candles = append(*charts.Candles, candles[i])

		tradeSignal := GetTradeSignal(
			strategyParams,
			true,
			&investapi.Candle{
				Open:  candles[i].Open,
				High:  candles[i].High,
				Low:   candles[i].Low,
				Close: candles[i].Close,
			},
			true,
			charts,
		)

		if tradeSignal != nil {
			var maxDealValue float64
			var lots int64
			switch tradeSignal.Direction {
			case investapi.OrderDirection_ORDER_DIRECTION_BUY:
				if allowMargin {
					liquidPortfolio := account.freeMoney +
						float64(account.lotsHave)*
							FloatFromQuotation(candles[i].Close)*
							float64(share.Lot)
					startMargin := float64(account.lotsHave) *
						FloatFromQuotation(candles[i].Close) *
						float64(share.Lot) *
						FloatFromQuotation(share.Dlong)
					maxDealValue = (liquidPortfolio - startMargin) / FloatFromQuotation(share.Dlong)
				} else {
					maxDealValue = account.freeMoney
				}
				lots = int64(maxDealValue / (FloatFromQuotation(candles[i].Close) * float64(share.Lot)))
				if lots == 0 {
					continue
				}
				testBuy(FloatFromQuotation(candles[i].Close)*float64(share.Lot)*(1+fee),
					lots, account)
				break
			case investapi.OrderDirection_ORDER_DIRECTION_SELL:
				if allowMargin {
					var liquidPortfolio, startMargin float64
					if account.lotsHave >= 0 { // TODO: добавить маржин-колл
						liquidPortfolio = account.freeMoney +
							float64(account.lotsHave)*
								FloatFromQuotation(candles[i].Close)*
								float64(share.Lot)
						startMargin = float64(account.lotsHave) *
							FloatFromQuotation(candles[i].Close) *
							float64(share.Lot) *
							FloatFromQuotation(share.Dlong)
					} else {
						liquidPortfolio = account.freeMoney + account.lockedMoney +
							float64(account.lotsHave)*
								FloatFromQuotation(candles[i].Close)*
								float64(share.Lot)
						startMargin = math.Abs(float64(account.lotsHave)) *
							FloatFromQuotation(candles[i].Close) *
							float64(share.Lot) *
							FloatFromQuotation(share.Dshort)
					}
					maxDealValue = (liquidPortfolio - startMargin) / FloatFromQuotation(share.Dshort)
					lots = int64(maxDealValue / (FloatFromQuotation(candles[i].Close) * float64(share.Lot)))
				} else {
					lots = account.lotsHave
				}
				if lots == 0 {
					continue
				}
				testSell(FloatFromQuotation(candles[i].Close)*float64(share.Lot)*(1-fee),
					lots, account)
				break
			}
			*charts.Flags = append(*charts.Flags, make([]ChartsTradeFlag, 0))
			(*charts.Flags)[len(*charts.Flags)-1] = append((*charts.Flags)[len(*charts.Flags)-1],
				ChartsTradeFlag{
					Direction:   tradeSignal.Direction,
					Price:       FloatFromQuotation(candles[i].Close),
					Quantity:    lots * int64(share.Lot),
					CandleIndex: len(*charts.Candles) - 1,
				},
			)
			*charts.BalanceHistory = append(*charts.BalanceHistory, account.freeMoney)
		}
	}
	// закрываем незакрытые позиции на последней свече
	if account.lotsHave < 0 {
		log.Printf("!!! WARNING: force closing shorts")
		*charts.Flags = append(*charts.Flags, make([]ChartsTradeFlag, 0))
		(*charts.Flags)[len(*charts.Flags)-1] = append((*charts.Flags)[len(*charts.Flags)-1],
			ChartsTradeFlag{
				Direction:   investapi.OrderDirection_ORDER_DIRECTION_BUY,
				Price:       FloatFromQuotation(candles[len(candles)-1].Close),
				Quantity:    -account.lotsHave,
				CandleIndex: len(candles) - 1,
			},
		)
		testBuy(FloatFromQuotation(candles[len(candles)-1].Close),
			-account.lotsHave, account)
		*charts.BalanceHistory = append(*charts.BalanceHistory, account.freeMoney)
	} else if account.lotsHave > 0 {
		log.Printf("!!! WARNING: force closing longs")
		*charts.Flags = append(*charts.Flags, make([]ChartsTradeFlag, 0))
		(*charts.Flags)[len(*charts.Flags)-1] = append((*charts.Flags)[len(*charts.Flags)-1],
			ChartsTradeFlag{
				Direction:   investapi.OrderDirection_ORDER_DIRECTION_SELL,
				Price:       FloatFromQuotation(candles[len(candles)-1].Close),
				Quantity:    account.lotsHave,
				CandleIndex: len(candles) - 1,
			},
		)
		testSell(FloatFromQuotation(candles[len(candles)-1].Close),
			account.lotsHave, account)
		*charts.BalanceHistory = append(*charts.BalanceHistory, account.freeMoney)
	}
	log.Printf("\n"+
		"TEST RESULTS\n"+
		"Balance: %+v (%+v%%)",
		account.freeMoney, (account.freeMoney-startBalance)/startBalance*100)
}
