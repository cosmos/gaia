<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation

<a name="top"></a>

## Table of Contents

- [gaia/globalfee/v1beta1/query.proto](#gaia/globalfee/v1beta1/query.proto)
  - [QueryMinimumGasPricesRequest](#gaia.globalfee.v1beta1.QueryMinimumGasPricesRequest)
  - [QueryMinimumGasPricesResponse](#gaia.globalfee.v1beta1.QueryMinimumGasPricesResponse)
  
  - [Query](#gaia.globalfee.v1beta1.Query)
  
- [gaia/globalfee/v1beta1/genesis.proto](#gaia/globalfee/v1beta1/genesis.proto)
  - [GenesisState](#gaia.globalfee.v1beta1.GenesisState)
  - [Params](#gaia.globalfee.v1beta1.Params)
  
- [Scalar Value Types](#scalar-value-types)

<a name="gaia/globalfee/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## gaia/globalfee/v1beta1/query.proto

<a name="gaia.globalfee.v1beta1.QueryMinimumGasPricesRequest"></a>

### QueryMinimumGasPricesRequest

QueryMinimumGasPricesRequest is the request type for the
Query/MinimumGasPrices RPC method.

<a name="gaia.globalfee.v1beta1.QueryMinimumGasPricesResponse"></a>

### QueryMinimumGasPricesResponse

QueryMinimumGasPricesResponse is the response type for the
Query/MinimumGasPrices RPC method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `minimum_gas_prices` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) | repeated |  |

 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

<a name="gaia.globalfee.v1beta1.Query"></a>

### Query

Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `MinimumGasPrices` | [QueryMinimumGasPricesRequest](#gaia.globalfee.v1beta1.QueryMinimumGasPricesRequest) | [QueryMinimumGasPricesResponse](#gaia.globalfee.v1beta1.QueryMinimumGasPricesResponse) |  | GET|/gaia/globalfee/v1beta1/minimum_gas_prices|

 <!-- end services -->

<a name="gaia/globalfee/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## gaia/globalfee/v1beta1/genesis.proto

<a name="gaia.globalfee.v1beta1.GenesisState"></a>

### GenesisState

GenesisState - initial state of module

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#gaia.globalfee.v1beta1.Params) |  | Params of this module |

<a name="gaia.globalfee.v1beta1.Params"></a>

### Params

Params defines the set of module parameters.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `minimum_gas_prices` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) | repeated | Minimum stores the minimum gas price(s) for all TX on the chain. When multiple coins are defined then they are accepted alternatively. The list must be sorted by denoms asc. No duplicate denoms or zero amount values allowed. For more information see <https://docs.cosmos.network/main/modules/auth#concepts> |

 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |
