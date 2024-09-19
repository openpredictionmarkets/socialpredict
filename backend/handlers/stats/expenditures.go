package statshandlers

import "gorm.io/gorm"

func calculateTotalExpenditures(db *gorm.DB) int64 {
	// Sum up all expenditure components

	return sumMarketsSuccessfullyResolvedBonus(db)
}

func sumMarketsSuccessfullyResolvedBonus(db *gorm.DB) int64 {
	// sum up all markets successfully resolved as not N/A
	// fee should have been paid out
	// As of authoring there are no bonuses paid to market makers upon resolution
	return 0
}
