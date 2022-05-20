package main

import (
	"fmt"
	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"log"
	"net/http"
	"time"
)

// Charts - стркуктура, хранящая данные для графиков ECharts
type Charts struct {
	Candles        *[]sdk.Candle
	Intervals      *[][]float64
	Flags          *[][]ChartsTradeFlag
	BalanceHistory *[]float64
	StartBalance   *float64
	TestMode       *bool
}

// ChartsTradeFlag - структура, описывающая отметку о торговом сигнале на графике
type ChartsTradeFlag struct {
	Direction   sdk.OperationType
	Price       float64
	Quantity    int
	CandleIndex int
}

// HandleTradingChart отвечает за обработку запросов к основному торговому графику.
// Эта функция строит сводный график цен, интервалов BollingerBands и торговых сигналов
func (c *Charts) HandleTradingChart(w http.ResponseWriter, _ *http.Request) {
	kline := charts.NewKLine()
	kline.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "1920px",
			Height: "900px",
			Theme:  types.ThemeInfographic,
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Scale: true,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Start:      100 - 50/float32(len(*c.Candles))*100,
			End:        100,
			Throttle:   16.666,
			XAxisIndex: []int{0},
			Type:       "inside",
		}))

	klineX := make([]time.Time, 0)
	klineY := make([]opts.KlineData, 0)
	scatterYBuy := make([]opts.ScatterData, 0)
	scatterYSell := make([]opts.ScatterData, 0)
	lineYLower := make([]opts.LineData, 0)
	lineYUpper := make([]opts.LineData, 0)
	flagContainerIndex := 0
	for i, candle := range *c.Candles {
		klineX = append(klineX, candle.TS)
		klineY = append(klineY, opts.KlineData{Value: []float64{candle.OpenPrice, candle.ClosePrice, candle.LowPrice, candle.HighPrice}})

		if flagContainerIndex < len(*c.Flags) && i == (*c.Flags)[flagContainerIndex][0].CandleIndex {
			buyAvgPrice, sellAvgPrice := 0.0, 0.0
			buyQuantity, sellQuantity := 0, 0
			for _, flag := range (*c.Flags)[flagContainerIndex] {
				switch flag.Direction {
				case sdk.BUY:
					buyAvgPrice += flag.Price
					if buyQuantity > 0 {
						buyAvgPrice /= 2
					}
					buyQuantity += flag.Quantity
					break
				case sdk.SELL:
					sellAvgPrice += flag.Price
					if sellQuantity > 0 {
						sellAvgPrice /= 2
					}
					sellQuantity += flag.Quantity
					break
				}
			}
			if buyQuantity > 0 {
				scatterYBuy = append(scatterYBuy, opts.ScatterData{
					Value:        buyAvgPrice,
					Symbol:       "triangle",
					SymbolSize:   20,
					SymbolRotate: 0,
					Name:         fmt.Sprintf("buy %d for avg.", buyQuantity),
					YAxisIndex:   0,
				})
			} else {
				scatterYBuy = append(scatterYBuy, opts.ScatterData{
					SymbolSize: 0,
				})
			}
			if sellQuantity > 0 {
				scatterYSell = append(scatterYSell, opts.ScatterData{
					Value:        sellAvgPrice,
					Symbol:       "triangle",
					SymbolSize:   20,
					SymbolRotate: 180,
					Name:         fmt.Sprintf("sell %d for avg.", sellQuantity),
					YAxisIndex:   1,
				})
			} else {
				scatterYSell = append(scatterYSell, opts.ScatterData{
					SymbolSize: 0,
				})
			}

			flagContainerIndex++
		} else {
			scatterYBuy = append(scatterYBuy, opts.ScatterData{
				SymbolSize: 0,
			})
			scatterYSell = append(scatterYSell, opts.ScatterData{
				SymbolSize: 0,
			})
		}
		if i-(len(*c.Candles)-len(*c.Intervals)) >= 0 ||
			(*c.TestMode && i-(len(*c.Candles)-len(*c.Intervals)) >= 1) {
			lineYLower = append(lineYLower, opts.LineData{
				Value: (*c.Intervals)[i-(len(*c.Candles)-len(*c.Intervals))][0],
			})
			lineYUpper = append(lineYUpper, opts.LineData{
				Value: (*c.Intervals)[i-(len(*c.Candles)-len(*c.Intervals))][1],
			})
		} else {
			lineYLower = append(lineYLower, opts.LineData{SymbolSize: 0})
			lineYUpper = append(lineYUpper, opts.LineData{SymbolSize: 0})
		}
	}
	kline.SetXAxis(klineX).
		AddSeries("Price", klineY).
		SetSeriesOptions(
			charts.WithItemStyleOpts(opts.ItemStyle{
				Color:        "#00000000",
				Color0:       "#00000000",
				BorderColor:  "#00FF00",
				BorderColor0: "#FF0000",
			}),
		)

	scatterBuy := charts.NewScatter()
	scatterSell := charts.NewScatter()
	scatterBuy.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "1920px",
			Height: "900px",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Scale: true,
		}),
	)
	scatterSell.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "1920px",
			Height: "900px",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Scale: true,
		}),
	)

	scatterBuy.SetXAxis(klineX).
		AddSeries("Buy flags", scatterYBuy).
		SetSeriesOptions(
			charts.WithItemStyleOpts(opts.ItemStyle{
				Color:       "#55FF55",
				BorderColor: "#55FF55",
			}),
			charts.WithLabelOpts(opts.Label{
				Show:      true,
				Color:     "white",
				Position:  "bottom",
				Formatter: "{b0} {c0}",
			}),
		)
	scatterSell.SetXAxis(klineX).
		AddSeries("Sell flags", scatterYSell).
		SetSeriesOptions(
			charts.WithItemStyleOpts(opts.ItemStyle{
				Color:       "#FF5555",
				BorderColor: "#FF5555",
			}),
			charts.WithLabelOpts(opts.Label{
				Show:      true,
				Color:     "white",
				Position:  "top",
				Formatter: "{b0} {c0}",
			}),
		)
	kline.Overlap(scatterBuy)
	kline.Overlap(scatterSell)

	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "1920px",
			Height: "900px",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Scale: true,
		}),
	)
	line.SetXAxis(klineX).
		AddSeries("Bollinger (lower)", lineYLower).
		AddSeries("Bollinger (upper)", lineYUpper)
	kline.Overlap(line)

	err := kline.Render(w)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = w.Write([]byte("<style> body { background-color: black; }</style>"))
	if err != nil {
		log.Fatalln(err)
	}
	if !*c.TestMode {
		// будем перезагружать страницу каждые 30 секунд
		_, err = w.Write([]byte("<meta http-equiv=\"refresh\" content=\"30\" />"))
		if err != nil {
			log.Fatalln(err)
		}
	}
}

// HandleBalanceChart отвечает за обработку запросов к графику истории баланса
func (c *Charts) HandleBalanceChart(w http.ResponseWriter, _ *http.Request) {
	bar := charts.NewBar()
	line := charts.NewLine()
	barX := make([]int, 0)
	barY := make([]opts.BarData, 0)
	lineY := make([]opts.LineData, 0)
	for i, balance := range *c.BalanceHistory {
		barX = append(barX, i)
		barY = append(barY, opts.BarData{Value: balance})
		lineY = append(lineY, opts.LineData{Value: c.StartBalance})
	}

	bar.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "1920px",
			Height: "900px",
			Theme:  types.ThemeInfographic,
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Scale: true,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Start:      0,
			End:        100,
			Throttle:   16.666,
			XAxisIndex: []int{0},
		}),
	)
	bar.SetXAxis(barX).AddSeries("Balance", barY)

	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "1920px",
			Height: "900px",
			Theme:  types.ThemeInfographic,
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Scale: true,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Start:      0,
			End:        100,
			Throttle:   16.666,
			XAxisIndex: []int{0},
		}),
	)
	line.SetXAxis(barX).AddSeries("Start balance", lineY)

	_, err := w.Write([]byte("<style> body { background-color: black; }</style>"))
	if err != nil {
		log.Fatalln(err)
	}
	bar.Overlap(line)
	err = bar.Render(w)
	if err != nil {
		log.Fatalln(err)
	}
}
