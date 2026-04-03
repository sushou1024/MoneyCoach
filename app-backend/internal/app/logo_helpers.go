package app

import "context"

func (s *Server) resolveHoldingLogos(ctx context.Context, holdings []portfolioHolding) map[string]string {
	logos := make(map[string]string)
	if s == nil || len(holdings) == 0 {
		return logos
	}

	if s.market != nil {
		coinIDs := make([]string, 0, len(holdings))
		for _, holding := range holdings {
			if holding.AssetType != "crypto" || holding.CoinGeckoID == "" {
				continue
			}
			coinIDs = append(coinIDs, holding.CoinGeckoID)
		}
		coinIDs = uniqueSortedStrings(coinIDs)
		if len(coinIDs) > 0 {
			coinLogos, err := s.market.coinGeckoLogos(ctx, coinIDs)
			if err != nil {
				s.logger.Printf("coingecko logos error: %v", err)
			} else {
				for _, holding := range holdings {
					if holding.AssetType != "crypto" || holding.CoinGeckoID == "" || holding.AssetKey == "" {
						continue
					}
					if logo := coinLogos[holding.CoinGeckoID]; logo != "" {
						logos[holding.AssetKey] = logo
					}
				}
			}
		}
	}

	if s.logos != nil {
		for _, holding := range holdings {
			if holding.AssetType != "stock" || holding.AssetKey == "" {
				continue
			}
			displaySymbol := hkDisplaySymbol(holding.Symbol)
			if !isHongKongStockSymbol(displaySymbol, holding.ExchangeMIC) {
				if logo := s.logos.stockLogoURL(holding.ExchangeMIC, displaySymbol); logo != "" {
					logos[holding.AssetKey] = logo
				}
				continue
			}
			if logo := s.logos.hkLogoURL(ctx, displaySymbol); logo != "" {
				logos[holding.AssetKey] = logo
			}
		}
	}

	return logos
}
