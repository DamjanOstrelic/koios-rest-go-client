// Copyright 2022 The Howijd.Network Authors
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//   http://www.apache.org/licenses/LICENSE-2.0
//   or LICENSE file in repository root.
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/howijd/koios-rest-go-client"
	"github.com/tidwall/pretty"
	"github.com/urfave/cli/v2"
)

func main() {
	api, err := koios.New()
	handleErr(err)

	app := &cli.App{
		Version: koios.LibraryVersion,
		Flags:   globalFlags(),
		Authors: []*cli.Author{
			&cli.Author{
				Name: "The Howijd.Network Authors",
			},
		},
		Copyright:            "(c) 2022",
		Usage:                "CLI Client to consume Koios API https://api.koios.rest",
		EnableBashCompletion: true,
		Before: func(c *cli.Context) error {
			handleErr(koios.Host(c.String("host"))(api))
			handleErr(koios.APIVersion(c.String("api-version"))(api))
			handleErr(koios.Port(uint16(c.Uint("port")))(api))
			handleErr(koios.Schema(c.String("schema"))(api))
			handleErr(koios.RateLimit(uint8(c.Uint("rate-limit")))(api))
			handleErr(koios.Origin(c.String("origin"))(api))
			return nil
		},
	}

	addGeneralCommands(app, api)
	addNetworkCommands(app, api)
	addEpochCommands(app, api)
	addBlockCommands(app, api)
	addTransactionsCommands(app, api)
	addAddressCommands(app, api)
	addAccountCommands(app, api)
	addAssetCommands(app, api)
	addPoolCommands(app, api)
	addScriptCommands(app, api)
	addDevCommands(app, api)

	handleErr(app.Run(os.Args))
}

func handleErr(err error) {
	if err == nil {
		return
	}
	trace := err
	for errors.Unwrap(trace) != nil {
		trace = errors.Unwrap(trace)
		log.Println(trace)
	}
	log.Fatal(err)
}

func printResponseBody(ctx *cli.Context, body []byte) {
	if ctx.Bool("ugly") {
		fmt.Println(string(body))
		return
	}
	fmt.Println(string(pretty.Pretty(body)))
}

type printable interface {
	JSON() []byte
}

func output(ctx *cli.Context, data interface{}, err error) {
	out, err := json.Marshal(data)
	handleErr(err)
	printResponseBody(ctx, out)
}

func globalFlags() []cli.Flag {
	return []cli.Flag{
		&cli.UintFlag{
			Name:    "port",
			Aliases: []string{"p"},
			Usage:   "Set port",
			Value:   uint(koios.DefaultPort),
		},
		&cli.StringFlag{
			Name:  "host",
			Usage: "Set host",
			Value: koios.MainnetHost,
		},
		&cli.StringFlag{
			Name:  "api-version",
			Usage: "Set API version",
			Value: koios.DefaultAPIVersion,
		},
		&cli.StringFlag{
			Name:  "schema",
			Usage: "Set URL schema",
			Value: koios.DefaultSchema,
		},
		&cli.StringFlag{
			Name:  "origin",
			Usage: "Set Origin header for requests.",
			Value: koios.DefaultOrigin,
		},
		&cli.UintFlag{
			Name:  "rate-limit",
			Usage: "Set API Client rate limit for outgoing requests",
			Value: uint(koios.DefaultRateLimit),
		},
		&cli.BoolFlag{
			Name:  "ugly",
			Usage: "Ugly prints response json strings directly without calling json pretty.",
			Value: false,
		},
	}
}
