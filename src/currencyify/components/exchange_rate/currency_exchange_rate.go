package exchange_rate

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"currencyify/components"
	"currencyify/constants"
	"currencyify/utils"

	"github.com/gomodule/redigo/redis"
	"github.com/microcosm-cc/bluemonday"
)

type CurrencyExchangeRateComponent struct {
	components.BaseComponent
}

type CurrencyExchangeRate interface {
	GetCurrencyExchangeRate(*CurrencyExchangeRateForm) (*CurrencyExchangeRateResponse, error)

	GetCurrencyExchangeRateForm() *CurrencyExchangeRateForm
	GetCurrencyExchangeRateAppError() *utils.AppError
	SetCurrencyExchangeRateAppError(int, error)
}

type CurrencyExchangeRateForm struct {
	BaseCurrency     string   `json:"base_currency"`
	TargetCurrencies []string `json:"target_currencies"`
}

type CurrencyExchangeRateResponse struct {
	BaseCurrency  string              `json:"base_currency"`
	ExchangeRates map[string]Currency `json:"exchange_rates"`
}

var currencyCodesMap map[string]int

type Currency struct {
	CurrencyExchangeRate string    `json:"currency_exchange_rate"`
	LastUpdateTime       time.Time `json:"last_update_time"`
}

// GetCurrencyExchangeRate is used to get the currency exchange rates of the given currency codes. If data not found in cache then it will fetch from third party API.
// It returns the currency exchange rates of given currency codes and error.
func (cec *CurrencyExchangeRateComponent) GetCurrencyExchangeRate(form *CurrencyExchangeRateForm) (*CurrencyExchangeRateResponse, error) {
	resp := new(CurrencyExchangeRateResponse)
	var err error
	if err = form.Valid(); err != nil {
		cec.AppError = &utils.AppError{
			Status: http.StatusBadRequest,
			Error:  err,
		}

		return nil, err
	}

	if currencyExchangeRates, err := cec.getCurrencyExchangeRate(form); err != nil {
		cec.SetCurrencyExchangeRateAppError(http.StatusInternalServerError, err)
		return nil, err
	} else {
		resp.ExchangeRates = currencyExchangeRates
		resp.BaseCurrency = form.BaseCurrency
	}

	return resp, nil
}

func (cec *CurrencyExchangeRateComponent) getCurrencyExchangeRate(form *CurrencyExchangeRateForm) (map[string]Currency, error) {
	data := new(Currency)
	result := make(map[string]Currency)
	pendingCurrencyCodes := make([]string, 0)
	for _, currencyCode := range form.TargetCurrencies {
		if isDataInCache(cec.RedisConn, fmt.Sprintf("%s-%s", form.BaseCurrency, currencyCode), data) {
			result[currencyCode] = *data
		} else {
			pendingCurrencyCodes = append(pendingCurrencyCodes, currencyCode)
		}
	}

	if len(pendingCurrencyCodes) == 0 {
		return result, nil
	}

	if resp, err := fetchCurrencyExchangeRate(cec.ReqCtx, form.BaseCurrency, pendingCurrencyCodes); err != nil {
		return result, err
	} else if err = processCurrencyExchangeRate(cec.RedisConn, form.BaseCurrency, resp, result); err != nil {
		return result, err
	}

	return result, nil
}

func fetchCurrencyExchangeRate(reqCtx context.Context, baseCurrencyCode string, pendingCurrencyCodes []string) (utils.Data, error) {
	url := fmt.Sprintf("%v", constants.FX_RATES_API_URL)
	reqHeaders := map[string]string{"Content-Type": "application/json"}
	currencyCodes := strings.Join(pendingCurrencyCodes, ",")
	params := map[string]string{
		"base":       baseCurrencyCode,
		"currencies": currencyCodes,
		"resolution": "1m",
		"amount":     "1",
		"format":     "json",
		"places":     "6",
	}
	var resp interface{}
	var err error
	if resp, err = utils.GetAPIResponse(reqCtx, "GetLatestCurrencyExchangeRates", url, http.MethodGet, nil, params, reqHeaders); err != nil {
		return nil, err
	}
	caMap, _ := resp.(map[string]interface{})

	log.Printf("fetched latest currency exchange rates for: %s", currencyCodes)

	return caMap, nil
}

func processCurrencyExchangeRate(redisConn redis.Conn, baseCurrencyCode string, resp utils.Data, result map[string]Currency) error {
	if rates, ok := resp["rates"].(map[string]interface{}); ok {
		parsedTime, _ := time.Parse(time.RFC3339, resp["date"].(string))
		for currencyCode, rate := range rates {
			data := new(Currency)
			switch rate.(type) {
			case float64:
				data.CurrencyExchangeRate = strconv.FormatFloat(rate.(float64), 'f', -1, 64)
			default:
				data.CurrencyExchangeRate = rate.(string)
			}

			data.LastUpdateTime = parsedTime
			cacheData(redisConn, fmt.Sprintf("%s-%s", baseCurrencyCode, currencyCode), data)
			result[currencyCode] = *data
		}
	} else {
		return errors.New("error while processing exchange rates of vendor API data")
	}

	log.Printf("processed exchange rates data")

	return nil
}

func isDataInCache(redisConn redis.Conn, currencyCode string, cacheData *Currency) bool {
	if redisConn == nil {
		log.Printf("redis conn not found to fetch data from cache")
		return false
	}
	dataStr, err := utils.GetData(redisConn, currencyCode)
	if err != nil {
		log.Printf("data not found in cache for: %v", currencyCode)
	} else if dataBytes, err := base64.StdEncoding.DecodeString(dataStr); err != nil {
		log.Printf("error while decoding base64 cache data")
	} else if err := json.Unmarshal(dataBytes, cacheData); err != nil {
		log.Printf("error unmarshaling cache data")
	} else {
		log.Printf("data found in cache for: %v", currencyCode)
		return true
	}

	return false
}

func cacheData(redisConn redis.Conn, currencyCode string, cacheData *Currency) {
	if redisConn == nil {
		log.Printf("redis conn not found to store in cache")
		return
	}
	if respBytes, err := json.Marshal(cacheData); err != nil {
		log.Printf("error marshaling data to store in cache")
	} else if respStr := base64.StdEncoding.EncodeToString(respBytes); respStr != "" {
		ttl, _ := strconv.Atoi(constants.REDIS_DEFAULT_EXPIRY)
		if status, err := utils.SetData(redisConn, currencyCode, respStr, ttl); err != nil || !status {
			log.Printf("error setting data in cache")
		} else {
			log.Printf("data succesfully stored in cache")
		}
	}
}

// GetCurrencyExchangeRateForm is used to create a new currency exchange rate form instance.
// It returns currency exchange rate form instance.
func (cec *CurrencyExchangeRateComponent) GetCurrencyExchangeRateForm() *CurrencyExchangeRateForm {
	return new(CurrencyExchangeRateForm)
}

// GetCurrencyExchangeRateAppError is used to retrieve app error from the currency exchange rate component.
// It returns app error of the component.
func (cec *CurrencyExchangeRateComponent) GetCurrencyExchangeRateAppError() *utils.AppError {
	return cec.AppError
}

// SetCurrencyExchangeRateAppError is used to set the app error for the currency exchange rate component.
func (cec *CurrencyExchangeRateComponent) SetCurrencyExchangeRateAppError(status int, err error) {
	cec.AppError = &utils.AppError{
		Status: status,
		Error:  err,
	}
}

// Valid validates and sanitizes the currency converter form.
func (f *CurrencyExchangeRateForm) Valid() error {
	currencyCodesStr, err := os.ReadFile(constants.CURRENCY_CODES_JSON_FILE_NAME)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(currencyCodesStr, &currencyCodesMap); err != nil {
		return err
	}

	errMsg := ""
	addErrMsg := func(msg string) {
		if errMsg != "" {
			errMsg += "\n"
		}
		errMsg += msg
	}

	checkCurrency := func(currency string) bool {
		_, ok := currencyCodesMap[strings.ToLower(currency)]
		return ok
	}

	if f.BaseCurrency == "" {
		addErrMsg("`base_currency` parameter is required")
	} else if !checkCurrency(f.BaseCurrency) {
		addErrMsg("`base_currency` not found in our database. Please check the `base_currency` input param, it should be a valid international-standard 3-letter ISO currency code")
	} else {
		f.BaseCurrency = strings.ToUpper(f.BaseCurrency)
	}

	if len(f.TargetCurrencies) == 0 {
		addErrMsg("`target_currencies` parameter is required")
	} else {
		for index, targetCurrency := range f.TargetCurrencies {
			if !checkCurrency(targetCurrency) {
				addErrMsg(fmt.Sprintf("`target_currency` (%s) not found in our database. Please check the `target_currencies` input param, it should be a valid international-standard 3-letter ISO currency code", targetCurrency))
			} else {
				f.TargetCurrencies[index] = strings.ToUpper(targetCurrency)
			}
		}
	}

	p := bluemonday.UGCPolicy()
	f.BaseCurrency = p.Sanitize(f.BaseCurrency)

	if errMsg != "" {
		return errors.New(errMsg)
	}

	return nil
}

func init() {
	components.ComponentMap["CurrencyExchangeRate"] = func(bc *components.BaseComponent) interface{} {
		c := &CurrencyExchangeRateComponent{BaseComponent: *bc}

		return CurrencyExchangeRate(c)
	}
}
