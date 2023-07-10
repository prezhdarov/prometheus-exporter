package api

import (
	"flag"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prezhdarov/prometheus-exporter/collector"
)

var (
	// This is example of all things necessary for simple http (REST anyone?) API configuration. Get all these defined and let the feast begin.
	apiUser   = flag.String("api.username", "", "Username to login")
	apiPasswd = flag.String("api.password", "", "Password for the user above")
	apiServer = flag.String("api.server", "", "Server address in host:port format.")
	apiSchema = flag.String("api.schema", "https", "Use HTTP or HTTPS")
	apiSSL    = flag.Bool("api.ssl", false, "Trust SSL or trust")
)

// Nothing to see here..
type APIClient struct {
	//logger  log.Logger
}

// If someone comes with more elegant way to load this into the main program (that sentence feels like 1960s... all we need is a punch card )
func Load(logger log.Logger) {

	level.Info(logger).Log("msg", "Loading Example API")

}

// This here puts it into the collector settings. Remember those handlers and handle functions? Yes, there!
func init() {

	collector.RegisterAPI(NewAPI())

}

// This Always felt this particular approach in golang is like pulling up on its own hair. Germans know what I mean.. But it works.
func NewAPI() *APIClient {

	return &APIClient{}
}

// The Login function. Takes a target name or address as input and returns a map where anything can be stored. From API key to set of cookes - you name it.
// Not sure this - the return map of string and anything -  is as elegant as I want it to be, but quite handy.
func (vm *APIClient) Login(target string) (map[string]interface{}, error) {

	loginData := make(map[string]interface{}, 0)

	if target == "" {

		target = *apiServer

	}

	loginData["target"] = target

	return loginData, nil
}

// The Logout - just pass the map created in Login... Your logout should know what to do with it (if anything at all)
func (vm *APIClient) Logout(loginData map[string]interface{}) error {

	return nil

}

// This one can return virtually anything... and an error. To each (API and exporter) their own as they say.
func (vm *APIClient) Get(loginData, extraConfig map[string]interface{}) (interface{}, error) {

	return nil, nil

}
