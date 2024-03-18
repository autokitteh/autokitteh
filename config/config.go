package config

// FIXME: extract all config stuff, configsets etc from backend to here

const ServiceUrlConfigKey string = "http.service_url"

func ServiceUrlArg(akAddr string) []string {
	return []string{"--config", ServiceUrlConfigKey + "=http://" + akAddr}
}
