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

func setLoadbalancerPasshostheader(lb *dynamic.ServersLoadBalancer, val string) {
	v, err := strconv.ParseBool(val)
	if err != nil {
		v = false
	}

	lb.PassHostHeader = &v
}

func newSticky() *dynamic.Sticky {
	return &dynamic.Sticky{Cookie: &dynamic.Cookie{}}
}

func setLoadbalancerSticky(lb *dynamic.ServersLoadBalancer, val string) {
	v, err := strconv.ParseBool(val)
	if err != nil {
		v = false
	}

	if v {
		if lb.Sticky == nil {
			lb.Sticky = newSticky()
		}
	}
}

func setLoadbalancerStickySecure(lb *dynamic.ServersLoadBalancer, val string) {
	v, err := strconv.ParseBool(val)
	if err != nil {
		v = false
	}

	if v {
		if lb.Sticky == nil {
			lb.Sticky = newSticky()
		}
		lb.Sticky.Cookie.Secure = v
	}
}

func setLoadbalancerStickyHTTPOnly(lb *dynamic.ServersLoadBalancer, val string) {
	v, err := strconv.ParseBool(val)
	if err != nil {
		v = false
	}

	if v {
		if lb.Sticky == nil {
			lb.Sticky = newSticky()
		}
		lb.Sticky.Cookie.HTTPOnly = v
	}
}

func setLoadbalancerStickySameSite(lb *dynamic.ServersLoadBalancer, val string) {
	if lb.Sticky == nil {
		lb.Sticky = newSticky()
	}

	// Value must be "none", "lax", or "strict".
	valid := map[string]bool{"none": true, "lax": true, "strict": true}
	if valid[strings.ToLower(val)] {
		lb.Sticky.Cookie.SameSite = val
	} else {
		log.Printf("Unrecognized value '%s' provided for Cookie.SameSite", val)
		return
	}
}

func setLoadbalancerStickyCookieName(lb *dynamic.ServersLoadBalancer, val string) {
	if lb.Sticky == nil {
		lb.Sticky = newSticky()
	}
	lb.Sticky.Cookie.Name = val
}

func setLoadbalancerHealthcheckPath(lb *dynamic.ServersLoadBalancer, val string) {
	if lb.HealthCheck == nil {
		lb.HealthCheck = &dynamic.ServerHealthCheck{}
	}

	lb.HealthCheck.Path = val
}

func setLoadbalancerHealthcheckInterval(lb *dynamic.ServersLoadBalancer, val string) {
	if lb.HealthCheck == nil {
		lb.HealthCheck = &dynamic.ServerHealthCheck{}
	}

	lb.HealthCheck.Interval = val
}

func setLoadbalancerHealthcheckScheme(lb *dynamic.ServersLoadBalancer, val string) {
	if lb.HealthCheck == nil {
		lb.HealthCheck = &dynamic.ServerHealthCheck{}
	}

	lb.HealthCheck.Scheme = val
}

func setMiddlewareStriptprefixPrefixes(name string, middlewares map[string]*dynamic.Middleware, router *dynamic.Router, val string) {
	m, ok := middlewares[name]
	if !ok {
		m = &dynamic.Middleware{
			StripPrefix: &dynamic.StripPrefix{},
		}
	}

	m.StripPrefix.Prefixes = []string{val}
	middlewares[name] = m

	router.Middlewares = append(router.Middlewares, name)
}
