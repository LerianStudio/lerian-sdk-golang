package leriantest

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/fees"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/shopspring/decimal"
)

func findMatchingFakePackage(all []fees.Package, input *fees.FeeCalculate) (*fees.Package, bool, error) {
	amount, ok := decimalFromAny(input.Transaction.Send.Value)
	if !ok {
		return nil, false, nil
	}

	bestScore := -1
	var bestMatches []fees.Package

	for _, pack := range all {
		if pack.LedgerID != input.LedgerID {
			continue
		}

		if pack.Enable != nil && !*pack.Enable {
			continue
		}

		score, matches := packageSpecificity(pack, input, amount)
		if !matches {
			continue
		}

		cloned := clonePackage(pack)
		if score > bestScore {
			bestScore = score
			bestMatches = []fees.Package{cloned}
			continue
		}

		if score == bestScore {
			bestMatches = append(bestMatches, cloned)
		}
	}

	if len(bestMatches) == 0 {
		return nil, false, nil
	}

	if len(bestMatches) == 1 {
		return &bestMatches[0], true, nil
	}

	return nil, false, sdkerrors.NewValidation("Fees.Calculate", "Fee", "more than one package matched the transaction")
}

func packageSpecificity(pack fees.Package, input *fees.FeeCalculate, amount decimal.Decimal) (int, bool) {
	if !packageAmountMatches(pack, amount) {
		return 0, false
	}

	score := 0
	if pack.TransactionRoute != nil {
		if *pack.TransactionRoute != input.Transaction.Route {
			return 0, false
		}

		score++
	}

	if pack.SegmentID != nil {
		if input.SegmentID == nil || *pack.SegmentID != *input.SegmentID {
			return 0, false
		}

		score++
	}

	return score, true
}

func packageAmountMatches(pack fees.Package, amount decimal.Decimal) bool {
	minAmount, ok := decimalFromAny(pack.MinimumAmount)
	if !ok {
		return false
	}

	maxAmount, ok := decimalFromAny(pack.MaximumAmount)
	if !ok {
		return false
	}

	return amount.GreaterThanOrEqual(minAmount) && amount.LessThanOrEqual(maxAmount)
}

func packageMatchesEstimate(pack fees.Package, input *fees.FeeEstimateInput) bool {
	if input == nil {
		return false
	}

	if pack.LedgerID != input.LedgerID {
		return false
	}

	if pack.Enable != nil && !*pack.Enable {
		return false
	}

	if pack.TransactionRoute != nil && *pack.TransactionRoute != input.Transaction.Route {
		return false
	}

	amount, ok := decimalFromAny(input.Transaction.Send.Value)
	if !ok {
		return false
	}

	return packageAmountMatches(pack, amount)
}

func applyFakeFeesToTransaction(pack fees.Package, input fees.TransactionDSL) (fees.TransactionDSL, bool) {
	amount, ok := decimalFromAny(input.Send.Value)
	if !ok || !packageAmountMatches(pack, amount) || len(pack.Fees) == 0 {
		return cloneTransactionDSL(input), false
	}

	fromAmounts, ok := buildFakeLegAmounts(input.Send.Source.From, input.Send.Asset, amount)
	if !ok {
		return cloneTransactionDSL(input), false
	}

	toAmounts, ok := buildFakeLegAmounts(input.Send.Distribute.To, input.Send.Asset, amount)
	if !ok {
		return cloneTransactionDSL(input), false
	}

	result := cloneTransactionDSL(input)
	originalAmount := amount
	referenceAmount := amount
	resultAmount := amount
	waived := cloneWaivedAccounts(pack.WaivedAccounts)
	applied := false

	feeEntries := make([]struct {
		key string
		fee fees.Fee
	}, 0, len(pack.Fees))
	for key, fee := range pack.Fees {
		feeEntries = append(feeEntries, struct {
			key string
			fee fees.Fee
		}{key: key, fee: fee})
	}

	sort.SliceStable(feeEntries, func(i, j int) bool {
		if feeEntries[i].fee.Priority == feeEntries[j].fee.Priority {
			return feeEntries[i].key < feeEntries[j].key
		}

		return feeEntries[i].fee.Priority < feeEntries[j].fee.Priority
	})

	for feeIndex, entry := range feeEntries {
		feeAmount, ok := calculateFakeFeeAmount(entry.fee, originalAmount, referenceAmount)
		if !ok || feeAmount.LessThanOrEqual(decimal.Zero) {
			continue
		}

		if entry.fee.IsDeductibleFrom != nil && *entry.fee.IsDeductibleFrom {
			updatedTo, changed := applyFakeProportionalFee(entry.fee, entryFeeLabel(entry.fee, entry.key), feeIndex, toAmounts, feeAmount, &waived, true)
			toAmounts = updatedTo
			if changed {
				referenceAmount = referenceAmount.Sub(feeAmount)
				applied = true
			}
			continue
		}

		updatedFrom, updatedTo, changed := applyFakeProportionalFeeWithCounterpart(entry.fee, entryFeeLabel(entry.fee, entry.key), feeIndex, fromAmounts, toAmounts, feeAmount, &waived)
		fromAmounts = updatedFrom
		toAmounts = updatedTo
		if changed {
			referenceAmount = referenceAmount.Add(feeAmount)
			resultAmount = resultAmount.Add(feeAmount)
			applied = true
		}
	}

	if !applied {
		return result, false
	}

	result.Send.Source.From = fakeAmountsToLegs(fromAmounts)
	result.Send.Distribute.To = fakeAmountsToLegs(toAmounts)
	result.Send.Value = formatMoney(resultAmount)
	if result.Metadata == nil {
		result.Metadata = make(map[string]any)
	}
	result.Metadata["packageAppliedID"] = pack.ID

	return result, true
}

type fakeAmount struct {
	Asset string
	Value decimal.Decimal
	Leg   fees.TransactionDSLLeg
}

func buildFakeLegAmounts(legs []fees.TransactionDSLLeg, asset string, total decimal.Decimal) (map[string]fakeAmount, bool) {
	amounts := make(map[string]fakeAmount, len(legs))
	for i, leg := range legs {
		key := fakeLegKey(leg)
		if key == "" {
			continue
		}

		value, ok := fakeLegValue(leg, asset, total)
		if !ok {
			return nil, false
		}

		key = uniqueFakeLegKey(amounts, key, i)
		amounts[key] = value
	}

	return amounts, len(amounts) > 0
}

func fakeLegKey(leg fees.TransactionDSLLeg) string {
	if leg.AccountAlias == "" {
		return ""
	}

	if leg.Route != "" {
		return leg.AccountAlias + "->" + leg.Route
	}

	return leg.AccountAlias
}

func fakeLegValue(leg fees.TransactionDSLLeg, asset string, total decimal.Decimal) (fakeAmount, bool) {
	if leg.Amount != nil {
		value, ok := decimalFromAny(leg.Amount.Value)
		if !ok {
			return fakeAmount{}, false
		}

		legAsset := leg.Amount.Asset
		if legAsset == "" {
			legAsset = asset
		}

		legCopy := cloneTransactionDSLLeg(leg)
		return fakeAmount{Asset: legAsset, Value: value, Leg: legCopy}, true
	}

	if leg.Share != nil && leg.Share.Percentage > 0 {
		share := decimal.NewFromInt(leg.Share.Percentage)
		value := total.Mul(share).Div(decimal.NewFromInt(100))
		legCopy := cloneTransactionDSLLeg(leg)
		return fakeAmount{Asset: asset, Value: value, Leg: legCopy}, true
	}

	return fakeAmount{}, false
}

func applyFakeProportionalFeeWithCounterpart(fee fees.Fee, feeLabel string, feeIndex int, fromAmounts, toAmounts map[string]fakeAmount, feeAmount decimal.Decimal, waived *[]string) (updatedFrom, updatedTo map[string]fakeAmount, changed bool) {
	maxAccount := findMaxFakeAccount(fromAmounts, waived)
	updatedFrom = cloneFakeAmounts(fromAmounts)
	updatedTo = cloneFakeAmounts(toAmounts)
	totalPaying := totalNonWaived(updatedFrom, waived)
	if totalPaying.IsZero() {
		return fromAmounts, toAmounts, false
	}

	allocated := decimal.Zero
	keys := sortedFakeAmountKeys(updatedFrom)
	for _, key := range keys {
		amount := updatedFrom[key]
		if isWaivedAccount(key, waived) {
			continue
		}

		feeApplied := proportionalFeeShare(amount.Value, totalPaying, feeAmount, key == maxAccount)
		allocated = allocated.Add(feeApplied)

		fromKey := key + "->fee" + strconv.Itoa(feeIndex)
		if route := derefString(fee.RouteFrom); route != "" {
			fromKey += "->" + route
		}
		updatedFrom[fromKey] = fakeAmount{Asset: amount.Asset, Value: feeApplied, Leg: newSyntheticFeeLeg(cleanFakeLegAccountAlias(key), derefString(fee.RouteFrom), feeLabel)}

		toKey := fee.CreditAccount + "->fee_source" + strconv.Itoa(feeIndex) + "->" + key
		if route := derefString(fee.RouteTo); route != "" {
			toKey += "->" + route
		}
		updatedTo[toKey] = fakeAmount{Asset: amount.Asset, Value: feeApplied, Leg: newSyntheticFeeLeg(fee.CreditAccount, derefString(fee.RouteTo), feeLabel)}

		*waived = append(*waived, fromKey, toKey)
	}

	applyFeeDelta(updatedFrom, updatedTo, feeIndex, feeAmount.Sub(allocated), maxAccount)
	*waived = append(*waived, fee.CreditAccount)

	return updatedFrom, updatedTo, true
}

func applyFakeProportionalFee(fee fees.Fee, feeLabel string, feeIndex int, amounts map[string]fakeAmount, feeAmount decimal.Decimal, waived *[]string, deductible bool) (map[string]fakeAmount, bool) {
	maxAccount := findMaxFakeAccount(amounts, waived)
	updated := cloneFakeAmounts(amounts)
	totalPaying := totalNonWaived(updated, waived)
	if totalPaying.IsZero() {
		return amounts, false
	}

	allocated := decimal.Zero
	keys := sortedFakeAmountKeys(updated)
	for _, key := range keys {
		amount := updated[key]
		if isWaivedAccount(key, waived) {
			continue
		}

		feeApplied := proportionalFeeShare(amount.Value, totalPaying, feeAmount, key == maxAccount)
		allocated = allocated.Add(feeApplied)

		entryKey := fee.CreditAccount + "->fee_source" + strconv.Itoa(feeIndex) + "->" + key
		if route := derefString(fee.RouteTo); route != "" {
			entryKey += "->" + route
		}
		updated[entryKey] = fakeAmount{Asset: amount.Asset, Value: feeApplied, Leg: newSyntheticFeeLeg(fee.CreditAccount, derefString(fee.RouteTo), feeLabel)}
		amount.Value = amount.Value.Sub(feeApplied)
		updated[key] = amount
		*waived = append(*waived, entryKey)
	}

	if deductible {
		applyDeductibleDelta(updated, feeIndex, feeAmount.Sub(allocated), maxAccount)
	}
	*waived = append(*waived, fee.CreditAccount)

	return updated, true
}

func proportionalFeeShare(amount, total, feeAmount decimal.Decimal, isMax bool) decimal.Decimal {
	share := amount.Div(total)
	calculated := feeAmount.Mul(share)
	if hasRepeatingDecimal(share) {
		if isMax {
			return calculated.RoundCeil(2)
		}

		return calculated.RoundFloor(2)
	}

	return calculated.Round(2)
}

func hasRepeatingDecimal(value decimal.Decimal) bool {
	stringValue := value.String()
	parts := strings.SplitN(stringValue, ".", 2)
	if len(parts) < 2 {
		return false
	}

	decimals := strings.TrimRight(parts[1], "0")
	if len(decimals) < 4 {
		return false
	}

	for size := 1; size <= len(decimals)/2 && size <= 6; size++ {
		pattern := decimals[:size]
		repeats := len(decimals) / size
		if repeats < 2 {
			continue
		}

		matches := true
		for i := 0; i < repeats; i++ {
			start := i * size
			if start+size > len(decimals) || decimals[start:start+size] != pattern {
				matches = false
				break
			}
		}

		if matches {
			return true
		}
	}

	return false
}

func totalNonWaived(amounts map[string]fakeAmount, waived *[]string) decimal.Decimal {
	total := decimal.Zero
	for key, amount := range amounts {
		if isWaivedAccount(key, waived) {
			continue
		}
		total = total.Add(amount.Value)
	}

	return total
}

func findMaxFakeAccount(amounts map[string]fakeAmount, waived *[]string) string {
	maxValue := decimal.Zero
	maxKey := ""
	for _, key := range sortedFakeAmountKeys(amounts) {
		amount := amounts[key]
		if isWaivedAccount(key, waived) {
			continue
		}
		if amount.Value.GreaterThanOrEqual(maxValue) {
			maxValue = amount.Value
			maxKey = key
		}
	}

	return maxKey
}

func applyFeeDelta(fromAmounts, toAmounts map[string]fakeAmount, feeIndex int, delta decimal.Decimal, maxAccount string) {
	if delta.IsZero() || maxAccount == "" {
		return
	}

	for key, amount := range fromAmounts {
		if strings.HasPrefix(key, maxAccount+"->fee"+strconv.Itoa(feeIndex)) {
			amount.Value = amount.Value.Add(delta)
			fromAmounts[key] = amount
			break
		}
	}

	for key, amount := range toAmounts {
		if strings.Contains(key, "fee_source"+strconv.Itoa(feeIndex)+"->"+maxAccount) {
			amount.Value = amount.Value.Add(delta)
			toAmounts[key] = amount
			break
		}
	}
}

func applyDeductibleDelta(amounts map[string]fakeAmount, feeIndex int, delta decimal.Decimal, maxAccount string) {
	if delta.IsZero() || maxAccount == "" {
		return
	}

	for key, amount := range amounts {
		if strings.Contains(key, "fee_source"+strconv.Itoa(feeIndex)+"->"+maxAccount) {
			amount.Value = amount.Value.Add(delta)
			amounts[key] = amount
			break
		}
	}
}

func fakeAmountsToLegs(amounts map[string]fakeAmount) []fees.TransactionDSLLeg {
	legs := make([]fees.TransactionDSLLeg, 0, len(amounts))
	for _, key := range sortedFakeAmountKeys(amounts) {
		amount := amounts[key]
		leg := cloneTransactionDSLLeg(amount.Leg)
		if leg.AccountAlias == "" {
			cleanAccount, metadata, route := parseFakeLegKey(key)
			leg.AccountAlias = cleanAccount
			leg.Route = route
			if len(metadata) > 0 {
				leg.Metadata = metadata
			}
		}
		if leg.Amount == nil {
			leg.Amount = &fees.TransactionDSLAmount{}
		}
		leg.Amount.Asset = amount.Asset
		leg.Amount.Value = formatMoney(amount.Value)
		legs = append(legs, leg)
	}

	return legs
}

func parseFakeLegKey(key string) (string, map[string]any, string) {
	normalizedKey := normalizeFakeAmountKey(key)
	parts := strings.Split(normalizedKey, "->")
	cleanAccount := parts[0]
	metadata := map[string]any{}
	route := ""

	if len(parts) >= 3 && strings.Contains(parts[1], "fee_source") {
		metadata["feeLabel"] = normalizeFakeAmountKey(parts[2])
		if len(parts) > 3 {
			route = parts[len(parts)-1]
		}
		return cleanAccount, metadata, route
	}

	if len(parts) >= 3 && strings.HasPrefix(parts[1], "fee") {
		if len(parts) > 2 {
			route = parts[len(parts)-1]
		}
		return cleanAccount, metadata, route
	}

	if len(parts) > 1 {
		route = parts[len(parts)-1]
	}

	return cleanAccount, metadata, route
}

func uniqueFakeLegKey(amounts map[string]fakeAmount, base string, index int) string {
	if _, exists := amounts[base]; !exists {
		return base
	}

	return base + "#dup" + strconv.Itoa(index)
}

func normalizeFakeAmountKey(key string) string {
	if idx := strings.LastIndex(key, "#dup"); idx >= 0 {
		return key[:idx]
	}

	return key
}

func cloneFakeAmounts(input map[string]fakeAmount) map[string]fakeAmount {
	output := make(map[string]fakeAmount, len(input))
	for key, value := range input {
		output[key] = value
	}

	return output
}

func newSyntheticFeeLeg(accountAlias, route, feeLabel string) fees.TransactionDSLLeg {
	leg := fees.TransactionDSLLeg{AccountAlias: accountAlias, Route: route}
	if feeLabel != "" {
		leg.Metadata = map[string]any{"feeLabel": feeLabel}
	}

	return leg
}

func cleanFakeLegAccountAlias(key string) string {
	accountAlias, _, _ := parseFakeLegKey(key)
	return accountAlias
}

func entryFeeLabel(fee fees.Fee, fallback string) string {
	if strings.TrimSpace(fee.FeeLabel) != "" {
		return fee.FeeLabel
	}

	return fallback
}

func sortedFakeAmountKeys(amounts map[string]fakeAmount) []string {
	keys := make([]string, 0, len(amounts))
	for key := range amounts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func isWaivedAccount(account string, waived *[]string) bool {
	if waived == nil {
		return false
	}

	normalizedAccount := normalizeFakeAmountKey(account)
	for _, exempt := range *waived {
		if normalizeFakeAmountKey(exempt) == normalizedAccount {
			return true
		}
	}

	return false
}

func cloneWaivedAccounts(input *[]string) []string {
	if input == nil {
		return []string{}
	}

	output := make([]string, len(*input))
	copy(output, *input)
	return output
}

func calculateFakeFeeAmount(fee fees.Fee, originalAmount, currentAmount decimal.Decimal) (decimal.Decimal, bool) {
	if fee.CalculationModel == nil || len(fee.CalculationModel.Calculations) == 0 {
		return decimal.Zero, false
	}

	reference := originalAmount
	if fee.ReferenceAmount == "afterFeesAmount" {
		reference = currentAmount
	}

	results := make([]decimal.Decimal, 0, len(fee.CalculationModel.Calculations))
	for _, calc := range fee.CalculationModel.Calculations {
		value, err := decimal.NewFromString(calc.Value)
		if err != nil {
			return decimal.Zero, false
		}

		switch calc.Type {
		case "flat":
			results = append(results, value)
		case "percentage":
			results = append(results, reference.Mul(value).Div(decimal.NewFromInt(100)))
		}
	}

	if len(results) == 0 {
		return decimal.Zero, false
	}

	switch fee.CalculationModel.ApplicationRule {
	case "maxBetweenTypes":
		maximumResult := results[0]
		for _, value := range results[1:] {
			if value.GreaterThan(maximumResult) {
				maximumResult = value
			}
		}
		return maximumResult, true
	default:
		return results[0], true
	}
}

func decimalFromAny(input any) (decimal.Decimal, bool) {
	if input == nil {
		return decimal.Zero, false
	}

	value := reflect.ValueOf(input)
	if value.Kind() == reflect.Pointer && value.IsNil() {
		return decimal.Zero, false
	}

	switch typed := input.(type) {
	case decimal.Decimal:
		return typed, true
	case string:
		parsed, err := decimal.NewFromString(typed)
		return parsed, err == nil
	case int:
		return decimal.NewFromInt(int64(typed)), true
	case int8:
		return decimal.NewFromInt(int64(typed)), true
	case int16:
		return decimal.NewFromInt(int64(typed)), true
	case int32:
		return decimal.NewFromInt(int64(typed)), true
	case int64:
		return decimal.NewFromInt(typed), true
	case uint:
		return decimalFromUnsigned(uint64(typed))
	case uint8:
		return decimalFromUnsigned(uint64(typed))
	case uint16:
		return decimalFromUnsigned(uint64(typed))
	case uint32:
		return decimalFromUnsigned(uint64(typed))
	case uint64:
		return decimalFromUnsigned(typed)
	case float32:
		return decimal.NewFromFloat32(typed), true
	case float64:
		return decimal.NewFromFloat(typed), true
	case fmt.Stringer:
		if stringerValue := reflect.ValueOf(typed); stringerValue.Kind() == reflect.Pointer && stringerValue.IsNil() {
			return decimal.Zero, false
		}
		parsed, err := decimal.NewFromString(typed.String())
		return parsed, err == nil
	default:
		parsed, err := decimal.NewFromString(fmt.Sprint(typed))
		return parsed, err == nil
	}
}

func decimalFromUnsigned(value uint64) (decimal.Decimal, bool) {
	parsed, err := decimal.NewFromString(strconv.FormatUint(value, 10))
	return parsed, err == nil
}

func formatMoney(value decimal.Decimal) string {
	return value.StringFixed(2)
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

func startOfDay(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, value.Location())
}

func endOfDay(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 23, 59, 59, int(time.Second-time.Nanosecond), value.Location())
}
