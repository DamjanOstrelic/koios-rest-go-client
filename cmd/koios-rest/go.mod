module github.com/howijd/koios-rest-go-client/cmd/koios-rest

go 1.18

replace (
	github.com/howijd/koios-rest-go-client v0.0.0 => ../..
	github.com/shopspring/decimal v1.3.1 => github.com/howijd/decimal v1.3.1
)

require (
	github.com/howijd/koios-rest-go-client v0.0.0
	github.com/urfave/cli/v2 v2.3.1-0.20220204072150-1bf639b391aa
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	golang.org/x/text v0.3.7 // indirect
)
