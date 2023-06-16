package healthcheck

const (
	NginxApp = "nginx"
)

type healthcheck func() error

func ValidateSupportedApps(app string) bool {
	return true
}

func ApplyHealthchecks(app string) (err error) {
	healthchecks := getHealthCheckFunctions(app)
	for _, f := range healthchecks {
		err = f()
		if err != nil {
			return err
		}
	}
	return nil
}

func getHealthCheckFunctions(app string) []healthcheck {
	switch app {
	case NginxApp:
		return []healthcheck{
			indexHealthy,
		}
	default:
		return nil
	}
}
