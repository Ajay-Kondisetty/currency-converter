package exchange_rate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"currencyify/components"
	"currencyify/constants"
	"currencyify/utils"

	"github.com/stretchr/testify/assert"
)

func TestCurrencyExchangeRateForm_Valid(t *testing.T) {
	constants.CURRENCY_CODES_JSON_FILE_NAME = "../../currency_codes.json"

	type vars struct {
		form *CurrencyExchangeRateForm
	}

	testCases := []struct {
		name string

		vars vars

		hasErr bool
		err    string
	}{
		{
			name: "should fail when base currency code is empty",
			vars: vars{
				form: &CurrencyExchangeRateForm{
					TargetCurrencies: []string{"inr", "jpy"},
				},
			},
			hasErr: true,
			err:    "`base_currency` parameter is required",
		},
		{
			name: "should fail when target currencies are empty",
			vars: vars{
				form: &CurrencyExchangeRateForm{
					BaseCurrency: "INR",
				},
			},
			hasErr: true,
			err:    "`target_currencies` parameter is required",
		},
		{
			name: "should fail when base currency code format is not international-standard 3-letter ISO currency code",
			vars: vars{
				form: &CurrencyExchangeRateForm{
					BaseCurrency: "England",
				},
			},
			hasErr: true,
			err:    "`base_currency` not found in our database. Please check the `base_currency` input param, it should be a valid international-standard 3-letter ISO currency code",
		},
		{
			name: "should fail when any of the target currencies code format is not international-standard 3-letter ISO currency code",
			vars: vars{
				form: &CurrencyExchangeRateForm{
					BaseCurrency:     "USD",
					TargetCurrencies: []string{"India", "Jpy"},
				},
			},
			hasErr: true,
			err:    "`target_currency` (India) not found in our database. Please check the `target_currencies` input param, it should be a valid international-standard 3-letter ISO currency code",
		},
		{
			name: "should success to validate the currency exchange rate input form",
			vars: vars{
				form: &CurrencyExchangeRateForm{
					BaseCurrency:     "USD",
					TargetCurrencies: []string{"INR", "Jpy"},
				},
			},
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// Setup
			form := tCase.vars.form

			// Run test
			err := form.Valid()

			// Assert
			if tCase.hasErr {
				if assert.Errorf(t, err, "case: %v", tCase) {
					assert.Containsf(t, err.Error(), tCase.err, "case: %v", tCase)
				}
			} else {
				assert.NoErrorf(t, err, "case: %v", tCase)
			}
		})
	}
}

func TestCurrencyExchangeRateComponent_GetCurrencyExchangeRateAppError(t *testing.T) {
	type vars struct {
		component components.BaseComponent
	}

	testCases := []struct {
		name string

		vars vars

		want *utils.AppError
	}{
		{
			name: "should success to fetch App Error from the component",
			vars: vars{
				component: components.BaseComponent{
					ReqCtx: context.Background(),
					AppError: &utils.AppError{
						Error:  errors.New("some error"),
						Status: http.StatusBadRequest,
					},
				},
			},
			want: &utils.AppError{
				Error:  errors.New("some error"),
				Status: http.StatusBadRequest,
			},
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// Setup
			cec := &CurrencyExchangeRateComponent{
				BaseComponent: tCase.vars.component,
			}

			// Run test
			got := cec.GetCurrencyExchangeRateAppError()

			// Assert
			assert.Exactlyf(t, tCase.want, got, "case: %v", tCase)
		})
	}
}

func TestCurrencyExchangeRateComponent_GetCurrencyExchangeRateForm(t *testing.T) {
	type vars struct {
		component components.BaseComponent
	}

	testCases := []struct {
		name string

		vars vars

		want *CurrencyExchangeRateForm
	}{
		{
			name: "should success to fetch the form from the component",
			vars: vars{
				component: components.BaseComponent{
					ReqCtx: context.Background(),
				},
			},
			want: new(CurrencyExchangeRateForm),
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// Setup
			cec := &CurrencyExchangeRateComponent{
				BaseComponent: tCase.vars.component,
			}

			// Run test
			got := cec.GetCurrencyExchangeRateForm()

			// Assert
			assert.Exactlyf(t, tCase.want, got, "case: %v", tCase)
		})
	}
}

func TestCurrencyExchangeRateComponent_SetCurrencyExchangeRateAppError(t *testing.T) {
	type vars struct {
		component components.BaseComponent
		status    int
		err       error
	}

	testCases := []struct {
		name string

		vars vars

		want *utils.AppError
	}{
		{
			name: "should success to set App Error to the component",
			vars: vars{
				component: components.BaseComponent{
					ReqCtx: context.Background(),
				},
				status: http.StatusInternalServerError,
				err:    errors.New("some error"),
			},
			want: &utils.AppError{
				Error:  errors.New("some error"),
				Status: http.StatusInternalServerError,
			},
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// Setup
			cec := &CurrencyExchangeRateComponent{
				BaseComponent: tCase.vars.component,
			}

			// Run test
			cec.SetCurrencyExchangeRateAppError(tCase.vars.status, tCase.vars.err)

			// Assert
			assert.Exactlyf(t, tCase.want, cec.GetCurrencyExchangeRateAppError(), "case: %v", tCase)
		})
	}
}

func TestCurrencyExchangeRateComponent_GetCurrencyExchangeRate(t *testing.T) {
	constants.CURRENCY_CODES_JSON_FILE_NAME = "../../currency_codes.json"

	type vars struct {
		component components.BaseComponent

		form *CurrencyExchangeRateForm

		headers map[string]string
	}

	testCases := []struct {
		name string

		vars vars

		want   string
		hasErr bool
		err    string
	}{
		{
			name: "should success to fetch the currency exchange rates of the given currency codes in accordance with base currency",
			vars: vars{
				component: components.BaseComponent{
					ReqCtx: context.Background(),
				},
				form: &CurrencyExchangeRateForm{
					BaseCurrency:     "USD",
					TargetCurrencies: []string{"INR", "JPY"},
				},
				headers: map[string]string{
					"x-mock-api": "default",
				},
			},
			want: ` { "base_currency": "USD", "exchange_rates": { "INR": { "currency_exchange_rate": "82.771291", "last_update_time": "2024-02-26T12:04:00Z" }, "JPY": { "currency_exchange_rate": "150.608807", "last_update_time": "2024-02-26T12:04:00Z" } } }`,
		},
		{
			name: "should fail to fetch the currency exchange rates of the given currency codes in accordance with base currency",
			vars: vars{
				component: components.BaseComponent{
					ReqCtx: context.Background(),
				},
				form: &CurrencyExchangeRateForm{
					BaseCurrency:     "USD",
					TargetCurrencies: []string{"INR", "JPY"},
				},
				headers: map[string]string{
					"x-mock-api": "error_response",
				},
			},
			hasErr: true,
			err:    "error",
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// Setup
			form := tCase.vars.form
			ttc := &CurrencyExchangeRateComponent{
				BaseComponent: tCase.vars.component,
			}
			ctx := ttc.ReqCtx
			ctx = context.WithValue(ctx, "x-mock-headers", tCase.vars.headers)
			ttc.ReqCtx = ctx

			// Run test
			got, err := ttc.GetCurrencyExchangeRate(form)

			// Assert
			if tCase.hasErr {
				if assert.Errorf(t, err, "case: %v", tCase) {
					assert.Containsf(t, err.Error(), tCase.err, "case: %v", tCase)
				}
			} else {
				assert.NoErrorf(t, err, "case: %v", tCase)
				tempWant := new(CurrencyExchangeRateResponse)
				_ = json.Unmarshal([]byte(tCase.want), tempWant)
				fmt.Println("@@@@@@@@@@@@@:- ", tempWant)
				assert.Equal(t, tempWant, got, "case: %v", tCase)
			}
		})
	}

}
