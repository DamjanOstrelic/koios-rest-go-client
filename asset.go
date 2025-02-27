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

package koios

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

type (
	// Asset.
	Asset struct {
		// Asset Name (hex).
		Name string `json:"asset_name"`

		// Asset Policy ID (hex).
		PolicyID PolicyID `json:"policy_id"`

		// Quantity
		// Input: asset balance on the selected input transaction.
		// Output: sum of assets for output UTxO.
		// Mint: sum of minted assets (negative on burn).
		Quantity Lovelace `json:"quantity"`
	}

	// Asset metadata registered on the Cardano Token Registry.
	TokenRegistryMetadata struct {
		Decimals    int    `json:"decimals"`
		Description string `json:"description"`

		// A PNG image file as a byte string
		Logo   string `json:"logo"`
		Name   string `json:"name"`
		Ticker string `json:"ticker"`
		URL    string `json:"url"`
	}

	// AssetSummary aggregated asset summary.
	AssetSummary struct {
		// Asset Name (hex)
		AssetName string `json:"asset_name"`

		// Asset Policy ID (hex)
		PolicyID PolicyID `json:"policy_id"`

		// Total number of registered wallets holding the given asset
		StakedWallets int64 `json:"staked_wallets"`

		// Total number of transactions including the given asset
		TotalTransactions int64 `json:"total_transactions"`

		// Total number of payment addresses (not belonging
		// to registered wallets) holding the given asset
		UnstakedAddresses int64 `json:"unstaked_addresses"`
	}

	// AssetInfo info about the asset.
	AssetInfo struct {
		// Asset Name (hex).
		Name string `json:"asset_name"`

		// Asset Name (ASCII)
		NameASCII string `json:"asset_name_ascii"`

		// The CIP14 fingerprint of the asset
		Fingerprint string `json:"fingerprint"`

		// MintingTxMetadata minting Tx JSON payload if it can be decoded as JSON
		MintingTxMetadata *TxInfoMetadata `json:"minting_tx_metadata"`

		// Asset metadata registered on the Cardano Token Registry
		TokenRegistryMetadata *TokenRegistryMetadata `json:"token_registry_metadata"`

		// Asset Policy ID (hex).
		PolicyID PolicyID `json:"policy_id"`

		// TotalSupply of Asset
		TotalSupply Lovelace `json:"total_supply"`

		// CreationTime of Asset
		CreationTime string `json:"creation_time"`
	}

	// AssetTxs Txs info for the given asset (latest first).
	AssetTxs struct {
		// AssetName (hex)
		AssetName string `json:"asset_name"`

		// PoliciID Asset Policy ID (hex)
		PolicyID PolicyID `json:"policy_id"`

		// TxHashes List of Tx hashes
		TxHashes []TxHash `json:"tx_hashes"`
	}

	// AssetListItem used to represent response from /asset_list`.
	AssetListItem struct {
		PolicyID   PolicyID `json:"policy_id"`
		AssetNames struct {
			HEX   []string `json:"hex"`
			ASCII []string `json:"ascii"`
		} `json:"asset_names"`
	}

	// AssetListResponse represents response from `/asset_list` endpoint.
	AssetListResponse struct {
		Response
		Data []AssetListItem `json:"response"`
	}

	// AssetHolder payment addresses holding the given token (including balance).
	AssetHolder struct {
		PaymentAddress Address  `json:"payment_address"`
		Quantity       Lovelace `json:"quantity"`
	}

	// AssetAddressListResponse represents response from `/asset_address_list` endpoint.
	AssetAddressListResponse struct {
		Response
		Data []AssetHolder `json:"response"`
	}

	// AssetInfoResponse represents response from `/asset_info` endpoint.
	AssetInfoResponse struct {
		Data *AssetInfo `json:"response"`
		Response
	}

	// AssetSummaryResponse represents response from `/asset_summary` endpoint.
	AssetSummaryResponse struct {
		Response
		Data *AssetSummary `json:"response"`
	}

	// AssetTxsResponse represents response from `/asset_txs` endpoint.
	AssetTxsResponse struct {
		Response
		Data *AssetTxs `json:"response"`
	}
)

// GetTip returns the list of all native assets (paginated).
func (c *Client) GetAssetList(ctx context.Context) (res *AssetListResponse, err error) {
	res = &AssetListResponse{}
	rsp, err := c.request(ctx, &res.Response, "GET", "/asset_list", nil, nil, nil)
	if err != nil {
		res.applyError(nil, err)
		return
	}

	body, err := readResponseBody(rsp)
	if err != nil {
		res.applyError(body, err)
		return
	}

	if err = json.Unmarshal(body, &res.Data); err != nil {
		res.applyError(body, err)
		return
	}

	if rsp.StatusCode != http.StatusOK {
		res.applyError(body, err)
		return
	}
	return res, nil
}

// GetAssetAddressList returns the list of all addresses holding a given asset.
func (c *Client) GetAssetAddressList(
	ctx context.Context,
	policy PolicyID,
	name AssetName,
) (res *AssetAddressListResponse, err error) {
	res = &AssetAddressListResponse{}

	params := url.Values{}
	params.Set("_asset_policy", string(policy))
	params.Set("_asset_name", string(name))

	rsp, err := c.request(ctx, &res.Response, "GET", "/asset_address_list", nil, params, nil)
	if err != nil {
		res.applyError(nil, err)
		return
	}

	body, err := readResponseBody(rsp)
	if err != nil {
		res.applyError(body, err)
		return
	}

	if err = json.Unmarshal(body, &res.Data); err != nil {
		res.applyError(body, err)
		return
	}

	if rsp.StatusCode != http.StatusOK {
		res.applyError(body, err)
		return
	}
	return res, nil
}

// GetAssetInfo returns the information of an asset including
// first minting & token registry metadata.
//nolint: dupl
func (c *Client) GetAssetInfo(
	ctx context.Context,
	policy PolicyID,
	name AssetName,
) (res *AssetInfoResponse, err error) {
	res = &AssetInfoResponse{}

	params := url.Values{}
	params.Set("_asset_policy", string(policy))
	params.Set("_asset_name", string(name))

	rsp, err := c.request(ctx, &res.Response, "GET", "/asset_info", nil, params, nil)
	if err != nil {
		res.applyError(nil, err)
		return
	}

	body, err := readResponseBody(rsp)
	if err != nil {
		res.applyError(body, err)
		return
	}

	info := []AssetInfo{}

	if err = json.Unmarshal(body, &info); err != nil {
		res.applyError(body, err)
		return
	}

	if rsp.StatusCode != http.StatusOK {
		res.applyError(body, err)
		return
	}
	if len(info) == 1 {
		res.Data = &info[0]
	}
	res.ready()
	return res, nil
}

// GetAssetSummary returns the summary of an asset
// (total transactions exclude minting/total wallets
// include only wallets with asset balance).
//nolint: dupl
func (c *Client) GetAssetSummary(
	ctx context.Context,
	policy PolicyID,
	name AssetName,
) (res *AssetSummaryResponse, err error) {
	res = &AssetSummaryResponse{}

	params := url.Values{}
	params.Set("_asset_policy", string(policy))
	params.Set("_asset_name", string(name))

	rsp, err := c.request(ctx, &res.Response, "GET", "/asset_summary", nil, params, nil)
	if err != nil {
		res.applyError(nil, err)
		return
	}

	body, err := readResponseBody(rsp)
	if err != nil {
		res.applyError(body, err)
		return
	}

	summary := []AssetSummary{}

	if err = json.Unmarshal(body, &summary); err != nil {
		res.applyError(body, err)
		return
	}

	if rsp.StatusCode != http.StatusOK {
		res.applyError(body, err)
		return
	}
	if len(summary) == 1 {
		res.Data = &summary[0]
	}
	res.ready()
	return res, nil
}

// GetAssetTxs returns the list of all asset transaction hashes (newest first).
//nolint: dupl
func (c *Client) GetAssetTxs(
	ctx context.Context,
	policy PolicyID,
	name AssetName,
) (res *AssetTxsResponse, err error) {
	res = &AssetTxsResponse{}

	params := url.Values{}
	params.Set("_asset_policy", string(policy))
	params.Set("_asset_name", string(name))

	rsp, err := c.request(ctx, &res.Response, "GET", "/asset_txs", nil, params, nil)
	if err != nil {
		res.applyError(nil, err)
		return
	}

	body, err := readResponseBody(rsp)
	if err != nil {
		res.applyError(body, err)
		return
	}

	atxs := []AssetTxs{}

	if err = json.Unmarshal(body, &atxs); err != nil {
		res.applyError(body, err)
		return
	}

	if rsp.StatusCode != http.StatusOK {
		res.applyError(body, err)
		return
	}
	if len(atxs) == 1 {
		res.Data = &atxs[0]
	}
	res.ready()
	return res, nil
}
