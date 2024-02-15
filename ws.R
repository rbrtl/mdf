# install.packages(c("httpuv", "jsonlite"))
library(httpuv)
library(jsonlite)

generate_spread <- function(ticker, price, volume, pots) {
	pop <- rnorm(pots, price, sd = price^(1/3))
	prices <- sample(pop, pots)
	volume <- abs(sample(rnorm(100, volume/pots, sd = volume/(pots^2)), pots))
	df <- data.frame(symbol = ticker, price = prices, volume)
}

serve_market_data <- function(ws) {
	ws$onMessage(function(i, msg) {
		fields <- fromJSON(msg)

		ticker <- fields$ticker
		price <- as.integer(fields$price)
		volume <- as.numeric(fields$volume)
		pots <- as.integer(fields$pots)

		ws$send(toJSON(generate_spread(ticker, price, volume, pots)))
	})
}

s <- startServer("0.0.0.0", 8080, list(
		onWSOpen = serve_market_data, 
		onWSClose = function() {NULL}
	)
)

if (exists("s")) {
	s$stop()
}
httpuv::service()
cat("WebSocket server is running on port 8080\n")
