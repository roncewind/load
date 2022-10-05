package input

import (
	// "encoding/json"
	// "errors"
	"fmt"
	"net"
	// "net/http"
	"net/url"
	// "os"

	// amqp "github.com/rabbitmq/amqp091-go"
	"github.com/roncewind/load/input/rabbitmq"
	// "github.com/roncewind/szrecord"
	// "github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ----------------------------------------------------------------------------
func parseURL(urlString string) (*url.URL){
	u, err := url.Parse(urlString)
	if err != nil {
		panic(err)
	}


	fmt.Println("===============================")
	fmt.Println("\tScheme: ", u.Scheme)
	fmt.Println("\tUser full: ", u.User)
	fmt.Println("\tUser name: ", u.User.Username())
	p, _ := u.User.Password()
	fmt.Println("\tPassword: ", p)

	fmt.Println("\tHost full: ", u.Host)
	host, port, _ := net.SplitHostPort(u.Host)
	fmt.Println("\tHost: ", host)
	fmt.Println("\tPort: ", port)

	fmt.Println("\tPath: ", u.Path)
	fmt.Println("\tFragment: ", u.Fragment)

	fmt.Println("\tQuery string: ", u.RawQuery)
	m, _ := url.ParseQuery(u.RawQuery)
	fmt.Println("\tParsed query string: ", m)
	// fmt.Println(m["k"][0])
	fmt.Println("===============================")

	return u
}

// ----------------------------------------------------------------------------
func Read() (bool) {
	u := parseURL(viper.GetString("inputURL"))
	switch u.Scheme {
	case "amqp":
		if( viper.IsSet("inputURL") &&
			viper.IsSet("exchange") &&
			viper.IsSet("inputQueue")) {
			rabbitmq.Read(viper.GetString("inputURL"), viper.GetString("exchange"), viper.GetString("inputQueue"))
		} else {
			return false
		}
	default:
		fmt.Println("Unknown input mechanism: %s", u.Scheme)
	}
	return true
}

