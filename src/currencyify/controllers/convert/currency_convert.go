package convert

import (
	"encoding/json"
	"log"
	"net/http"

	"currencyify/components/convert"
	"currencyify/controllers"
	"currencyify/utils"
)

type CurrencyConvertController struct {
	controllers.BaseController
	Component convert.CurrencyConverter
}

// UpdateComponent is used to update the component object.
func (c *CurrencyConvertController) UpdateComponent(component interface{}) {
	c.Component, _ = component.(convert.CurrencyConverter)
}

func (c *CurrencyConvertController) ConvertCurrency() {
	var d *convert.CurrencyConverterResponse
	var err error
	var status int

	form := c.Component.GetCurrencyConverterForm()

	if err = json.Unmarshal(c.GetRequestBody(), form); err != nil {
		status = http.StatusInternalServerError
	} else if d, err = c.Component.ConvertCurrency(form); err != nil {
		status = c.Component.GetCurrencyConverterAppError().Status
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
