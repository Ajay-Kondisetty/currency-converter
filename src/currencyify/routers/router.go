package routers

import (
	"fmt"

	"currencyify/constants"
	"currencyify/controllers/convert"
	"currencyify/controllers/exchange_rate"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
)

func InitRoutes() {
	ns := web.NewNamespace(fmt.Sprintf("/%v", constants.API_PATH),
		web.NSGet("/healthcheck", func(ctx *context.Context) {
			_ = ctx.Output.Body([]byte("i am alive"))
		}),

		web.NSNamespace("/convert",
			web.NSNamespace(
				"/currency-convert",
				web.NSInclude(
					&convert.CurrencyConvertController{},
				),
			),
		),

		web.NSNamespace("/exchange-rate",
			web.NSNamespace(
				"/currency-exchange-rate",
				web.NSInclude(
					&exchange_rate.CurrencyExchangeRateController{},
				),
			),
		),
	)

	web.AddNamespace(ns)
}
