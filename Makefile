.PHONY: build

build:
	cat assets/angular.min.js assets/angular-route.min.js assets/app.js > build/app.js
	cat assets/index.html > build/index.html
	rice embed
