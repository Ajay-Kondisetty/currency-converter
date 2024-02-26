package convert

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"currencyify/components"
	"currencyify/constants"
	"currencyify/utils"

	"github.com/stretchr/testify/assert"
)

func TestCurrencyConverterForm_Valid(t *testing.T) {
	constants.CURRENCY_CODES_JSON_FILE_NAME = "../../currency_codes.json"

	type vars struct {
		form *CurrencyConverterForm
	}

	testCases := []struct {
		name string

		vars vars

		hasErr bool
		err    string
	}{
		{
			name: "should fail when source currency code is empty",
			vars: vars{
				form: &CurrencyConverterForm{
					TargetCurrency: "INR",
					Amount:         10.0,
				},
			},
			hasErr: true,
			err:    "`source_currency` parameter is required",
		},
		{
			name: "should fail when target currency code is empty",
			vars: vars{
				form: &CurrencyConverterForm{
					SourceCurrency: "INR",
					Amount:         10.0,
				},
			},
			hasErr: true,
			err:    "`target_currency` parameter is required",
		},
		{
			name: "should fail when amount is empty",
			vars: vars{
				form: &CurrencyConverterForm{
					SourceCurrency: "USD",
					TargetCurrency: "INR",
				},
			},
			hasErr: true,
			err:    "`amount` parameter is required",
		},
		{
			name: "should fail when source currency code format is not international-standard 3-letter ISO currency code",
			vars: vars{
				form: &CurrencyConverterForm{
					SourceCurrency: "England",
					TargetCurrency: "INR",
					Amount:         10,
				},
			},
			hasErr: true,
			err:    "`source_currency` not found in our database. Please check the `source_currency` input param, it should be a valid international-standard 3-letter ISO currency code",
		},
		{
			name: "should fail when target currency code format is not international-standard 3-letter ISO currency code",
			vars: vars{
				form: &CurrencyConverterForm{
					SourceCurrency: "USD",
					TargetCurrency: "India",
					Amount:         10,
				},
			},
			hasErr: true,
			err:    "`target_currency` not found in our database. Please check the `target_currency` input param, it should be a valid international-standard 3-letter ISO currency code",
		},
		{
			name: "should success to validate the currency converter input form",
			vars: vars{
				form: &CurrencyConverterForm{
					SourceCurrency: "USD",
					TargetCurrency: "INR",
					Amount:         10,
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

func TestCurrencyConvertComponent_GetCurrencyConverterAppError(t *testing.T) {
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
			ccc := &CurrencyConvertComponent{
				BaseComponent: tCase.vars.component,
			}

			// Run test
			got := ccc.GetCurrencyConverterAppError()

			// Assert
			assert.Exactlyf(t, tCase.want, got, "case: %v", tCase)
		})
	}
}

func TestCurrencyConvertComponent_GetCurrencyConverterForm(t *testing.T) {
	type vars struct {
		component components.BaseComponent
	}

	testCases := []struct {
		name string

		vars vars

		want *CurrencyConverterForm
	}{
		{
			name: "should success to fetch the form from the component",
			vars: vars{
				component: components.BaseComponent{
					ReqCtx: context.Background(),
				},
			},
			want: new(CurrencyConverterForm),
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// Setup
			ccc := &CurrencyConvertComponent{
				BaseComponent: tCase.vars.component,
			}

			// Run test
			got := ccc.GetCurrencyConverterForm()

			// Assert
			assert.Exactlyf(t, tCase.want, got, "case: %v", tCase)
		})
	}
}

func TestCurrencyConvertComponent_SetCurrencyConverterAppError(t *testing.T) {
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
			ccc := &CurrencyConvertComponent{
				BaseComponent: tCase.vars.component,
			}

			// Run test
			ccc.SetCurrencyConverterAppError(tCase.vars.status, tCase.vars.err)

			// Assert
			assert.Exactlyf(t, tCase.want, ccc.GetCurrencyConverterAppError(), "case: %v", tCase)
		})
	}
}

func TestCurrencyConvertComponent_ConvertCurrency(t *testing.T) {
	constants.CURRENCY_CODES_JSON_FILE_NAME = "../../currency_codes.json"

	type vars struct {
		component components.BaseComponent

		form *CurrencyConverterForm

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
			name: "should success to convert the given amount from source currency to target currency",
			vars: vars{
				component: components.BaseComponent{
					ReqCtx: context.Background(),
				},
				form: &CurrencyConverterForm{
					SourceCurrency: "USD",
					TargetCurrency: "INR",
					Amount:         100,
				},
				headers: map[string]string{
					"x-mock-api": "default",
				},
			},
			want: `{ "source_currency": "USD", "target_currency": "INR", "amount": 100, "converted_amount": 8277.1291 }`,
		},
		{
			name: "should fail to convert the given amount from source currency to target currency",
			vars: vars{
				component: components.BaseComponent{
					ReqCtx: context.Background(),
				},
				form: &CurrencyConverterForm{
					SourceCurrency: "USD",
					TargetCurrency: "INR",
					Amount:         100,
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
			ccc := &CurrencyConvertComponent{
				BaseComponent: tCase.vars.component,
			}
			ctx := ccc.ReqCtx
			ctx = context.WithValue(ctx, "x-mock-headers", tCase.vars.headers)
			ccc.ReqCtx = ctx

			// Run test
			got, err := ccc.ConvertCurrency(form)

			// Assert
			if tCase.hasErr {
				if assert.Errorf(t, err, "case: %v", tCase) {
					assert.Containsf(t, err.Error(), tCase.err, "case: %v", tCase)
				}
			} else {
				assert.NoErrorf(t, err, "case: %v", tCase)
				tempWant := new(CurrencyConverterResponse)
				_ = json.Unmarshal([]byte(tCase.want), tempWant)
				assert.Equal(t, tempWant, got, "case: %v", tCase)
			}
		})
	}

}
