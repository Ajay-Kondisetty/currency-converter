package main

import (
	"log"

	"currencyify/constants"
	"currencyify/routers"

	_ "github.com/beego/beego/v2/core/config/yaml"
	"github.com/beego/beego/v2/server/web"
	"github.com/joho/godotenv"
)

func main() {
	// Generated using http://patorjk.com/software/taag/#p=display&f=Graffiti
	log.Printf(`
                                                               .__   _____        
  ____   __ __ _______ _______   ____    ____    ____  ___.__.|__|_/ ____\___.__.
_/ ___\ |  |  \\_  __ \\_  __ \_/ __ \  /    \ _/ ___\<   |  ||  |\   __\<   |  |
\  \___ |  |  / |  | \/ |  | \/\  ___/ |   |  \\  \___ \___  ||  | |  |   \___  |
 \___  >|____/  |__|    |__|    \___  >|___|  / \___  >/ ____||__| |__|   / ____|
     \/                             \/      \/      \/ \/                 \/      
	`)
	web.BConfig.Log.AccessLogs = true
	web.Run()
}

func init() {
	// Load app conf.
	appConfigFile := "conf/local.app.yaml"
	if err := web.LoadAppConfig("yaml", appConfigFile); err != nil {
		log.Fatal("Error loading app config: ", err)
	} else {
		log.Printf("Loaded app config: %v", appConfigFile)
	}

	// Load env vars and init const from envs
	if err := godotenv.Load("local_env"); err != nil {
		log.Fatal("Error loading env variables: ", err)
	}
	constants.InitConstantsVars()

	// Init routes
	routers.InitRoutes()
}
