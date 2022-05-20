package main

import (
	"encoding/json"
	"fmt"
	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	"io/ioutil"
	"log"
	"math"
	"os"
	"time"
)

// getHistoricalCandles подгружает из кэша или, если нет кэшированных, то загружает и кэширует исторические свечи
func getHistoricalCandles(client *RestClientV2, figi string, daysBeforeNow int, candleInterval sdk.CandleInterval) []sdk.Candle {
	var candles []sdk.Candle
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
				time.Now().AddDate(0, 0, -day-1),
				time.Now().AddDate(0, 0, -day),
				candleInterval,
				figi,
			)
			MaybeCrash(err)
			candles = append(candles, portion...)
		}
	} else {
		err = json.Unmarshal(data, &candles)
		if err != nil {
			log.Fatalln(err)
		}

		start := len(candles)
		for i := range candles {
			if candles[len(candles)-1-i].TS.Unix() <= time.Now().AddDate(0, 0, -daysBeforeNow).Unix() {
				break
			}
			start--
		}
		candles = candles[start:]

		candleDuration := candles[1].TS.Sub(candles[0].TS)
		if err != nil {
			log.Fatalln(err)
		}
		if candles[0].TS.Unix() > time.Now().AddDate(0, 0, -daysBeforeNow).Round(candleDuration).Unix() {
			var missingCandles []sdk.Candle
			for day := daysBeforeNow; day >= int(time.Since(candles[0].TS).Hours()/24); day-- {
				portion, err := client.GetCandles(
					time.Now().AddDate(0, 0, -day-1),
					time.Now().AddDate(0, 0, -day),
					candleInterval,
					figi,
				)
				MaybeCrash(err)
				missingCandles = append(missingCandles, portion...)
			}
			missingEnd := 0
			for _, currentCandle := range missingCandles {
				if currentCandle.TS == candles[0].TS {
					break
				}
				missingEnd++
			}
			candles = append(missingCandles[:missingEnd], candles...)
		}
		if candles[len(candles)-1].TS.Unix() < time.Now().Round(candleDuration).Unix() {
			var missingCandles []sdk.Candle
			for day := int(time.Since(candles[len(candles)-1].TS).Hours() / 24); day >= 0; day-- {
				portion, err := client.GetCandles(
					time.Now().AddDate(0, 0, -day-1),
					time.Now().AddDate(0, 0, -day),
					candleInterval,
					figi,
				)
				MaybeCrash(err)
				missingCandles = append(missingCandles, portion...)
			}
			missingStart := 0
			for _, currentCandle := range missingCandles {
				missingStart++
				if currentCandle.TS == candles[len(candles)-1].TS {
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
	lotsHave    int
}

func testBuy(price float64, quantity int, account *testAccount) {
	for i := 0; i < quantity; i++ {
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

func testSell(price float64, quantity int, account *testAccount) {
	for i := 0; i < quantity; i++ {
		if account.lotsHave > 0 {
			account.freeMoney += price
		} else {
			account.lockedMoney += price
		}
		account.lotsHave--
	}
}

// TestOnHistoricalData тестирует стратегию на исторических данных по заданному инструменту
func TestOnHistoricalData(token string, figi string, daysBeforeNow int, candleInterval sdk.CandleInterval,
	strategyParams StrategyParams, startBalance float64, fee float64, allowMargin bool, charts *Charts) {
	account := &testAccount{
		freeMoney:   startBalance,
		lockedMoney: 0,
		lotsHave:    0,
	}

	*charts.StartBalance = startBalance

	client := RestClientV2{token: token, appname: AppName}

	share, err := client.ShareBy(FIGI, "", figi)
	MaybeCrash(err)

	if allowMargin &&
		((share.Instrument.Dshort.Units == "0" && share.Instrument.Dshort.Nano == 0) ||
			(share.Instrument.Dlong.Units == "0" && share.Instrument.Dlong.Nano == 0)) {
		log.Fatalf("can't margin-trade %v (%v)", share.Instrument.Ticker, figi)
	}

	log.Printf("\n"+
		"instrument: %v (%v)\n"+
		"start balance: %+v",
		share.Instrument.Ticker, figi, account.freeMoney)

	candles := getHistoricalCandles(&client, figi, daysBeforeNow, candleInterval)
	*charts.Candles = append(*charts.Candles, candles[:strategyParams.Window]...)

	for i := strategyParams.Window; i < len(candles); i++ {
		*charts.Candles = append(*charts.Candles, candles[i])

		tradeSignal := GetTradeSignal(
			strategyParams,
			true,
			candles[i],
			true,
			charts,
		)

		if tradeSignal != nil {
			var maxDealValue float64
			var lots int
			switch tradeSignal.Direction {
			case sdk.BUY:
				if allowMargin {
					liquidPortfolio := account.freeMoney + float64(account.lotsHave)*candles[i].ClosePrice*float64(share.Instrument.Lot)
					startMargin := float64(account.lotsHave) * candles[i].ClosePrice * float64(share.Instrument.Lot) * share.Instrument.Dlong.ToFloat()
					maxDealValue = (liquidPortfolio - startMargin) / share.Instrument.Dlong.ToFloat()
				} else {
					maxDealValue = account.freeMoney
				}
				lots = int(maxDealValue / (candles[i].ClosePrice * float64(share.Instrument.Lot)))
				if lots == 0 {
					continue
				}
				testBuy(candles[i].ClosePrice*float64(share.Instrument.Lot)*(1+fee), lots, account)
				break
			case sdk.SELL:
				if allowMargin {
					var liquidPortfolio, startMargin float64
					if account.lotsHave >= 0 { // TODO: добавить маржин-колл
						liquidPortfolio = account.freeMoney + float64(account.lotsHave)*candles[i].ClosePrice*float64(share.Instrument.Lot)
						startMargin = float64(account.lotsHave) * candles[i].ClosePrice * float64(share.Instrument.Lot) * share.Instrument.Dlong.ToFloat()
					} else {
						liquidPortfolio = account.freeMoney + account.lockedMoney + float64(account.lotsHave)*candles[i].ClosePrice*float64(share.Instrument.Lot)
						startMargin = math.Abs(float64(account.lotsHave)) * candles[i].ClosePrice * float64(share.Instrument.Lot) * share.Instrument.Dshort.ToFloat()
					}
					maxDealValue = (liquidPortfolio - startMargin) / share.Instrument.Dshort.ToFloat()
					lots = int(maxDealValue / (candles[i].ClosePrice * float64(share.Instrument.Lot)))
				} else {
					lots = account.lotsHave
				}
				if lots == 0 {
					continue
				}
				testSell(candles[i].ClosePrice*float64(share.Instrument.Lot)*(1-fee), lots, account)
				break
			}
			*charts.Flags = append(*charts.Flags, make([]ChartsTradeFlag, 0))
			(*charts.Flags)[len(*charts.Flags)-1] = append((*charts.Flags)[len(*charts.Flags)-1],
				ChartsTradeFlag{
					Direction:   tradeSignal.Direction,
					Price:       candles[i].ClosePrice,
					Quantity:    lots * share.Instrument.Lot,
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
				Direction:   sdk.BUY,
				Price:       candles[len(candles)-1].ClosePrice,
				Quantity:    -account.lotsHave,
				CandleIndex: len(candles) - 1,
			},
		)
		testBuy(candles[len(candles)-1].ClosePrice, -account.lotsHave, account)
		*charts.BalanceHistory = append(*charts.BalanceHistory, account.freeMoney)
	} else if account.lotsHave > 0 {
		log.Printf("!!! WARNING: force closing longs")
		*charts.Flags = append(*charts.Flags, make([]ChartsTradeFlag, 0))
		(*charts.Flags)[len(*charts.Flags)-1] = append((*charts.Flags)[len(*charts.Flags)-1],
			ChartsTradeFlag{
				Direction:   sdk.SELL,
				Price:       candles[len(candles)-1].ClosePrice,
				Quantity:    account.lotsHave,
				CandleIndex: len(candles) - 1,
			},
		)
		testSell(candles[len(candles)-1].ClosePrice, account.lotsHave, account)
		*charts.BalanceHistory = append(*charts.BalanceHistory, account.freeMoney)
	}
	log.Printf("\n"+
		"TEST RESULTS\n"+
		"Balance: %+v (%+v%%)",
		account.freeMoney, (account.freeMoney-startBalance)/startBalance*100)
}
