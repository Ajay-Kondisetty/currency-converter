package constants

import (
	"os"
)

const (
	API_PATH = "api/v1/currencyify"
)

var (
	FX_RATES_API_URL              = ""
	CURRENCY_CODES_JSON_FILE_NAME = ""

	REDIS_HOST           = ""
	REDIS_PORT           = ""
	REDIS_DEFAULT_EXPIRY = ""
)

func InitConstantsVars() {
	FX_RATES_API_URL = os.Getenv("FX_RATES_API_URL")

	CURRENCY_CODES_JSON_FILE_NAME = os.Getenv("CURRENCY_CODES_JSON_FILE_NAME")

	REDIS_HOST = os.Getenv("REDIS_HOST")
	REDIS_PORT = os.Getenv("REDIS_PORT")
	REDIS_DEFAULT_EXPIRY = os.Getenv("REDIS_DEFAULT_EXPIRY")
}
