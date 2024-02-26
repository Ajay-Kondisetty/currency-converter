# currency-converter
A small microservice which exposes a few endpoints to convert currency from source country to destination country and to fetch the exchange rates of the given countries.

## Description

A small microservice has an endpoint which takes source currency code, target currency code, amount as inputs and converts the given amount to the target currency code as per the conversion rate. It also has an endpoint which takes a base currency and list of target currencies and return the currency conversion rates as per the given base currency code.

## Getting Started

### Dependencies

* Docker
* Docker Compose
* Create a file named `local_env` in the `currencyify` folder and the following variables with appropriate values.
* The `source_currency`, `target_currency`, `base_currency`, and `target_currencies` input params should follow international-standard 3-letter ISO currency code.
```
ENVIRONMENT=local

CURRENCY_CODES_JSON_FILE_NAME=currency_codes.json

FX_RATES_API_URL=https://api.fxratesapi.com/latest

# HTTP Request config
HTTP_RESPONSE_HEADER_TIMEOUT=60s

# Redis database
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_DEFAULT_EXPIRY=10800
```

### Installing

* Clone the repo
```
git clone https://github.com/Ajay-Kondisetty/currency-converter
```

### Executing program

* Change to root of the application which has Dockerfile
* Execute docker compose command to spin-up the app(if you are running it on Windows machine then make sure to launch Docker Desktop app first)
```
docker-compose up
```
* The app should be up and ready to handle connections within few seconds

## Authors
Ajay Kondisetty
[@ajaykondisetty](https://www.linkedin.com/in/i-am-ajay/)

## Version History
* 1.0
    * Initial Release

## License

This project is licensed under the MIT License by Ajay Kondisetty - see the LICENSE.md file for details