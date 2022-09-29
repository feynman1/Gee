package gee

import (
	"net/http"
	"strings"
)

type router struct {
	handlers map[string]HandlerFunc
	roots    map[string]*node
}

func newRouter() *router {
	return &router{
		handlers: make(map[string]HandlerFunc),
		roots:    make(map[string]*node),
	}
}

func (r *router) parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")
	parts := make([]string, 0)
	for _, s := range vs {
		if s != "" {
			parts = append(parts, s)
			if s[0] == '*' {
				break
			}
		}
	}
	return parts
}

func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	parts := r.parsePattern(pattern)
	key := method + "-" + pattern
	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = &node{}
	}
	r.roots[method].insert(pattern, &parts, 0)
	r.handlers[key] = handler
}

func (r *router) getRoute(path string, method string) (*node, map[string]string) {
	searchParts := r.parsePattern(path)
	params := make(map[string]string)
	root, ok := r.roots[method]
	if !ok {
		return nil, nil
	}
	n := root.search(&searchParts, 0)
	if n != nil {
		parts := r.parsePattern(n.pattern)
		for index, part := range parts {
			if part[0] == '*' {
				params[part[1:]] = searchParts[index]
			}
			if part[0] == ':' {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, params
	}
	return nil, nil
}

func (r *router) handler(context *Context) {
	root, params := r.getRoute(context.Path, context.Method)
	if root != nil {
		context.Params = params
		key := context.Method + "-" + root.pattern
		r.handlers[key](context)
	} else {
		context.String(http.StatusNotFound, "404 NOT FOUND: %s\n", context.Path)
	}
}
