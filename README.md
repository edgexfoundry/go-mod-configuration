# go-mod-configuration
Configuration client library for use by Go implementation of EdgeX micro services.  This project contains the abstract Configuration API and an implementation for Consul. The API initializes a connection to the Configuration service and push/pull configuration values to/from the Configuration service.

### What is this repository for? ###
* Initialize connection to a Configuration service
* Push a service's configuration in to the Configuration
* Pull service's configuration from the Configuration service into its configuration struct
* Listen for configuration updates

### Installation ###
* Make sure you have Go Modules enabled, i.e. have an initialized  go.mod file 
* If your code is in your GOPATH then make sure ```GO111MODULE=on``` is set
* Run ```go get github.com/edgexfoundry/go-mod-configuration```
    * This will add the go-mod-configuration to the go.mod file and download it into the module cache
    
### How to Use ###
This library is used by Go programs for interacting with the Configuration service (i.e. Consul) and requires that a Configuration service be running somewhere that the Configuration Client can connect.  The types.ServiceConfig struct is used to specify the service implementation details :

```go
type ServiceConfig struct {
	Protocol string
	Host string
	Port int
	Type string
	BasePath string
	AccessToken string
}
```

The following code snippets demonstrate how a service uses this Configuration module to store and load configuration, listen to for configuration updates.

This code snippet shows how to connect to the Configuration service, store and load the service's configuration from the Configuration service.  
```
func initializeConfiguration(useConfigService bool, useProfile string) (*ConfigurationStruct, error) {
	configuration := &ConfigurationStruct{}
	err := config.LoadFromFile(useProfile, configuration)
	if err != nil {
		return nil, err
	}

    if useConfigService {
        serviceConfig := types.Config{
            Host:            conf.Configuration.Host,
            Port:            conf.Configuration.Port,
            Type:            conf.Configuration.Type,
            BasePath:        internal.ConfigStem + internal.MyServiceKey,
            AccessToken:     <AccessTokenLoadedFromSecretFile>,
        }

        ConfigClient, err = configuration.NewConfigurationClient(serviceConfig)
    	if err != nil {
    		return fmt.Errorf("connection to Configuration service could not be made: %v", err.Error())
    	}


		hasConfig, err := ConfigClient.HasConfiguration()
		if hasConfig {
            // Get the service's configuration from the Configuration service
            rawConfig, err := ConfigClient.GetConfiguration(configuration)
            if err != nil {
                return fmt.Errorf("could not get configuration from Configuration: %v", err.Error())
            }

            actual, ok := rawConfig.(*ConfigurationStruct)
            if !ok {
                return fmt.Errorf("configuration from Configuration failed type check")
            }

            *configuration = actual
        } else {
            err = ConfigClient.PutConfiguration(configuration, true)
			if err != nil {
				return fmt.Errorf("could not push configuration into Configuration service: %v", err)
			}
        }
        
        // Run as go func so doesn't block
        go listenForConfigChanges()
    }
```

This code snippet shows how to listen for configuration changes from the Configuration service after connecting  above.

```
func listenForConfigChanges() {
	if ConfigClient == nil {
		LoggingClient.Error("listenForConfigChanges() configuration client not set")
		return
	}

	ConfigClient.WatchForChanges(updateChannel, errChannel, &WritableInfo{}, internal.WritableKey)

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <- signalChan:
			// Quietly and gracefully stop when SIGINT/SIGTERM received
			return

		case ex := <-errChannel:
			LoggingClient.Error(ex.Error())

		case raw, ok := <-updateChannel:
			if !ok {
				return
			}

			actual, ok := raw.(*WritableInfo)
			if !ok {
				LoggingClient.Error("listenForConfigChanges() type check failed")
				return
			}

			Configuration.Writable = *actual

			LoggingClient.Info("Writeable configuration has been updated")
			LoggingClient.SetLogLevel(Configuration.Writable.LogLevel)
		}
	}
}
```