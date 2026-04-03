package app

import "errors"

var errPaymentConflict = errors.New("payment already recorded for another user")
var errSubscriptionClaimed = errors.New("subscription already claimed by another user")
