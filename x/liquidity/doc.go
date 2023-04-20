/*
Package liquidity implements a Cosmos SDK module,that provides an implementation
of the serves AMM(Automated Market Makers) style decentralized liquidity providing and
coin swap functions. The module enable anyone to create a liquidity pool, deposit or withdraw coins
from the liquidity pool, and request coin swap to the liquidity pool

Please refer to the specification under /spec and Resources below for further information.

Resources
  - Liquidity Module Spec https://github.com/cosmos/gaia/v9/blob/develop/x/liquidity/spec
  - Liquidity Module Lite Paper https://github.com/cosmos/gaia/v9/blob/develop/doc/LiquidityModuleLightPaper_EN.pdf
  - Proposal and milestone https://github.com/b-harvest/Liquidity-Module-For-the-Hub
  - Swagger HTTP API doc https://app.swaggerhub.com/apis-docs/bharvest/cosmos-sdk_liquidity_module_rest_and_g_rpc_gateway_docs
  - Client doc https://github.com/cosmos/gaia/v9/blob/develop/doc/client.md
*/
package liquidity
