package exchange_rate

import (
	"encoding/json"
	"log"
	"net/http"

	"currencyify/components/exchange_rate"
	"currencyify/controllers"
	"currencyify/utils"
)

type CurrencyExchangeRateController struct {
	controllers.BaseController
	Component exchange_rate.CurrencyExchangeRate
}

// UpdateComponent is used to update the component object.
func (c *CurrencyExchangeRateController) UpdateComponent(component interface{}) {
	c.Component, _ = component.(exchange_rate.CurrencyExchangeRate)
}

func (c *CurrencyExchangeRateController) GetCurrencyExchangeRate() {
	var d *exchange_rate.CurrencyExchangeRateResponse
	var err error
	var status int

	form := c.Component.GetCurrencyExchangeRateForm()

	if err = json.Unmarshal(c.GetRequestBody(), form); err != nil {
		status = http.StatusInternalServerError
	} else if d, err = c.Component.GetCurrencyExchangeRate(form); err != nil {
		status = c.Component.GetCurrencyExchangeRateAppError().Status
	}

	if err != nil {
		log.Printf("Some error occurred: %v", err)
	} else {
		status = http.StatusOK
	}

	c.Data["json"] = utils.PrepareResponse(d, err, status)
	c.AddHeaders(status, map[string]bool{"no_cache": true})
	_ = c.ServeJSON()
}
