package routers

import (
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context/param"
)

func init() {

	beego.GlobalControllerRouter["currencyify/controllers/convert:CurrencyConvertController"] = append(beego.GlobalControllerRouter["currencyify/controllers/convert:CurrencyConvertController"],
		beego.ControllerComments{
			Method:           "ConvertCurrency",
			Router:           `/`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["currencyify/controllers/exchange_rate:CurrencyExchangeRateController"] = append(beego.GlobalControllerRouter["currencyify/controllers/exchange_rate:CurrencyExchangeRateController"],
		beego.ControllerComments{
			Method:           "GetCurrencyExchangeRate",
			Router:           `/`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})
}
