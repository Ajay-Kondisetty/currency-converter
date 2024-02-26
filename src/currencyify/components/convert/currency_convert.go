package convert

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

type CurrencyConvertComponent struct {
	components.BaseComponent
}

type CurrencyConverter interface {
	ConvertCurrency(*CurrencyConverterForm) (*CurrencyConverterResponse, error)

	GetCurrencyConverterForm() *CurrencyConverterForm
	GetCurrencyConverterAppError() *utils.AppError
	SetCurrencyConverterAppError(int, error)
}

type CurrencyConverterForm struct {
	SourceCurrency string  `json:"source_currency"`
	TargetCurrency string  `json:"target_currency"`
	Amount         float64 `json:"amount"`
}

type CurrencyConverterResponse struct {
	CurrencyConverterForm
	ConvertedAmount float64 `json:"converted_amount"`
}

var currencyCodesMap map[string]int

type Currency struct {
	CurrencyExchangeRate string
	LastUpdateTime       time.Time
}

// ConvertCurrency is used to convert the given amount from source currency to target currency. If data not found in cache then it will hit external APIs to fetch the conversion rates.
// It returns the converted data and error.
func (ccc *CurrencyConvertComponent) ConvertCurrency(form *CurrencyConverterForm) (*CurrencyConverterResponse, error) {
	resp := new(CurrencyConverterResponse)
	var err error
	if err = form.Valid(); err != nil {
		ccc.AppError = &utils.AppError{
			Status: http.StatusBadRequest,
			Error:  err,
		}

		return nil, err
	}

	if sourceCurrencyRate, err := ccc.getCurrencyRate(form.SourceCurrency); err != nil {
		ccc.SetCurrencyConverterAppError(http.StatusInternalServerError, err)
		return nil, err
	} else if targetCurrencyRate, err := ccc.getCurrencyRate(form.TargetCurrency); err != nil {
		ccc.SetCurrencyConverterAppError(http.StatusInternalServerError, err)
		return nil, err
	} else {
		resp.CurrencyConverterForm = *form
		amountInUSD := form.Amount / sourceCurrencyRate
		resp.ConvertedAmount = amountInUSD * targetCurrencyRate
	}

	return resp, nil
}

func (ccc *CurrencyConvertComponent) getCurrencyRate(currencyCode string) (float64, error) {
	data := new(Currency)

	if !isDataInCache(ccc.RedisConn, currencyCode, data) {
		if resp, err := fetchCurrencyExchangeRate(ccc.ReqCtx, currencyCode); err != nil {
			return 0.0, err
		} else if err = processCurrencyExchangeRate(currencyCode, resp, data); err != nil {
			return 0.0, err
		} else {
			cacheData(ccc.RedisConn, currencyCode, data)
		}
	}

	currencyExchangeRate, err := strconv.ParseFloat(data.CurrencyExchangeRate, 64)
	if err != nil {
		return 0.0, err
	}

	return currencyExchangeRate, nil
}

func fetchCurrencyExchangeRate(reqCtx context.Context, currencyCode string) (utils.Data, error) {
	url := fmt.Sprintf("%v", constants.FX_RATES_API_URL)
	reqHeaders := map[string]string{"Content-Type": "application/json"}
	params := map[string]string{
		"base":       "USD",
		"currencies": currencyCode,
		"resolution": "1m",
		"amount":     "1",
		"format":     "json",
		"places":     "6",
	}
	var resp interface{}
	var err error
	if resp, err = utils.GetAPIResponse(reqCtx, "GetLatestCurrencyRate", url, http.MethodGet, nil, params, reqHeaders); err != nil {
		return nil, err
	}
	caMap, _ := resp.(map[string]interface{})

	log.Printf("fetched latest currency rate for: %s", currencyCode)

	return caMap, nil
}

func processCurrencyExchangeRate(currencyCode string, resp utils.Data, data *Currency) error {
	if rates, ok := resp["rates"].(map[string]interface{}); ok {
		if rate, ok := rates[currencyCode]; ok {
			switch rate.(type) {
			case float64:
				data.CurrencyExchangeRate = strconv.FormatFloat(rate.(float64), 'f', -1, 64)
			default:
				data.CurrencyExchangeRate = rate.(string)
			}

			data.LastUpdateTime, _ = time.Parse(time.RFC3339, resp["date"].(string))
		} else {
			return errors.New("received empty rates data from vendor API. Please check input params")
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

// GetCurrencyConverterForm is used to create a new currency converter form instance.
// It returns currency converter form instance.
func (ccc *CurrencyConvertComponent) GetCurrencyConverterForm() *CurrencyConverterForm {
	return new(CurrencyConverterForm)
}

// GetCurrencyConverterAppError is used to retrieve app error from the currency converter component.
// It returns app error of the component.
func (ccc *CurrencyConvertComponent) GetCurrencyConverterAppError() *utils.AppError {
	return ccc.AppError
}

// SetCurrencyConverterAppError is used to set the app error for the currency converter component.
func (ccc *CurrencyConvertComponent) SetCurrencyConverterAppError(status int, err error) {
	ccc.AppError = &utils.AppError{
		Status: status,
		Error:  err,
	}
}

// Valid validates and sanitizes the currency converter form.
func (f *CurrencyConverterForm) Valid() error {
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

	if f.SourceCurrency == "" {
		addErrMsg("`source_currency` parameter is required")
	} else if !checkCurrency(f.SourceCurrency) {
		addErrMsg("`source_currency` not found in our database. Please check the `source_currency` input param, it should be a valid international-standard 3-letter ISO currency code")
	} else {
		f.SourceCurrency = strings.ToUpper(f.SourceCurrency)
	}

	if f.TargetCurrency == "" {
		addErrMsg("`target_currency` parameter is required")
	} else if !checkCurrency(f.TargetCurrency) {
		addErrMsg("`target_currency` not found in our database. Please check the `target_currency` input param, it should be a valid international-standard 3-letter ISO currency code")
	} else {
		f.TargetCurrency = strings.ToUpper(f.TargetCurrency)
	}

	if f.Amount == 0.0 {
		addErrMsg("`amount` parameter is required")
	}

	p := bluemonday.UGCPolicy()
	f.SourceCurrency = p.Sanitize(f.SourceCurrency)
	f.TargetCurrency = p.Sanitize(f.TargetCurrency)

	if errMsg != "" {
		return errors.New(errMsg)
	}

	return nil
}

func init() {
	components.ComponentMap["CurrencyConvert"] = func(bc *components.BaseComponent) interface{} {
		c := &CurrencyConvertComponent{BaseComponent: *bc}

		return CurrencyConverter(c)
	}
}
