import { valueToChartY } from '../components/StrategyCandlestickChart'

describe('StrategyCandlestickChart valueToChartY', () => {
  it('spans the full chart height when range < 1', () => {
    const chartTop = 8
    const chartHeight = 140
    const minValue = 0.4
    const maxValue = 0.5

    expect(valueToChartY({ value: maxValue, minValue, maxValue, chartTop, chartHeight })).toBeCloseTo(chartTop)
    expect(valueToChartY({ value: (maxValue + minValue) / 2, minValue, maxValue, chartTop, chartHeight })).toBeCloseTo(
      chartTop + chartHeight / 2
    )
    expect(valueToChartY({ value: minValue, minValue, maxValue, chartTop, chartHeight })).toBeCloseTo(
      chartTop + chartHeight
    )
  })
})
