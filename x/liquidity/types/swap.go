package types

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Type of match
type MatchType int

const (
	ExactMatch MatchType = iota + 1
	NoMatch
	FractionalMatch
)

// Direction of price
type PriceDirection int

const (
	Increasing PriceDirection = iota + 1
	Decreasing
	Staying
)

// Direction of order
type OrderDirection int

const (
	DirectionXtoY OrderDirection = iota + 1
	DirectionYtoX
)

// Type of order map to index at price, having the pointer list of the swap batch message.
type Order struct {
	Price         sdk.Dec
	BuyOfferAmt   sdk.Int
	SellOfferAmt  sdk.Int
	SwapMsgStates []*SwapMsgState
}

// OrderBook is a list of orders
type OrderBook []Order

// Len implements sort.Interface for OrderBook
func (orderBook OrderBook) Len() int { return len(orderBook) }

// Less implements sort.Interface for OrderBook
func (orderBook OrderBook) Less(i, j int) bool {
	return orderBook[i].Price.LT(orderBook[j].Price)
}

// Swap implements sort.Interface for OrderBook
func (orderBook OrderBook) Swap(i, j int) { orderBook[i], orderBook[j] = orderBook[j], orderBook[i] }

// increasing sort orderbook by order price
func (orderBook OrderBook) Sort() {
	sort.Slice(orderBook, func(i, j int) bool {
		return orderBook[i].Price.LT(orderBook[j].Price)
	})
}

// decreasing sort orderbook by order price
func (orderBook OrderBook) Reverse() {
	sort.Slice(orderBook, func(i, j int) bool {
		return orderBook[i].Price.GT(orderBook[j].Price)
	})
}

// Get number of not matched messages on the list.
func CountNotMatchedMsgs(swapMsgStates []*SwapMsgState) int {
	cnt := 0
	for _, m := range swapMsgStates {
		if m.Executed && !m.Succeeded {
			cnt++
		}
	}
	return cnt
}

// Get number of fractional matched messages on the list.
func CountFractionalMatchedMsgs(swapMsgStates []*SwapMsgState) int {
	cnt := 0
	for _, m := range swapMsgStates {
		if m.Executed && m.Succeeded && !m.ToBeDeleted {
			cnt++
		}
	}
	return cnt
}

// Order map type indexed by order price at price
type OrderMap map[string]Order

// Make orderbook by sort orderMap.
func (orderMap OrderMap) SortOrderBook() (orderBook OrderBook) {
	for _, o := range orderMap {
		orderBook = append(orderBook, o)
	}
	orderBook.Sort()
	return orderBook
}

// struct of swap matching result of the batch
type BatchResult struct {
	MatchType      MatchType
	PriceDirection PriceDirection
	SwapPrice      sdk.Dec
	EX             sdk.Dec
	EY             sdk.Dec
	OriginalEX     sdk.Int
	OriginalEY     sdk.Int
	PoolX          sdk.Dec
	PoolY          sdk.Dec
	TransactAmt    sdk.Dec
}

// return of zero object, to avoid nil
func NewBatchResult() BatchResult {
	return BatchResult{
		SwapPrice:   sdk.ZeroDec(),
		EX:          sdk.ZeroDec(),
		EY:          sdk.ZeroDec(),
		OriginalEX:  sdk.ZeroInt(),
		OriginalEY:  sdk.ZeroInt(),
		PoolX:       sdk.ZeroDec(),
		PoolY:       sdk.ZeroDec(),
		TransactAmt: sdk.ZeroDec(),
	}
}

// struct of swap matching result of each Batch swap message
type MatchResult struct {
	OrderDirection         OrderDirection
	OrderMsgIndex          uint64
	OrderPrice             sdk.Dec
	OfferCoinAmt           sdk.Dec
	TransactedCoinAmt      sdk.Dec
	ExchangedDemandCoinAmt sdk.Dec
	OfferCoinFeeAmt        sdk.Dec
	ExchangedCoinFeeAmt    sdk.Dec
	SwapMsgState           *SwapMsgState
}

// The price and coins of swap messages in orderbook are calculated
// to derive match result with the price direction.
func (orderBook OrderBook) Match(x, y sdk.Dec) (BatchResult, bool) {
	currentPrice := x.Quo(y)
	priceDirection := orderBook.PriceDirection(currentPrice)
	if priceDirection == Staying {
		return orderBook.CalculateMatchStay(currentPrice), true
	}
	return orderBook.CalculateMatch(priceDirection, x, y)
}

// Check orderbook validity naively
func (orderBook OrderBook) Validate(currentPrice sdk.Dec) bool {
	if !currentPrice.IsPositive() {
		return false
	}
	maxBuyOrderPrice := sdk.ZeroDec()
	minSellOrderPrice := sdk.NewDec(1000000000000)
	for _, order := range orderBook {
		if order.BuyOfferAmt.IsPositive() && order.Price.GT(maxBuyOrderPrice) {
			maxBuyOrderPrice = order.Price
		}
		if order.SellOfferAmt.IsPositive() && (order.Price.LT(minSellOrderPrice)) {
			minSellOrderPrice = order.Price
		}
	}
	if maxBuyOrderPrice.GT(minSellOrderPrice) ||
		maxBuyOrderPrice.Quo(currentPrice).GT(sdk.MustNewDecFromStr("1.10")) ||
		minSellOrderPrice.Quo(currentPrice).LT(sdk.MustNewDecFromStr("0.90")) {
		return false
	}
	return true
}

// Calculate results for orderbook matching with unchanged price case
func (orderBook OrderBook) CalculateMatchStay(currentPrice sdk.Dec) (r BatchResult) {
	r = NewBatchResult()
	r.SwapPrice = currentPrice
	r.OriginalEX, r.OriginalEY = orderBook.ExecutableAmt(r.SwapPrice)
	r.EX = r.OriginalEX.ToDec()
	r.EY = r.OriginalEY.ToDec()
	r.PriceDirection = Staying

	s := r.SwapPrice.Mul(r.EY)
	if r.EX.IsZero() || r.EY.IsZero() {
		r.MatchType = NoMatch
	} else if r.EX.Equal(s) { // Normalization to an integrator for easy determination of exactMatch
		r.MatchType = ExactMatch
	} else {
		// Decimal Error, When calculating the Executable value, conservatively Truncated decimal
		r.MatchType = FractionalMatch
		if r.EX.GT(s) {
			r.EX = s
		} else if r.EX.LT(s) {
			r.EY = r.EX.Quo(r.SwapPrice)
		}
	}
	return
}

// Calculates the batch results with the logic for each direction
func (orderBook OrderBook) CalculateMatch(direction PriceDirection, x, y sdk.Dec) (maxScenario BatchResult, found bool) {
	currentPrice := x.Quo(y)
	lastOrderPrice := currentPrice
	var matchScenarios []BatchResult
	start, end, delta := 0, len(orderBook)-1, 1
	if direction == Decreasing {
		start, end, delta = end, start, -1
	}
	for i := start; i != end+delta; i += delta {
		order := orderBook[i]
		if (direction == Increasing && order.Price.LT(currentPrice)) ||
			(direction == Decreasing && order.Price.GT(currentPrice)) {
			continue
		} else {
			orderPrice := order.Price
			r := orderBook.CalculateSwap(direction, x, y, orderPrice, lastOrderPrice)
			// Check to see if it exceeds a value that can be a decimal error
			if (direction == Increasing && r.PoolY.Sub(r.EX.Quo(r.SwapPrice)).GTE(sdk.OneDec())) ||
				(direction == Decreasing && r.PoolX.Sub(r.EY.Mul(r.SwapPrice)).GTE(sdk.OneDec())) {
				continue
			}
			matchScenarios = append(matchScenarios, r)
			lastOrderPrice = orderPrice
		}
	}
	maxScenario = NewBatchResult()
	for _, s := range matchScenarios {
		MEX, MEY := orderBook.MustExecutableAmt(s.SwapPrice)
		if s.EX.GTE(MEX.ToDec()) && s.EY.GTE(MEY.ToDec()) {
			if s.MatchType == ExactMatch && s.TransactAmt.IsPositive() {
				maxScenario = s
				found = true
				break
			} else if s.TransactAmt.GT(maxScenario.TransactAmt) {
				maxScenario = s
				found = true
			}
		}
	}
	maxScenario.PriceDirection = direction
	return maxScenario, found
}

// CalculateSwap calculates the batch result.
func (orderBook OrderBook) CalculateSwap(direction PriceDirection, x, y, orderPrice, lastOrderPrice sdk.Dec) BatchResult {
	r := NewBatchResult()
	r.OriginalEX, r.OriginalEY = orderBook.ExecutableAmt(lastOrderPrice.Add(orderPrice).Quo(sdk.NewDec(2)))
	r.EX = r.OriginalEX.ToDec()
	r.EY = r.OriginalEY.ToDec()

	r.SwapPrice = x.Add(r.EX.MulInt64(2)).Quo(y.Add(r.EY.MulInt64(2))) // P_s = (X + 2EX) / (Y + 2EY)

	if direction == Increasing {
		r.PoolY = r.SwapPrice.Mul(y).Sub(x).Quo(r.SwapPrice.MulInt64(2)) // (P_s * Y - X / 2P_s)
		if lastOrderPrice.LT(r.SwapPrice) && r.SwapPrice.LT(orderPrice) && !r.PoolY.IsNegative() {
			if r.EX.IsZero() && r.EY.IsZero() {
				r.MatchType = NoMatch
			} else {
				r.MatchType = ExactMatch
			}
		}
	} else if direction == Decreasing {
		r.PoolX = x.Sub(r.SwapPrice.Mul(y)).QuoInt64(2) // (X - P_s * Y) / 2
		if orderPrice.LT(r.SwapPrice) && r.SwapPrice.LT(lastOrderPrice) && !r.PoolX.IsNegative() {
			if r.EX.IsZero() && r.EY.IsZero() {
				r.MatchType = NoMatch
			} else {
				r.MatchType = ExactMatch
			}
		}
	}

	if r.MatchType == 0 {
		r.OriginalEX, r.OriginalEY = orderBook.ExecutableAmt(orderPrice)
		r.EX = r.OriginalEX.ToDec()
		r.EY = r.OriginalEY.ToDec()
		r.SwapPrice = orderPrice
		// When calculating the Pool value, conservatively Truncated decimal, so Ceil it to reduce the decimal error
		if direction == Increasing {
			r.PoolY = r.SwapPrice.Mul(y).Sub(x).Quo(r.SwapPrice.MulInt64(2)) // (P_s * Y - X) / 2P_s
			r.EX = sdk.MinDec(r.EX, r.EY.Add(r.PoolY).Mul(r.SwapPrice)).Ceil()
			r.EY = sdk.MaxDec(sdk.MinDec(r.EY, r.EX.Quo(r.SwapPrice).Sub(r.PoolY)), sdk.ZeroDec()).Ceil()
		} else if direction == Decreasing {
			r.PoolX = x.Sub(r.SwapPrice.Mul(y)).QuoInt64(2) // (X - P_s * Y) / 2
			r.EY = sdk.MinDec(r.EY, r.EX.Add(r.PoolX).Quo(r.SwapPrice)).Ceil()
			r.EX = sdk.MaxDec(sdk.MinDec(r.EX, r.EY.Mul(r.SwapPrice).Sub(r.PoolX)), sdk.ZeroDec()).Ceil()
		}
		r.MatchType = FractionalMatch
	}

	if direction == Increasing {
		if r.SwapPrice.LT(x.Quo(y)) || r.PoolY.IsNegative() {
			r.TransactAmt = sdk.ZeroDec()
		} else {
			r.TransactAmt = sdk.MinDec(r.EX, r.EY.Add(r.PoolY).Mul(r.SwapPrice))
		}
	} else if direction == Decreasing {
		if r.SwapPrice.GT(x.Quo(y)) || r.PoolX.IsNegative() {
			r.TransactAmt = sdk.ZeroDec()
		} else {
			r.TransactAmt = sdk.MinDec(r.EY, r.EX.Add(r.PoolX).Quo(r.SwapPrice))
		}
	}
	return r
}

// Get Price direction of the orderbook with current Price
func (orderBook OrderBook) PriceDirection(currentPrice sdk.Dec) PriceDirection {
	buyAmtOverCurrentPrice := sdk.ZeroDec()
	buyAmtAtCurrentPrice := sdk.ZeroDec()
	sellAmtUnderCurrentPrice := sdk.ZeroDec()
	sellAmtAtCurrentPrice := sdk.ZeroDec()

	for _, order := range orderBook {
		if order.Price.GT(currentPrice) {
			buyAmtOverCurrentPrice = buyAmtOverCurrentPrice.Add(order.BuyOfferAmt.ToDec())
		} else if order.Price.Equal(currentPrice) {
			buyAmtAtCurrentPrice = buyAmtAtCurrentPrice.Add(order.BuyOfferAmt.ToDec())
			sellAmtAtCurrentPrice = sellAmtAtCurrentPrice.Add(order.SellOfferAmt.ToDec())
		} else if order.Price.LT(currentPrice) {
			sellAmtUnderCurrentPrice = sellAmtUnderCurrentPrice.Add(order.SellOfferAmt.ToDec())
		}
	}
	if buyAmtOverCurrentPrice.GT(currentPrice.Mul(sellAmtUnderCurrentPrice.Add(sellAmtAtCurrentPrice))) {
		return Increasing
	} else if currentPrice.Mul(sellAmtUnderCurrentPrice).GT(buyAmtOverCurrentPrice.Add(buyAmtAtCurrentPrice)) {
		return Decreasing
	}
	return Staying
}

// calculate the executable amount of the orderbook for each X, Y
func (orderBook OrderBook) ExecutableAmt(swapPrice sdk.Dec) (executableBuyAmtX, executableSellAmtY sdk.Int) {
	executableBuyAmtX = sdk.ZeroInt()
	executableSellAmtY = sdk.ZeroInt()
	for _, order := range orderBook {
		if order.Price.GTE(swapPrice) {
			executableBuyAmtX = executableBuyAmtX.Add(order.BuyOfferAmt)
		}
		if order.Price.LTE(swapPrice) {
			executableSellAmtY = executableSellAmtY.Add(order.SellOfferAmt)
		}
	}
	return
}

// Check swap executable amount validity of the orderbook
func (orderBook OrderBook) MustExecutableAmt(swapPrice sdk.Dec) (mustExecutableBuyAmtX, mustExecutableSellAmtY sdk.Int) {
	mustExecutableBuyAmtX = sdk.ZeroInt()
	mustExecutableSellAmtY = sdk.ZeroInt()
	for _, order := range orderBook {
		if order.Price.GT(swapPrice) {
			mustExecutableBuyAmtX = mustExecutableBuyAmtX.Add(order.BuyOfferAmt)
		}
		if order.Price.LT(swapPrice) {
			mustExecutableSellAmtY = mustExecutableSellAmtY.Add(order.SellOfferAmt)
		}
	}
	return
}

// make orderMap key as swap price, value as Buy, Sell Amount from swap msgs, with split as Buy xToY, Sell yToX msg list.
func MakeOrderMap(swapMsgs []*SwapMsgState, denomX, denomY string, onlyNotMatched bool) (OrderMap, []*SwapMsgState, []*SwapMsgState) {
	orderMap := make(OrderMap)
	var xToY []*SwapMsgState // buying Y from X
	var yToX []*SwapMsgState // selling Y for X
	for _, m := range swapMsgs {
		if onlyNotMatched && (m.ToBeDeleted || m.RemainingOfferCoin.IsZero()) {
			continue
		}
		order := Order{
			Price:        m.Msg.OrderPrice,
			BuyOfferAmt:  sdk.ZeroInt(),
			SellOfferAmt: sdk.ZeroInt(),
		}
		orderPriceString := m.Msg.OrderPrice.String()
		switch {
		// buying Y from X
		case m.Msg.OfferCoin.Denom == denomX:
			xToY = append(xToY, m)
			if o, ok := orderMap[orderPriceString]; ok {
				order = o
				order.BuyOfferAmt = o.BuyOfferAmt.Add(m.RemainingOfferCoin.Amount)
			} else {
				order.BuyOfferAmt = m.RemainingOfferCoin.Amount
			}
		// selling Y for X
		case m.Msg.OfferCoin.Denom == denomY:
			yToX = append(yToX, m)
			if o, ok := orderMap[orderPriceString]; ok {
				order = o
				order.SellOfferAmt = o.SellOfferAmt.Add(m.RemainingOfferCoin.Amount)
			} else {
				order.SellOfferAmt = m.RemainingOfferCoin.Amount
			}
		default:
			panic(ErrInvalidDenom)
		}
		order.SwapMsgStates = append(order.SwapMsgStates, m)
		orderMap[orderPriceString] = order
	}
	return orderMap, xToY, yToX
}

// check validity state of the batch swap messages, and set to delete state to height timeout expired order
func ValidateStateAndExpireOrders(swapMsgStates []*SwapMsgState, currentHeight int64, expireThisHeight bool) {
	for _, order := range swapMsgStates {
		if !order.Executed {
			panic("not executed")
		}
		if order.RemainingOfferCoin.IsZero() {
			if !order.Succeeded || !order.ToBeDeleted {
				panic("broken state consistency for not matched order")
			}
			continue
		}
		// set toDelete, expired msgs
		if currentHeight > order.OrderExpiryHeight {
			if order.Succeeded || !order.ToBeDeleted {
				panic("broken state consistency for fractional matched order")
			}
			continue
		}
		if expireThisHeight && currentHeight == order.OrderExpiryHeight {
			order.ToBeDeleted = true
		}
	}
}

// Check swap price validity using list of match result.
func CheckSwapPrice(matchResultXtoY, matchResultYtoX []MatchResult, swapPrice sdk.Dec) bool {
	if len(matchResultXtoY) == 0 && len(matchResultYtoX) == 0 {
		return true
	}
	// Check if it is greater than a value that can be a decimal error
	for _, m := range matchResultXtoY {
		if m.TransactedCoinAmt.Quo(swapPrice).Sub(m.ExchangedDemandCoinAmt).Abs().GT(sdk.OneDec()) {
			return false
		}
	}
	for _, m := range matchResultYtoX {
		if m.TransactedCoinAmt.Mul(swapPrice).Sub(m.ExchangedDemandCoinAmt).Abs().GT(sdk.OneDec()) {
			return false
		}
	}
	return !swapPrice.IsZero()
}

// Find matched orders and set status for msgs
func FindOrderMatch(direction OrderDirection, swapMsgStates []*SwapMsgState, executableAmt, swapPrice sdk.Dec, height int64) (
	matchResults []MatchResult, poolXDelta, poolYDelta sdk.Dec) {
	poolXDelta = sdk.ZeroDec()
	poolYDelta = sdk.ZeroDec()

	if executableAmt.IsZero() {
		return
	}

	if direction == DirectionXtoY {
		sort.SliceStable(swapMsgStates, func(i, j int) bool {
			return swapMsgStates[i].Msg.OrderPrice.GT(swapMsgStates[j].Msg.OrderPrice)
		})
	} else if direction == DirectionYtoX {
		sort.SliceStable(swapMsgStates, func(i, j int) bool {
			return swapMsgStates[i].Msg.OrderPrice.LT(swapMsgStates[j].Msg.OrderPrice)
		})
	}

	matchAmt := sdk.ZeroInt()
	accumMatchAmt := sdk.ZeroInt()
	var matchedSwapMsgStates []*SwapMsgState //nolint:prealloc

	for i, order := range swapMsgStates {
		// include the matched order in matchAmt, matchedSwapMsgStates
		if (direction == DirectionXtoY && order.Msg.OrderPrice.LT(swapPrice)) ||
			(direction == DirectionYtoX && order.Msg.OrderPrice.GT(swapPrice)) {
			break
		}

		matchAmt = matchAmt.Add(order.RemainingOfferCoin.Amount)
		matchedSwapMsgStates = append(matchedSwapMsgStates, order)

		if i == len(swapMsgStates)-1 || !swapMsgStates[i+1].Msg.OrderPrice.Equal(order.Msg.OrderPrice) {
			if matchAmt.IsPositive() {
				var fractionalMatchRatio sdk.Dec
				if accumMatchAmt.Add(matchAmt).ToDec().GTE(executableAmt) {
					fractionalMatchRatio = executableAmt.Sub(accumMatchAmt.ToDec()).Quo(matchAmt.ToDec())
					if fractionalMatchRatio.GT(sdk.NewDec(1)) {
						panic("fractionalMatchRatio should be between 0 and 1")
					}
				} else {
					fractionalMatchRatio = sdk.OneDec()
				}
				if !fractionalMatchRatio.IsPositive() {
					fractionalMatchRatio = sdk.OneDec()
				}
				for _, matchOrder := range matchedSwapMsgStates {
					offerAmt := matchOrder.RemainingOfferCoin.Amount.ToDec()
					matchResult := MatchResult{
						OrderDirection: direction,
						OfferCoinAmt:   offerAmt,
						// TransactedCoinAmt is a value that should not be lost, so Ceil it conservatively considering the decimal error.
						TransactedCoinAmt: offerAmt.Mul(fractionalMatchRatio).Ceil(),
						SwapMsgState:      matchOrder,
					}
					if matchResult.OfferCoinAmt.Sub(matchResult.TransactedCoinAmt).LTE(sdk.OneDec()) {
						// Use ReservedOfferCoinFee to avoid decimal errors when OfferCoinAmt and TransactedCoinAmt are almost equal in value.
						matchResult.OfferCoinFeeAmt = matchResult.SwapMsgState.ReservedOfferCoinFee.Amount.ToDec()
					} else {
						matchResult.OfferCoinFeeAmt = matchResult.SwapMsgState.ReservedOfferCoinFee.Amount.ToDec().Mul(fractionalMatchRatio)
					}
					if direction == DirectionXtoY {
						matchResult.ExchangedDemandCoinAmt = matchResult.TransactedCoinAmt.Quo(swapPrice)
						matchResult.ExchangedCoinFeeAmt = matchResult.OfferCoinFeeAmt.Quo(swapPrice)
					} else if direction == DirectionYtoX {
						matchResult.ExchangedDemandCoinAmt = matchResult.TransactedCoinAmt.Mul(swapPrice)
						matchResult.ExchangedCoinFeeAmt = matchResult.OfferCoinFeeAmt.Mul(swapPrice)
					}
					// Check for differences above maximum decimal error
					if matchResult.TransactedCoinAmt.GT(matchResult.OfferCoinAmt) {
						panic("bad TransactedCoinAmt")
					}
					if matchResult.OfferCoinFeeAmt.GT(matchResult.OfferCoinAmt) && matchResult.OfferCoinFeeAmt.GT(sdk.OneDec()) {
						panic("bad OfferCoinFeeAmt")
					}
					matchResults = append(matchResults, matchResult)
					if direction == DirectionXtoY {
						poolXDelta = poolXDelta.Add(matchResult.TransactedCoinAmt)
						poolYDelta = poolYDelta.Sub(matchResult.ExchangedDemandCoinAmt)
					} else if direction == DirectionYtoX {
						poolXDelta = poolXDelta.Sub(matchResult.ExchangedDemandCoinAmt)
						poolYDelta = poolYDelta.Add(matchResult.TransactedCoinAmt)
					}
				}
				accumMatchAmt = accumMatchAmt.Add(matchAmt)
			}

			matchAmt = sdk.ZeroInt()
			matchedSwapMsgStates = matchedSwapMsgStates[:0]
		}
	}
	return matchResults, poolXDelta, poolYDelta
}

// UpdateSwapMsgStates updates SwapMsgStates using the MatchResults.
func UpdateSwapMsgStates(x, y sdk.Dec, xToY, yToX []*SwapMsgState, matchResultXtoY, matchResultYtoX []MatchResult) (
	[]*SwapMsgState, []*SwapMsgState, sdk.Dec, sdk.Dec, sdk.Dec, sdk.Dec) {
	sort.SliceStable(xToY, func(i, j int) bool {
		return xToY[i].Msg.OrderPrice.GT(xToY[j].Msg.OrderPrice)
	})
	sort.SliceStable(yToX, func(i, j int) bool {
		return yToX[i].Msg.OrderPrice.LT(yToX[j].Msg.OrderPrice)
	})

	poolXDelta := sdk.ZeroDec()
	poolYDelta := sdk.ZeroDec()

	// Variables to accumulate and offset the values of int 1 caused by decimal error
	decimalErrorX := sdk.ZeroDec()
	decimalErrorY := sdk.ZeroDec()

	for _, match := range append(matchResultXtoY, matchResultYtoX...) {
		sms := match.SwapMsgState
		if match.OrderDirection == DirectionXtoY {
			poolXDelta = poolXDelta.Add(match.TransactedCoinAmt)
			poolYDelta = poolYDelta.Sub(match.ExchangedDemandCoinAmt)
		} else {
			poolXDelta = poolXDelta.Sub(match.ExchangedDemandCoinAmt)
			poolYDelta = poolYDelta.Add(match.TransactedCoinAmt)
		}
		if sms.RemainingOfferCoin.Amount.ToDec().Sub(match.TransactedCoinAmt).LTE(sdk.OneDec()) {
			// when RemainingOfferCoin and TransactedCoinAmt are almost equal in value, corrects the decimal error and processes as a exact match.
			sms.ExchangedOfferCoin.Amount = sms.ExchangedOfferCoin.Amount.Add(match.TransactedCoinAmt.TruncateInt())
			sms.RemainingOfferCoin.Amount = sms.RemainingOfferCoin.Amount.Sub(match.TransactedCoinAmt.TruncateInt())
			sms.ReservedOfferCoinFee.Amount = sms.ReservedOfferCoinFee.Amount.Sub(match.OfferCoinFeeAmt.TruncateInt())
			if sms.ExchangedOfferCoin.IsNegative() || sms.RemainingOfferCoin.IsNegative() || sms.ReservedOfferCoinFee.IsNegative() {
				panic("negative coin amount after update")
			}
			if sms.RemainingOfferCoin.Amount.Equal(sdk.OneInt()) {
				decimalErrorY = decimalErrorY.Add(sdk.OneDec())
				sms.RemainingOfferCoin.Amount = sdk.ZeroInt()
			}
			if !sms.RemainingOfferCoin.IsZero() || sms.ExchangedOfferCoin.Amount.GT(sms.Msg.OfferCoin.Amount) ||
				sms.ReservedOfferCoinFee.Amount.GT(sdk.OneInt()) {
				panic("invalid state after update")
			} else {
				sms.Succeeded = true
				sms.ToBeDeleted = true
			}
		} else {
			// fractional match
			sms.ExchangedOfferCoin.Amount = sms.ExchangedOfferCoin.Amount.Add(match.TransactedCoinAmt.TruncateInt())
			sms.RemainingOfferCoin.Amount = sms.RemainingOfferCoin.Amount.Sub(match.TransactedCoinAmt.TruncateInt())
			sms.ReservedOfferCoinFee.Amount = sms.ReservedOfferCoinFee.Amount.Sub(match.OfferCoinFeeAmt.TruncateInt())
			if sms.ExchangedOfferCoin.IsNegative() || sms.RemainingOfferCoin.IsNegative() || sms.ReservedOfferCoinFee.IsNegative() {
				panic("negative coin amount after update")
			}
			sms.Succeeded = true
			sms.ToBeDeleted = false
		}
	}

	// Offset accumulated decimal error values
	poolXDelta = poolXDelta.Add(decimalErrorX)
	poolYDelta = poolYDelta.Add(decimalErrorY)

	x = x.Add(poolXDelta)
	y = y.Add(poolYDelta)

	return xToY, yToX, x, y, poolXDelta, poolYDelta
}
