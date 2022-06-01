package api

import (
	"flag"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prezhdarov/prometheus-exporter/collector"
)

var (
	apiUser   = flag.String("api.username", "", "Username to login")
	apiPasswd = flag.String("api.password", "", "Password for the user above")
	apiServer = flag.String("vmware.vcenter", "", "Server address in host:port format.")
	apiSchema = flag.String("vmware.schema", "https", "Use HTTP or HTTPS")
	apiSSL    = flag.Bool("vmware.ssl", false, "Trust SSL or trust")
)

type APIClient struct {
	//logger  log.Logger
}

func Load(logger log.Logger) {

	level.Info(logger).Log("msg", "Loading Example API")

}

func init() {

	collector.RegisterAPI(NewAPI())

}

func NewAPI() *APIClient {

	return &APIClient{}
}

func (vm *APIClient) Login(target string) (map[string]interface{}, error) {

	loginData := make(map[string]interface{}, 0)

	if target == "" {

		target = *apiServer

	}

	loginData["target"] = target

	return loginData, nil
}

func (vm *APIClient) Logout(loginData map[string]interface{}) error {

	return nil

}

func (vm *APIClient) Get(loginData, extraConfig map[string]interface{}) (interface{}, error) {

	return nil, nil

}
