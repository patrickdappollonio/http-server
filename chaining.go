package main

import "net/http"

type middleware func(http.Handler) http.Handler

type middlechain struct {
	all []middleware
}

func chain(md ...middleware) middlechain {
	return middlechain{
		all: append([]middleware{}, md...),
	}
}

func (chain middlechain) then(handler http.Handler) http.Handler {
	if handler == nil {
		handler = http.DefaultServeMux
	}

	for pos := range chain.all {
		handler = chain.all[len(chain.all)-1-pos](handler)
	}

	return handler
}

func (chain middlechain) extend(md ...middleware) middlechain {
	fork := append([]middleware{}, append(chain.all, md...)...)
	return middlechain{fork}
}
