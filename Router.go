package pgo

import (
    "regexp"
    "strings"

    "github.com/pinguo/pgo/Util"
)

// format route string to CamelCase, eg.
// /api/foo-bar/say-hello => /Api/FooBar/SayHello
func routeFormatFunc(s string) string {
    s = strings.ToUpper(s)
    if s[0] == '-' {
        s = s[1:]
    }
    return s
}

type routeRule struct {
    rePat   *regexp.Regexp
    pattern string
    route   string
}

// Router the router component, configuration:
// router:
//     rules:
//         - "^/foo/all$ => /foo/index"
//         - "^/api/user/(\d+)$ => /api/user"
type Router struct {
    reFmt *regexp.Regexp
    rules []routeRule
}

func (r *Router) Construct() {
    r.reFmt = regexp.MustCompile(`([/-][a-z])`)
    r.rules = make([]routeRule, 0, 10)
}

// SetRules set rule list, format: `^/api/user/(\d+)$ => /api/user`
func (r *Router) SetRules(rules []interface{}) {
    for _, v := range rules {
        parts := strings.Split(v.(string), "=>")
        if len(parts) != 2 {
            panic("Router: invalid rule: " + Util.ToString(v))
        }

        pattern := strings.TrimSpace(parts[0])
        route := strings.TrimSpace(parts[1])
        r.AddRoute(pattern, route)
    }
}

// AddRoute add one route, the captured group will be passed to
// action method as function params
func (r *Router) AddRoute(pattern, route string) {
    rePat := regexp.MustCompile(pattern)
    rule := routeRule{rePat, pattern, route}
    r.rules = append(r.rules, rule)
}

// Resolve path to route and action params, then format route to CamelCase
func (r *Router) Resolve(path string) (route string, params []string) {
    path = Util.CleanPath(path)

    if len(r.rules) != 0 {
        for _, rule := range r.rules {
            matches := rule.rePat.FindStringSubmatch(path)
            if len(matches) != 0 {
                path = rule.route
                params = matches[1:]
                break
            }
        }
    }

    route = r.reFmt.ReplaceAllStringFunc(path, routeFormatFunc)
    return
}
