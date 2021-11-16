package traefikServiceFabricPlugin

import (
	"log"
	"strconv"
	"strings"

	"github.com/traefik/genconf/dynamic"
)

// GetStringValue get string value associated to a label.
func GetStringValue(labels map[string]string, labelName, defaultValue string) string {
	if value, ok := labels[labelName]; ok && len(value) > 0 {
		return value
	}
	return defaultValue
}

// GetBoolValue get bool value associated to a label.
func GetBoolValue(labels map[string]string, labelName string, defaultValue bool) bool {
	rawValue, ok := labels[labelName]
	if ok {
		v, err := strconv.ParseBool(rawValue)
		if err == nil {
			return v
		}
		log.Printf("Unable to parse %q: %q, falling back to %v. %v", labelName, rawValue, defaultValue, err)
	}
	return defaultValue
}

// GetIntValue get int value associated to a label.
func GetIntValue(labels map[string]string, labelName string, defaultValue int) int {
	if rawValue, ok := labels[labelName]; ok {
		value, err := strconv.Atoi(rawValue)
		if err == nil {
			return value
		}
		log.Printf("Unable to parse %q: %q, falling back to %v. %v", labelName, rawValue, defaultValue, err)
	}
	return defaultValue
}

func setLoadbalancerPasshostheader(lb *dynamic.ServersLoadBalancer, val string) error {
	v, err := strconv.ParseBool(val)
	if err != nil {
		v = false
	}

	lb.PassHostHeader = &v
	return nil
}

func setLoadbalancerSticky(lb *dynamic.ServersLoadBalancer, val string) error {
	v, err := strconv.ParseBool(val)
	if err != nil {
		v = false
	}

	if v {
		if lb.Sticky == nil {
			lb.Sticky = &dynamic.Sticky{ Cookie: &dynamic.Cookie{} }
		}
	}
	return nil
}

func setLoadbalancerStickySecure(lb *dynamic.ServersLoadBalancer, val string) error {
	v, err := strconv.ParseBool(val)
	if err != nil {
		v = false
	}

	if v {
		if lb.Sticky == nil {
			lb.Sticky = &dynamic.Sticky{ Cookie: &dynamic.Cookie{} }
		}
		lb.Sticky.Cookie.Secure = v
	}
	return nil
}

func setLoadbalancerStickyHttpOnly(lb *dynamic.ServersLoadBalancer, val string) error {
	v, err := strconv.ParseBool(val)
	if err != nil {
		v = false
	}

	if v {
		if lb.Sticky == nil {
			lb.Sticky = &dynamic.Sticky{ Cookie: &dynamic.Cookie{} }
		}
		lb.Sticky.Cookie.HTTPOnly = v
	}
	return nil
}

func setLoadbalancerStickySameSite(lb *dynamic.ServersLoadBalancer, val string) error {
	if lb.Sticky == nil {
		lb.Sticky = &dynamic.Sticky{ Cookie: &dynamic.Cookie{} }
	}
	
	// Value must be "none", "lax", or "strict".
	valid := map[string]bool{"none": true, "lax": true, "strict": true}
    if valid[strings.ToLower(val)] {
		lb.Sticky.Cookie.SameSite = val
	} else {
		log.Printf("Unrecognised value '%s' provided for Cookie.SameSite", val)
		return nil
	}
	return nil
}

func setLoadbalancerStickyCookieName(lb *dynamic.ServersLoadBalancer, val string) error {
	if lb.Sticky == nil {
		lb.Sticky = &dynamic.Sticky{ Cookie: &dynamic.Cookie{} }
	}
	lb.Sticky.Cookie.Name = val
	return nil
}


func setLoadbalancerHealthcheckPath(lb *dynamic.ServersLoadBalancer, val string) error {
	if lb.HealthCheck == nil {
		lb.HealthCheck = &dynamic.HealthCheck{}
	}

	lb.HealthCheck.Path = val
	return nil
}

func setLoadbalancerHealthcheckInterval(lb *dynamic.ServersLoadBalancer, val string) error {
	if lb.HealthCheck == nil {
		lb.HealthCheck = &dynamic.HealthCheck{}
	}

	lb.HealthCheck.Interval = val
	return nil
}

func setLoadbalancerHealthcheckScheme(lb *dynamic.ServersLoadBalancer, val string) error {
	if lb.HealthCheck == nil {
		lb.HealthCheck = &dynamic.HealthCheck{}
	}

	lb.HealthCheck.Scheme = val
	return nil
}

func setMiddlewareStriptprefixPrefixes(name string, middlewares map[string]*dynamic.Middleware, router *dynamic.Router, val string) error {
	m, ok := middlewares[name]
	if !ok {
		m = &dynamic.Middleware{
			StripPrefix: &dynamic.StripPrefix{},
		}
	}

	m.StripPrefix.Prefixes = []string{val}
	middlewares[name] = m

	router.Middlewares = append(router.Middlewares, name)

	return nil
}
