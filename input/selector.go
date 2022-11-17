package input

import (
	"net"
	"net/url"

	"github.com/roncewind/load/input/rabbitmq"
	"github.com/senzing/go-logging/messagelogger"
	"github.com/spf13/viper"
)

var (
	msglog messagelogger.MessageLoggerInterface
)

// load is 6201:  https://github.com/Senzing/knowledge-base/blob/main/lists/senzing-product-ids.md
const productIdentifier = 6201

var idMessages = map[int]string{
	21: "Scheme: %s",
	22: "User full: %s",
	23: "User name: %s",
	24: "Password: %s",
	25: "Host full: %s",
	26: "Host: %s",
	27: "Port: %s",
	28: "Path: %s",
	29: "Fragment: %s",
	30: "Query string: %s",
	31: "Parsed query string: %s",
}

// ----------------------------------------------------------------------------
func parseURL(urlString string) *url.URL {
	u, err := url.Parse(urlString)
	if err != nil {
		panic(err)
	}

	msglog.Log(21, u.Scheme, messagelogger.LevelInfo)
	msglog.Log(22, u.User, messagelogger.LevelInfo)
	msglog.Log(23, u.User.Username(), messagelogger.LevelInfo)
	p, _ := u.User.Password()
	if len(p) > 0 {
		msglog.Log(24, "SET, redacted from logs", messagelogger.LevelInfo)
	} else {
		msglog.Log(24, "NOT SET", messagelogger.LevelInfo)
	}

	msglog.Log(25, u.Host, messagelogger.LevelInfo)
	host, port, _ := net.SplitHostPort(u.Host)
	msglog.Log(26, host, messagelogger.LevelInfo)
	msglog.Log(27, port, messagelogger.LevelInfo)

	msglog.Log(28, u.Path, messagelogger.LevelInfo)
	msglog.Log(29, u.Fragment, messagelogger.LevelInfo)

	msglog.Log(30, u.RawQuery, messagelogger.LevelInfo)
	m, _ := url.ParseQuery(u.RawQuery)
	msglog.Log(31, m, messagelogger.LevelInfo)

	return u
}

// ----------------------------------------------------------------------------
func Read() bool {
	if viper.IsSet("logLevel") {
		msglog.SetLogLevelFromString(viper.GetString("logLevel"))
	}

	u := parseURL(viper.GetString("inputURL"))
	switch u.Scheme {
	case "amqp":
		if viper.IsSet("inputURL") &&
			viper.IsSet("exchange") &&
			viper.IsSet("inputQueue") {
			rabbitmq.Read(viper.GetString("inputURL"), viper.GetString("exchange"), viper.GetString("inputQueue"))
		} else {
			return false
		}
	default:
		msglog.Log(2001, u.Scheme, messagelogger.LevelWarn)
	}
	return true
}

// ----------------------------------------------------------------------------
func init() {
	msglog, _ = messagelogger.NewSenzingLogger(productIdentifier, idMessages)
}
