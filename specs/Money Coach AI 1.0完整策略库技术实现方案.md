**Money Coach AI 1.0完整策略库技术实现方案**

版本: 1.0 日期: 2026-01-04  

 

**目录**

1  	引言1.1. 设计原则  
1.2. 策略库全景图  
1.3. 策略匹配逻辑  
1.4. 技术架构
1.5. MVP Gap (实现侧偏离)

2  	MVP阶段策略 (S01-S05)
2.1. S01: 固定止损策略  
2.2. S02: 马丁格尔解套策略  
2.3. S03: 移动止盈策略  
2.4. S04: 分层止盈策略  
2.5. S05: 定投建议策略

3  	V1.5阶段策略 (S06-S13)
3.1. S06: 现货网格策略  
3.2. S07: 定时定额卖出策略  
3.3. S08: 价值平均策略  
3.4. S09: 反马丁格尔加仓策略  
3.5. S10: 金字塔加仓策略  
3.6. S11: 区间交易策略  
3.7. S12: 均值回归策略  
3.8. S13: 保本止损策略

4  	V2.0阶段策略 (S14-S20)
4.1. S14: 跨交易所套利策略  
4.2. S15: 三角套利策略  
4.3. S16: 资金费率套利策略  
4.4. S17: 期现套利策略  
4.5. S18: 趋势跟随策略  
4.6. S19: 突破交易策略  
4.7. S20: 合约对冲策略

5  	V3.0阶段策略 (S21-S25)
5.1. S21: 动态再平衡策略  
5.2. S22: 风险平价策略  
5.3. S23: 税收优化策略  
5.4. S24: 波动率套利策略  
5.5. S25: 事件驱动交易策略

6  	附录
6.1. 开发路线图与工作量估算  
6.2. 关键成功指标 (KPI)
 
**1. 引言**

本文档详细阐述了Money Coach AI产品中25个核心交易策略的技术实现方案。旨在为开发团队提供一份清晰、完整、可执行的开发指南。

 

**1.1. 设计原则**

•   	全场景覆盖: 策略库覆盖从新手到专家，从牛市到熊市，从盈利到亏损的各种场景。

•   	分层实施: 策略按MVP、V1.5、V2.0、V3.0四个阶段分层实施，逐步扩展功能。

•   	AI驱动匹配: 基于用户画像、市场环境、持仓状况智能推荐最适合的策略。

•   	风险透明: 每个策略都有清晰的风险等级、适用条件和潜在风险提示。

•   	可执行性: 从“建议”到“一键执行”，逐步实现自动化。

 

**1.2. 策略库全景图**

策略库共包含25个策略，分为风险管理、解套回本、自动化交易、套利、技术分析和组合管理六大类。详细矩阵请参见 money_coach_strategy_matrix.md。

 

**1.3. 策略匹配逻辑**

系统通过AI诊断引擎分析用户持仓，识别核心问题，然后根据问题类型、用户画像和市场环境，从策略库中匹配最合适的单个或组合策略。详细决策树请参见 money_coach_strategy_matrix.md。

 

**1.4. 技术架构**

[用户层]  
	↓  
[策略匹配引擎]  
	├─ 规则引擎  
	├─ AI推理引擎 (LLM)  
	└─ 优先级排序  
	↓  
[策略库 (25个策略)]  
	├─ 参数计算模块  
	├─ 回测验证模块  
	├─ 风险评估模块  
	└─ AI文案生成模块  
	↓  
[执行层]  
	├─ 建议模式 (MVP)  
	├─ 半自动模式 (V1.5)  
	└─ 全自动模式 (V2.0+)

**1.5. MVP Gap (实现侧偏离)**

以下偏离项以 PRD + prototypes 为实现侧权威，确保 MVP 可交付：

- 策略详细算法文件（/home/ubuntu/strategies_detailed_part*.md 等）不在 repo；MVP 参数与规则以 PRD + prototypes 为准，缺失内容不要推断。
- S02/S03：参数与触发规则已补齐，MVP 允许生成计划与参数（以 PRD/prototypes 为准）。
- 策略阶段：本文件标注为 V1.5/V2.0/V3.0 的 S09/S16/S18/S22，在 MVP 中已启用（实现以 PRD/prototypes 为准）。
- S01/S04：动态调整规则（止损/分层止盈的后续规则）在 MVP 暂不实现，仅输出初始参数。
- S01：受 MVP 计划数量上限影响，每份报告最多 3 个 plan，且 S01 仅在非现金持仓中选择 1 个输出（不改变 S01 的普适性，仅限制输出数量）。
- S03：分层移动止盈方案在 MVP 暂不实现，仅输出单一移动止盈参数。
- S04：holding_time 条件在 MVP 跳过，仅使用 avg_price + pnl_percent + 价格触发条件。
- S05：定投金额/频率改用 idle_cash_usd + non_cash_priced_value_usd 启发式规则（稳定币仅计入 idle_cash_usd；不采集收入/储蓄率，user_provided 不参与）。
- S09：base_addition_usd 使用 clamp 与最低资金门槛；addition_amount_usd 四舍五入保留 2 位。
- 参数字段中的百分比（如 *_pct、sell_percentage）在 MVP 输出为小数（0.02 = 2%）；本文示例中的“30%/40%”仅用于说明，前端负责渲染为百分比字符串。
- 经验等级：本文统一使用 Expert（问卷选项为 Beginner/Intermediate/Expert）。
- 四舍五入规则以 prototypes 为准（crypto 8 位、stock/forex 2 位等）；文中 `round(..., 2)` 仅为说明。
- S16：资格判断使用 net_edge_pct = funding_rate_8h * (holding_period_hours / 8) + max(0, -basis_pct) - fee_pct，仅用于 eligibility，不输出净收益。
- S18：MVP 不支持做空，向下趋势仅允许减仓/降仓（reduce-only）。
- S22：引入 vol_floor=0.05 作为权重下限，避免极低波动导致权重失真。

 

 

 

**2. MVP阶段策略 (S01-S05)**

此部分内容来自 */home/ubuntu/strategies_detailed_part1.md*
（该文件不在 repo；MVP 参数与规则以 PRD + prototypes 为准。）

 

 

 

**3. V1.5阶段策略 (S06-S13)**

此部分内容来自 */home/ubuntu/strategies_detailed_part2.md*
（该文件不在 repo；MVP 参数与规则以 PRD + prototypes 为准。）

 

 

 

**4. V2.0阶段策略 (S14-S20)**

此部分内容来自 */home/ubuntu/strategies_detailed_part3.md* (V2.0部分)
（该文件不在 repo；MVP 参数与规则以 PRD + prototypes 为准。）

 

 

 

**5. V3.0阶段策略 (S21-S25)**

此部分内容来自 */home/ubuntu/strategies_detailed_part3.md* (V3.0部分)
（该文件不在 repo；MVP 参数与规则以 PRD + prototypes 为准。）

 

 

 

**6. 附录**

**6.1. 开发路线图与工作量估算**

| 阶段 | 策略数量 | 开发工作量 | 累计策略数 |
| :---- | :---- | :---- | :---- |
| MVP | 5个 | 19天 | 5 |
| V1.5 | +8个 | 26天 | 13 |
| V2.0 | +7个 | 34天 | 20 |
| V3.0 | +5个 | 30天 | 25 |
| 总计 | 25个 | 109天 | 25 |

注：以上为纯开发工作量，不包括测试、优化和文档编写时间。

 

**6.2. 关键成功指标 (KPI)**

| 阶段 | 指标 | 目标 |
| :---- | :---- | :---- |
| MVP | 策略采纳率 | > 30% |
| V1.5 | 自动化策略使用率 | > 40% |
| V2.0 | 一键执行使用率 | > 60% |
| V3.0 | 机构用户占比 | > 20% |

**Money Coach策略库详细设计 - Part 1: MVP阶段（S01-S05）**

**S01: 固定止损策略 (Stop Loss)**

**策略概述**

| 项目 | 内容 |
| :---- | :---- |
| 策略ID | S01 |
| 策略名称 | 固定止损 |
| 策略类型 | 风险管理 |
| 适用场景 | 所有持仓，尤其是新手和高波动资产 |
| 风险等级 | 低 |
| 实施阶段 | MVP (P0) |

**触发条件**

•   	普适推荐：适用于所有持仓

•   	强烈推荐：

◦   	用户是新手（经验 < 6个月）

◦   	单一资产浮亏 > 10%

◦   	单一资产集中度 > 40%

◦   	资产30日波动率 > 6%

 

**核心参数计算**

**1. 止损百分比**

def calculate_stop_loss_percentage(user_risk_level, asset_volatility, user_experience, current_loss_pct):  
	"""  
	计算个性化的止损百分比  
	"""  
	# 基础止损百分比  
	base_stop_loss = {  
    	"conservative": 0.05,   # 5%  
    	"moderate": 0.08,    	# 8%  
    	"aggressive": 0.12   	# 12%  
	}  
     
	base_sl = base_stop_loss[user_risk_level]  
     
	# 波动率调整  
	if asset_volatility > 0.06:  
    	volatility_adj = 1.5  
	elif asset_volatility > 0.04:  
    	volatility_adj = 1.2  
	else:  
    	volatility_adj = 1.0  
     
	# 经验调整  
	experience_adj = {  
    	"beginner": 0.8,  	# 新手更严格  
    	"intermediate": 1.0,  
    "expert": 1.2   	# 老手更宽松  
	}  
     
	# 如果已经亏损，适当放宽止损  
	if current_loss_pct < -0.10:  
    	loss_adj = 1.2  
	else:  
    	loss_adj = 1.0  
     
	stop_loss_pct = base_sl * volatility_adj * experience_adj[user_experience] * loss_adj  
     
	# 限制在3%-15%之间  
	return max(0.03, min(stop_loss_pct, 0.15))

 

参数表：

 

| 风险偏好 | 基础% | 高波动调整 | 新手调整 | 已亏损调整 |
| :---- | :---- | :---- | :---- | :---- |
| 保守 | 5% | ×1.5 | ×0.8 | ×1.2 |
| 平衡 | 8% | ×1.2 | ×0.8 | ×1.2 |
| 激进 | 12% | ×1.0 | ×1.2 | ×1.2 |

**2. 止损价格**

def calculate_stop_loss_price(current_price, avg_cost, stop_loss_pct, current_pnl_pct, support_levels):  
	"""  
	计算止损价格，结合技术支撑位验证  
	"""  
	# 如果已经亏损，从当前价和成本价中选择更宽松的  
	if current_pnl_pct < 0:  
    	stop_from_current = current_price * (1 - stop_loss_pct)  
    	stop_from_cost = avg_cost * (1 - stop_loss_pct)  
    	stop_loss_price = min(stop_from_current, stop_from_cost)  
	else:  
    	# 如果盈利，从成本价计算  
    	stop_loss_price = avg_cost * (1 - stop_loss_pct)  
     
	# 技术位验证：如果止损价在支撑位上方5%以内，调整到支撑位下方2%  
	closest_support = find_closest_support(stop_loss_price, support_levels)  
	if closest_support and stop_loss_price > closest_support * 0.95:  
    	stop_loss_price = closest_support * 0.98  
    	adjustment_reason = f"调整至技术支撑位${closest_support:.2f}下方"  
	else:  
    	adjustment_reason = "基于风险偏好和波动率计算"  
     
	return {  
    	"stop_loss_price": round(stop_loss_price, 2),  
    	"stop_loss_pct": stop_loss_pct,  
    	"adjustment_reason": adjustment_reason  
	}

 

**3. 动态调整规则**

def generate_stop_loss_adjustment_rules(avg_cost, stop_loss_price):  
	"""  
	生成止损的动态调整规则  
	"""  
	rules = [  
    	{  
        	"trigger": "盈利5%",  
        	"action": "上移止损至保本价",  
        	"new_stop_price": avg_cost,  
        	"reason": "保护本金"  
    	},  
    	{  
        	"trigger": "盈利10%",  
        	"action": "上移止损至锁定5%利润",  
        	"new_stop_price": avg_cost * 1.05,  
        	"reason": "锁定部分利润"  
    	},  
    	{  
        	"trigger": "盈利20%",  
        	"action": "转换为10%移动止盈策略",  
        	"new_strategy": "S03",  
        	"reason": "更灵活地保护利润"  
    	}  
	]  
	return rules

 

**回测验证**

**历史触发率分析**

def analyze_stop_loss_trigger_rate(asset_symbol, stop_loss_pct_range, lookback_days=365):  
	"""  
	分析不同止损百分比的历史触发率和假触发率  
	"""  
	historical_data = fetch_historical_data(asset_symbol, lookback_days)  
	results = []  
     
	for sl_pct in stop_loss_pct_range:  
    	trigger_count = 0  
    	false_trigger_count = 0  
         
    	# 模拟每天入场  
    	for entry_idx in range(len(historical_data) - 30):  
        	entry_price = historical_data[entry_idx]  
        	stop_price = entry_price * (1 - sl_pct)  
             
        	triggered = False  
        	recovered = False  
             
        	# 检查后续30天  
        	for future_idx in range(entry_idx + 1, min(entry_idx + 31, len(historical_data))):  
            	future_price = historical_data[future_idx]  
                 
            	if future_price <= stop_price and not triggered:  
                	triggered = True  
                	trigger_count += 1  
                 
            	# 检查是否反弹回本  
            	if triggered and future_price >= entry_price:  
                	recovered = True  
                	false_trigger_count += 1  
                	break  
         
    	total_entries = len(historical_data) - 30  
    	trigger_rate = trigger_count / total_entries  
    	false_trigger_rate = false_trigger_count / trigger_count if trigger_count > 0 else 0  
         
    	results.append({  
        	"stop_loss_pct": f"{sl_pct*100:.0f}%",  
        	"trigger_rate": f"{trigger_rate*100:.1f}%",  
        	"false_trigger_rate": f"{false_trigger_rate*100:.1f}%",  
        	"recommendation": "推荐" if 0.15 < trigger_rate < 0.30 and false_trigger_rate < 0.40 else "不推荐"  
    	})  
     
	return results

 

**AI文案生成要点**

•   	强调止损是"主动风控"而非"认输"

•   	用具体案例说明不止损的风险

•   	清晰量化潜在损失

•   	提供动态调整规则，避免"一设了之"

 

 

 

## S02: 马丁格尔解套策略 (Martingale DCA)

 策略ID: S02  
策略名称: 马丁格尔解套 (Martingale DCA)  
适用场景: 已亏损20%+，想降低成本、加速回本  
风险等级: 中高  
MVP优先级: P0（必须实现）

 

**2.1 金融原理**

马丁格尔策略源于赌博理论，核心思想是"越跌越买"。在加密货币投资中，当持仓资产价格下跌时，通过分批加仓（买入更多数量），可以显著降低平均持仓成本。这样，价格只需小幅反弹，就能实现盈利或回本。

 

关键优势：

•   	有效降低平均成本，减少回本所需涨幅

•   	适合震荡和反弹行情

•   	纪律化执行，避免情绪化决策

 

主要风险：

•   	如果价格持续单边下跌，会持续亏损并消耗大量资金

•   	需要预留足够的资金支持多次加仓

•   	不适合基本面恶化的资产

 

**2.2 触发条件判断**

必须同时满足：

1  	单一资产浮亏 ≥ 20%

2  	用户风险偏好 ≥ 中等（从问卷获取）

3  	该资产在用户持仓中占比 \< 60%（避免过度集中）

4  	用户有足够的闲置资金（至少为当前持仓价值的30%）

 

不推荐条件：

•   	资产处于明显下降趋势（50日MA \< 200日MA且持续下跌）

•   	项目基本面出现重大负面（需要AI新闻分析）

•   	用户风险偏好为保守型

 

代码示例：

def should\_recommend\_martingale(user\_data, market\_data):  
	"""  
	判断是否推荐马丁格尔策略  
	"""  
	conditions \= \[  
    	user\_data\['loss\_percentage'\] \<= \-0.20,  
    	user\_data\['risk\_level'\] in \['moderate', 'aggressive'\],  
    	user\_data\['asset\_concentration'\] \< 0.60,  
    	user\_data\['available\_funds'\] \>= user\_data\['current\_value'\] \* 0.30,  
    	market\_data\['ma\_50'\] \>= market\_data\['ma\_200'\]  \# 不在下降趋势  
	\]  
     
	return all(conditions)

 

**2.3 核心参数计算**

所有参数都基于用户画像和市场数据动态生成，确保个性化和安全性。

 

**2.3.1 价格步长 (Price Steps)**

定义：每次加仓触发的价格下跌百分比。

 

计算公式：

价格步长 \= 基础步长(风险偏好) × 波动率调整因子

 

参数表：

 

| 风险偏好 | 基础步长 | 高波动调整 | 低波动调整 |
| :---- | :---- | :---- | :---- |
| 保守 (Conservative) | 3.0% | ×1.2 | ×0.8 |
| 平衡 (Moderate) | 2.0% | ×1.2 | ×0.8 |
| 激进 (Aggressive) | 1.5% | ×1.2 | ×0.8 |

代码示例：

def calculate\_price\_steps(asset\_volatility, user\_risk\_level):  
	base\_steps \= {  
    	"conservative": 0.03,  
    	"moderate": 0.02,  
    	"aggressive": 0.015  
	}  
     
	if asset\_volatility \> 0.05:  \# 高波动  
    	adjustment \= 1.2  
	elif asset\_volatility \< 0.02:  \# 低波动  
    	adjustment \= 0.8  
	else:  
    	adjustment \= 1.0  
     
	price\_steps \= base\_steps\[user\_risk\_level\] \* adjustment  
	return round(price\_steps, 4)

 

示例：

•   	ETH波动率4%，用户moderate → 价格步长 \= 2.0% × 1.0 \= 2.0%

 

**2.3.2 最大加仓次数 (Max Safety Orders)**

定义：策略允许的最大加仓次数，用于控制总风险敞口。

 

计算逻辑：

5  	根据用户风险偏好，确定最大可追加投资比例（保守30%，平衡50%，激进80%）

6  	计算实际可用资金

7  	使用等比递增公式（倍数1.5），计算能执行几次加仓

8  	限制在3-6次之间

 

代码示例：

def calculate\_max\_safety\_orders(available\_funds, current\_investment, user\_risk\_level):  
	max\_additional\_ratio \= {  
    	"conservative": 0.3,  
    	"moderate": 0.5,  
    	"aggressive": 0.8  
	}  
     
	max\_additional\_funds \= current\_investment \* max\_additional\_ratio\[user\_risk\_level\]  
	actual\_available \= min(available\_funds, max\_additional\_funds)  
     
	initial\_order \= current\_investment \* 0.1  
	total\_allocated \= 0  
	safety\_orders \= 0  
	multiplier \= 1.5  
     
	for i in range(1, 10):  
    	order\_amount \= initial\_order \* (multiplier \*\* (i\-1))  
    	if total\_allocated \+ order\_amount \<= actual\_available:  
        	total\_allocated \+= order\_amount  
        	safety\_orders \+= 1  
    	else:  
        	break  
     
	return max(3, min(safety\_orders, 6))

 

**2.3.3 安全单金额 (Safety Order Amount)**

定义：每次加仓的基础金额（首次加仓金额）。

 

计算逻辑：

9  	基础金额 \= 当前投资额 × 10%

10   使用1.5倍递增，计算总需求

11   如果超出可用资金，等比缩小

 

代码示例：

def calculate\_safety\_order\_amount(current\_investment, max\_safety\_orders, available\_funds):  
	base\_percentage \= 0.1  
	initial\_order \= current\_investment \* base\_percentage  
	multiplier \= 1.5  
     
	total\_needed \= sum(\[initial\_order \* (multiplier \*\* i) for i in range(max\_safety\_orders)\])  
     
	if total\_needed \> available\_funds:  
    	scale\_factor \= available\_funds / total\_needed  
    	initial\_order \*= scale\_factor  
     
	return round(initial\_order, 2)

 

**2.3.4 其他参数**

| 参数 | 值 | 说明 |
| :---- | :---- | :---- |
| 金额倍数 | 1.5 | 固定值，经过验证的平衡参数 |
| 止盈目标 | 2-5% | 基于风险偏好和亏损幅度动态调整 |
| 总止损位 | 当前亏损 \+ 10-20% | 基于风险偏好，避免无限亏损 |

**2.4 加仓计划表生成**

输出示例：

 

| 加仓次序 | 触发价格 | 跌幅 | 加仓金额 | 加仓数量 | 新平均成本 | 累计投资 | 止盈价格 | 回本涨幅 |
| :---- | :---- | :---- | :---- | :---- | :---- | :---- | :---- | :---- |
| 第1次 | $2,940 | \-2.0% | $1,847 | 0.628 | $3,751 | $39,847 | $3,864 | 31.4% |
| 第2次 | $2,880 | \-4.0% | $2,771 | 0.962 | $3,674 | $42,618 | $3,784 | 31.4% |
| 第3次 | $2,820 | \-6.0% | $4,156 | 1.474 | $3,568 | $46,774 | $3,675 | 30.3% |
| 第4次 | $2,760 | \-8.0% | $6,234 | 2.258 | $3,423 | $53,008 | $3,526 | 27.8% |

代码示例：

def generate\_martingale\_plan(params):  
	plan \= \[\]  
	cumulative\_investment \= params\["total\_investment"\]  
	cumulative\_amount \= params\["holding\_amount"\]  
     
	for i in range(1, params\["max\_safety\_orders"\] \+ 1):  
    	trigger\_price \= params\["current\_price"\] \* (1 \- params\["price\_steps"\] \* i)  
    	order\_amount\_usd \= params\["safety\_order\_amount"\] \* (1.5 \*\* (i \- 1))  
    	order\_amount\_asset \= order\_amount\_usd / trigger\_price  
         
    	cumulative\_investment \+= order\_amount\_usd  
    	cumulative\_amount \+= order\_amount\_asset  
    	new\_avg\_cost \= cumulative\_investment / cumulative\_amount  
    	take\_profit\_price \= new\_avg\_cost \* (1 \+ params\["take\_profit"\])  
    	breakeven\_from\_trigger \= (new\_avg\_cost \- trigger\_price) / trigger\_price  
         
    	plan.append({  
        	"order\_number": i,  
        	"trigger\_price": round(trigger\_price, 2),  
        	"drop\_from\_current": f"{params\['price\_steps'\] \* i \* 100:.1f}%",  
        	"order\_amount\_usd": round(order\_amount\_usd, 2),  
        	"order\_amount\_asset": round(order\_amount\_asset, 4),  
        	"new\_avg\_cost": round(new\_avg\_cost, 2),  
        	"cumulative\_investment": round(cumulative\_investment, 2),  
        	"cumulative\_amount": round(cumulative\_amount, 4),  
        	"take\_profit\_price": round(take\_profit\_price, 2),  
        	"breakeven\_from\_trigger": f"{breakeven\_from\_trigger \* 100:.1f}%"  
    	})  
     
	return plan

 

**2.5 回测验证**

**2.5.1 历史回测**

使用过去180天的日线数据，模拟执行生成的策略参数。

 

关键指标：

•   	胜率：策略最终盈利的概率

•   	平均回本周期：从入场到止盈的平均天数

•   	最大回撤：策略执行过程中的最大浮亏

•   	年化收益率：折算为年化的收益率

 

代码示例：

def backtest\_martingale\_strategy(historical\_data, entry\_price, entry\_date, strategy\_params):  
	position \= {  
    	"total\_investment": strategy\_params\["initial\_order\_amount"\],  
    	"total\_amount": strategy\_params\["initial\_order\_amount"\] / entry\_price,  
    	"avg\_cost": entry\_price,  
    	"safety\_orders\_executed": 0,  
    	"status": "active"  
	}  
     
	\# 遍历历史数据，检查加仓和止盈/止损触发  
	\# ... (详细实现见附录)  
     
	return {  
    	"position": position,  
    	"trades": trades,  
    	"performance": {  
        	"final\_pnl\_pct": "8.5%",  
        	"holding\_days": 42,  
        	"annualized\_return": "74.2%",  
        	"max\_drawdown": "-12.3%",  
        	"win\_rate": "68%"  
    	}  
	}

 

**2.5.2 蒙特卡洛模拟**

运行1000次基于历史波动率的随机价格路径模拟，给出策略的概率性表现。

 

输出示例：

•   	胜率：68%

•   	平均盈亏：+3.2%

•   	95%置信区间：\[-15%, \+25%\]

•   	最坏情况：-28%

 

**2.6 风险评估**

综合风险评分：0-100分，100为最高风险。

 

评分因子：

•   	策略固有风险：70分（马丁格尔本身是高风险策略）

•   	市场环境风险：+15分（如果波动率\>6%）

•   	用户适配性风险：+15分（如果是新手）

•   	资金充足性风险：+25分（如果资金不足）

 

风险等级：

•   	0-30分：低风险（绿色）

•   	30-60分：中等风险（黄色）

•   	60-100分：高风险（红色）

 

**2.7 AI文案生成**

System Prompt：

你是Quanta AI的专业投资策略顾问。你的任务是将马丁格尔解套策略的技术参数转化为用户友好的、有说服力的策略建议。  
   
核心原则：  
1\. 使用通俗易懂的语言，避免专业术语堆砌  
2\. 强调策略的实际效果（如"降低成本X%"），而非技术细节  
3\. 必须包含清晰的风险提示，使用显著的格式标注  
4\. 文案风格：专业但不冰冷，理性但有温度  
5\. 长度控制：概览100-150字，详情300-500字  
   
禁止事项：  
1\. 不要保证收益或暗示"稳赚不赔"  
2\. 不要使用过于激进的语言（如"暴富"、"必胜"）  
3\. 不要忽视或淡化风险

 

User Prompt 模板：

请为以下用户生成马丁格尔解套策略的详细说明文案。  
   
用户情况：  
\- 资产：ETH  
\- 当前价格：$3,000  
\- 持仓数量：10  
\- 平均成本：$3,800  
\- 当前浮亏：-21.1%（-$8,000）  
\- 可用资金：$15,000  
\- 风险偏好：平衡型  
   
策略参数（已优化）：  
\- 价格步长：每下跌2.0%加仓一次  
\- 最大加仓次数：4次  
\- ...  
   
请生成：  
1\. 策略概览卡片文案（100-150字）  
2\. 策略详情页文案（300-500字）  
3\. AI个性化建议（80-120字）

 

生成文案示例：

 

*策略概览卡片*

 

*马丁格尔解套：降低成本，加速回本*

 

*您的ETH当前浮亏21%（-$8,000），通过马丁格尔策略，在价格下跌时分4次加仓，可将平均成本从$3,800降至$3,423（降低10%），回本所需涨幅从27%减少至14%。历史回测显示，该策略在类似情况下的胜率为68%，平均42天回本。*

 

*⚠️ 风险等级：中高风险 | 需追加投资$15,000*

 

 

 

 

## S03: 移动止盈策略 (Trailing Stop)

策略名称: 移动止盈 (Trailing Stop)  
适用场景: 已盈利15%+，想锁定利润但不想过早离场  
风险等级: 低  
MVP优先级: P0（必须实现）

 

**1 金融原理**

移动止盈是一种动态的止损/止盈工具。它会随着价格向有利方向移动而自动上调止盈价格，但当价格反向移动时保持不变。这种机制允许交易者在锁定大部分利润的同时，继续享受价格上涨带来的收益，避免"坐电梯"（盈利变亏损）。

 

关键优势：

•   	让利润奔跑，不会过早离场

•   	自动保护已有利润

•   	纪律化执行，避免贪婪

 

主要风险：

•   	高波动市场中，可能被短期回调触发

•   	参数设置不当会影响利润捕获率

 

**2 触发条件判断**

必须同时满足：

1  	单一资产浮盈 ≥ 15%

2  	资产处于上升趋势（价格 \> 20日MA \> 50日MA）

 

**3 核心参数计算**

**3.1 移动百分比**

计算公式：

移动百分比 \= 基础百分比(风险偏好) × 波动率因子 × 趋势强度因子 × 盈利幅度因子

 

参数表：

 

| 风险偏好 | 基础百分比 | 强趋势调整 | 大盈利调整 |
| :---- | :---- | :---- | :---- |
| 保守 | 8% | ×1.2 | ×1.3 (盈利\>50%) |
| 平衡 | 10% | ×1.2 | ×1.1 (盈利\>30%) |
| 激进 | 15% | ×1.2 | ×1.0 |

代码示例：

def calculate\_trailing\_stop\_percentage(profit\_pct, asset\_volatility, user\_risk\_level, trend\_strength):  
	base\_trailing\_pct \= {  
    	"conservative": 0.08,  
    	"moderate": 0.10,  
    	"aggressive": 0.15  
	}  
     
	base\_pct \= base\_trailing\_pct\[user\_risk\_level\]  
     
	if asset\_volatility \> 0.05:  
    	volatility\_adj \= 1.3  
	elif asset\_volatility \< 0.02:  
    	volatility\_adj \= 0.8  
	else:  
    	volatility\_adj \= 1.0  
     
	trend\_adj \= {  
    	"weak": 0.8,  
    	"moderate": 1.0,  
    	"strong": 1.2  
	}  
     
	if profit\_pct \> 0.5:  
    	profit\_adj \= 1.3  
	elif profit\_pct \> 0.3:  
    	profit\_adj \= 1.1  
	else:  
    	profit\_adj \= 1.0  
     
	trailing\_pct \= base\_pct \* volatility\_adj \* trend\_adj\[trend\_strength\] \* profit\_adj  
     
	return max(0.05, min(trailing\_pct, 0.25))

 

示例：

•   	BTC盈利27%，波动率3.5%，用户moderate，强趋势 → 移动百分比 \= 10% × 1.0 × 1.2 × 1.0 \= 12%

 

**3.2 初始止盈价格**

initial\_trailing\_stop\_price \= current\_price \* (1 \- trailing\_pct)

 

示例：

•   	当前价$95,000，移动12% → 初始止盈价 \= $83,600

 

**3.3 保护利润计算**

protected\_profit \= (trailing\_stop\_price \- avg\_cost) \* holding\_amount  
protection\_ratio \= protected\_profit / current\_profit

 

示例：

•   	止盈价$83,600，成本$75,000，持有0.5 BTC

•   	保护利润 \= ($83,600 \- $75,000) × 0.5 \= $4,300

•   	当前利润$10,000，保护比例 \= 43%

 

**4 分层移动止盈方案**

将持仓分为3层，设置不同的移动百分比，平衡收益和保护。

 

示例（平衡型用户）：

 

| 层级 | 仓位占比 | 移动百分比 | 初始止盈价 | 说明 |
| :---- | :---- | :---- | :---- | :---- |
| 保守层 | 30% | 5% | $90,250 | 快速锁定部分利润 |
| 平衡层 | 40% | 10% | $85,500 | 平衡收益和保护 |
| 激进层 | 30% | 15% | $80,750 | 追求最大收益 |

**5 回测验证**

参数优化回测：回测过去1年，测试不同移动百分比的利润捕获率。

 

示例输出：

 

| 移动百分比 | 利润捕获率 | 平均持有天数 | 推荐 |
| :---- | :---- | :---- | :---- |
| 5% | 42% | 12天 | ⚠️ 过早退出 |
| 10% | 68% | 28天 | ✅ 推荐 |
| 15% | 81% | 45天 | ✅ 推荐 |
| 20% | 89% | 62天 | ⚠️ 可能回吐较多 |

**6 AI文案生成**

文案示例：

 

*移动止盈：让您的利润持续增长*

 

*恭喜！您的BTC已盈利27%（+$10,000）。启动12%的移动止盈策略，就像给您的利润上了一份保险。价格继续涨，您的收益跟着涨；价格一旦回调12%，系统将自动卖出，为您至少锁定$4,300（11.5%）的利润，避免"坐电梯"的烦恼。*

 

*根据历史回测，该移动百分比在类似情况下的利润捕获率为68%，平均持有28天后退出。您也可以选择分层方案，30%仓位设置5%移动止盈，70%仓位设置12%移动止盈，更加灵活。*

 

 

**S04: 分层止盈策略 (Layered Take Profit)**

**策略概述**

| 项目 | 内容 |
| :---- | :---- |
| 策略ID | S04 |
| 策略名称 | 分层止盈 |
| 策略类型 | 风险管理 |
| 适用场景 | 浮盈≥30%，想分批退出，平衡收益和保护 |
| 风险等级 | 低 |
| 实施阶段 | MVP (P0) |

**触发条件**

•   	必须满足：

◦   	单一资产浮盈 ≥ 30%

◦   	持仓时间 > 30天（避免短期投机）

•   	推荐满足：

◦   	用户风险偏好为保守或平衡

◦   	该资产在组合中占比 > 30%

 

**核心参数计算**

**1. 分层方案设计**

def generate_layered_take_profit_plan(user_risk_level, current_profit_pct, holding_amount, current_price, avg_cost):  
	"""  
	生成分层止盈方案  
	"""  
	# 根据风险偏好确定分层策略  
	layer_configs = {  
    	"conservative": {  
        	"layers": [  
            	{"name": "第一层", "percentage": 0.40, "target_profit": 1.30, "reason": "快速锁定40%仓位利润"},  
            	{"name": "第二层", "percentage": 0.35, "target_profit": 1.50, "reason": "中期目标，锁定35%仓位"},  
            	{"name": "第三层", "percentage": 0.25, "target_profit": 1.80, "reason": "长期持有，追求更高收益"}  
        	]  
    	},  
    	"moderate": {  
        	"layers": [  
            	{"name": "第一层", "percentage": 0.30, "target_profit": 1.40, "reason": "部分获利了结"},  
            	{"name": "第二层", "percentage": 0.40, "target_profit": 1.70, "reason": "主力仓位退出"},  
            	{"name": "第三层", "percentage": 0.30, "target_profit": 2.00, "reason": "博取更大收益"}  
        	]  
    	},  
    	"aggressive": {  
        	"layers": [  
            	{"name": "第一层", "percentage": 0.20, "target_profit": 1.50, "reason": "少量获利了结"},  
            	{"name": "第二层", "percentage": 0.30, "target_profit": 2.00, "reason": "中期退出"},  
            	{"name": "第三层", "percentage": 0.50, "target_profit": 3.00, "reason": "长期持有大部分仓位"}  
        	]  
    	}  
	}  
     
	config = layer_configs[user_risk_level]  
	plan = []  
     
	for layer in config["layers"]:  
    	sell_amount = holding_amount * layer["percentage"]  
    	target_price = avg_cost * layer["target_profit"]  
    	expected_profit = (target_price - avg_cost) * sell_amount  
         
    	# 如果当前价格已经超过目标价，调整目标价  
    	if current_price >= target_price:  
        	target_price = current_price * 1.05  # 在当前价上方5%  
         
    	plan.append({  
        	"layer_name": layer["name"],  
        	"sell_percentage": f"{layer['percentage']*100:.0f}%",  
        	"sell_amount": round(sell_amount, 4),  
        	"target_price": round(target_price, 2),  
        	"target_profit_pct": f"{(layer['target_profit']-1)*100:.0f}%",  
        	"expected_profit_usd": round(expected_profit, 2),  
        	"reason": layer["reason"]  
    	})  
     
	return plan

 

示例输出：

 

| 层级 | 卖出比例 | 卖出数量 | 目标价格 | 目标涨幅 | 预期利润 | 说明 |
| :---- | :---- | :---- | :---- | :---- | :---- | :---- |
| 第一层 | 30% | 3 ETH | $4,200 | +40% | $3,600 | 部分获利了结 |
| 第二层 | 40% | 4 ETH | $5,100 | +70% | $6,400 | 主力仓位退出 |
| 第三层 | 30% | 3 ETH | $6,000 | +100% | $5,400 | 博取更大收益 |

**2. 动态调整机制**

def generate_layered_adjustment_rules(plan, current_price):  
	"""  
	生成分层止盈的动态调整规则  
	"""  
	rules = []  
     
	for i, layer in enumerate(plan):  
    	# 如果第一层已触发，第二层自动上移止损  
    	if i > 0:  
        	previous_layer_price = plan[i-1]["target_price"]  
        	rules.append({  
            	"trigger": f"{plan[i-1]['layer_name']}卖出完成",  
            	"action": f"为{layer['layer_name']}设置保护性止损",  
            	"stop_price": previous_layer_price * 0.95,  
            	"reason": "保护已实现的部分利润"  
        	})  
     
	return rules

 

**回测验证**

def backtest_layered_take_profit(historical_data, entry_price, entry_date, plan):  
	"""  
	回测分层止盈策略  
	"""  
	executed_layers = []  
	remaining_amount = sum([layer["sell_amount"] for layer in plan])  
     
	for i, price in enumerate(historical_data[entry_date:]):  
    	current_date = entry_date + i  
         
    	# 检查每一层是否触发  
    	for layer in plan:  
        	if price >= layer["target_price"] and layer not in executed_layers:  
            	executed_layers.append({  
                	"layer": layer["layer_name"],  
                	"date": current_date,  
                	"price": price,  
                	"amount": layer["sell_amount"],  
                	"profit": (price - entry_price) * layer["sell_amount"]  
            	})  
            	remaining_amount -= layer["sell_amount"]  
     
	# 计算总利润  
	total_profit = sum([l["profit"] for l in executed_layers])  
     
	# 计算利润捕获率  
	max_price = max(historical_data[entry_date:])  
	max_potential_profit = (max_price - entry_price) * sum([layer["sell_amount"] for layer in plan])  
	profit_capture_rate = total_profit / max_potential_profit if max_potential_profit > 0 else 0  
     
	return {  
    	"executed_layers": executed_layers,  
    	"total_profit": round(total_profit, 2),  
    	"profit_capture_rate": f"{profit_capture_rate*100:.1f}%",  
    	"remaining_amount": remaining_amount  
	}

 

**AI文案生成要点**

•   	祝贺用户的盈利成果

•   	强调分层是"智慧的选择"，而非"贪婪"或"胆小"

•   	用具体数字展示每一层的目标和收益

•   	提供灵活性，用户可以自定义分层比例

 

 

 

**S05: 定投建议策略 (Dollar Cost Averaging - DCA)**

**策略概述**

| 项目 | 内容 |
| :---- | :---- |
| 策略ID | S05 |
| 策略名称 | 定投建议（DCA） |
| 策略类型 | 自动化交易 |
| 适用场景 | 看好长期价值，想分散风险，避免择时 |
| 风险等级 | 低 |
| 实施阶段 | MVP (P0) |

**触发条件**

•   	推荐场景：

◦   	用户表示看好某资产长期价值

◦   	用户是新手，不擅长择时

◦   	用户有稳定的现金流

◦   	资产波动率 > 4%（定投在高波动中更有效）

 

**核心参数计算**

**1. 定投金额**

def calculate_dca_amount(user_monthly_income, user_savings_rate, user_risk_level, asset_volatility):  
	"""  
	计算建议的定投金额  
	"""  
	# 基于月收入和储蓄率  
	monthly_investable = user_monthly_income * user_savings_rate  
     
	# 根据风险偏好分配到加密货币的比例  
	crypto_allocation = {  
    	"conservative": 0.10,   # 10%  
    	"moderate": 0.20,    	# 20%  
    	"aggressive": 0.40   	# 40%  
	}  
     
	monthly_crypto_budget = monthly_investable * crypto_allocation[user_risk_level]  
     
	# 如果用户已有持仓，考虑当前持仓价值  
	if user_has_holdings:  
    	# 避免过度集中  
    	max_single_asset_allocation = monthly_crypto_budget * 0.5  
	else:  
    	max_single_asset_allocation = monthly_crypto_budget  
     
	# 波动率调整：高波动资产，建议更小的单次金额，更高的频率  
	if asset_volatility > 0.06:  
    	frequency_adj = 2  # 改为每两周一次  
    	amount_adj = 0.5  
	else:  
    	frequency_adj = 1  # 每月一次  
    	amount_adj = 1.0  
     
	dca_amount = max_single_asset_allocation * amount_adj  
     
	return {  
    	"recommended_amount": round(dca_amount, 2),  
    	"frequency": "每两周" if frequency_adj == 2 else "每月",  
    	"total_annual_investment": round(dca_amount * (12 / frequency_adj), 2)  
	}

 

示例输出：

•   	月收入：$5,000

•   	储蓄率：30%

•   	风险偏好：平衡型

•   	→ 建议定投金额：$150/月

 

**2. 定投频率**

def calculate_dca_frequency(asset_volatility, user_preference):  
	"""  
	计算建议的定投频率  
	"""  
	# 基于波动率  
	if asset_volatility > 0.06:  
    	recommended_freq = "每两周"  
    	reason = "高波动资产，更频繁的定投可以更好地平滑成本"  
	elif asset_volatility > 0.04:  
    	recommended_freq = "每月"  
    	reason = "中等波动，月度定投平衡了成本和手续费"  
	else:  
    	recommended_freq = "每月"  
    	reason = "低波动资产，月度定投即可"  
     
	# 考虑用户偏好  
	if user_preference == "weekly":  
    	recommended_freq = "每周"  
    	reason += "，根据您的偏好调整为每周"  
     
	return {  
    	"frequency": recommended_freq,  
    	"reason": reason  
	}

 

**3. 定投时间点**

def calculate_dca_timing(asset_symbol, historical_data):  
	"""  
	分析最佳定投时间点（每月的哪一天）  
	"""  
	# 分析历史数据，找出每月价格相对较低的日期  
	monthly_price_patterns = analyze_monthly_patterns(historical_data)  
     
	# 找出平均价格最低的几天  
	best_days = sorted(monthly_price_patterns.items(), key=lambda x: x[1])[:3]  
     
	return {  
    	"recommended_days": [day for day, _ in best_days],  
    	"reason": f"历史数据显示，每月{best_days[0][0]}号左右价格相对较低",  
    	"note": "这只是历史统计，不保证未来表现"  
	}

 

**4. 定投计划表**

def generate_dca_plan(dca_amount, frequency, start_date, duration_months=12):  
	"""  
	生成未来12个月的定投计划表  
	"""  
	plan = []  
     
	if frequency == "每月":  
    	intervals = 1  
	elif frequency == "每两周":  
    	intervals = 0.5  
	elif frequency == "每周":  
    	intervals = 0.25  
     
	total_investments = int(duration_months / intervals)  
     
	for i in range(total_investments):  
    	investment_date = start_date + timedelta(days=int(i * intervals * 30))  
    	plan.append({  
        	"period": i + 1,  
        	"date": investment_date.strftime("%Y-%m-%d"),  
        	"amount": dca_amount,  
        	"cumulative_investment": dca_amount * (i + 1)  
    	})  
     
	return plan

 

**回测验证**

def backtest_dca_strategy(historical_data, dca_amount, frequency, start_date, duration_months):  
	"""  
	回测定投策略的历史表现  
	"""  
	total_invested = 0  
	total_amount = 0  
	investments = []  
     
	# 根据频率确定投资间隔  
	interval_days = {  
    	"每月": 30,  
    	"每两周": 14,  
    	"每周": 7  
	}  
     
	days_interval = interval_days[frequency]  
	current_date = start_date  
	end_date = start_date + timedelta(days=duration_months * 30)  
     
	while current_date <= end_date:  
    	# 获取当天价格  
    	price = get_price_on_date(historical_data, current_date)  
         
    	# 执行定投  
    	amount_bought = dca_amount / price  
    	total_invested += dca_amount  
    	total_amount += amount_bought  
         
    	investments.append({  
        	"date": current_date,  
        	"price": price,  
        	"invested": dca_amount,  
        	"amount_bought": amount_bought  
    	})  
         
    	current_date += timedelta(days=days_interval)  
     
	# 计算最终收益  
	final_price = historical_data[-1]  
	final_value = total_amount * final_price  
	total_profit = final_value - total_invested  
	roi = (final_value / total_invested - 1) * 100  
     
	# 计算平均成本  
	avg_cost = total_invested / total_amount  
     
	# 对比一次性投资  
	lump_sum_amount = total_invested / (get_price_on_date(historical_data, start_date))  
	lump_sum_value = lump_sum_amount * final_price  
	lump_sum_roi = (lump_sum_value / total_invested - 1) * 100  
     
	return {  
    	"total_invested": round(total_invested, 2),  
    	"total_amount": round(total_amount, 4),  
    	"avg_cost": round(avg_cost, 2),  
    	"final_value": round(final_value, 2),  
    	"total_profit": round(total_profit, 2),  
    	"roi": f"{roi:.2f}%",  
    	"lump_sum_roi": f"{lump_sum_roi:.2f}%",  
    	"dca_advantage": f"{roi - lump_sum_roi:+.2f}%",  
    	"investments_count": len(investments)  
	}

 

**AI文案生成要点**

•   	强调定投是"纪律投资"，避免情绪化决策

•   	用历史回测数据展示定投的平滑成本效果

•   	对比一次性投资和定投的差异

•   	提供灵活的频率和金额选择

•   	强调长期坚持的重要性

 

 

 

**MVP阶段策略总结**

| 策略 | 核心价值 | 技术复杂度 | 开发工作量 |
| :---- | :---- | :---- | :---- |
| S01 固定止损 | 风险控制基础 | 低 | 3天 |
| S02 马丁格尔解套 | 降低成本，加速回本 | 高 | 5天 |
| S03 移动止盈 | 保护利润，让利润奔跑 | 中 | 4天 |
| S04 分层止盈 | 灵活退出，平衡收益 | 中 | 4天 |
| S05 定投建议 | 纪律投资，平滑成本 | 低 | 3天 |

总计：5个策略，19天开发工作量（含联调测试）

**Money Coach策略库详细设计 - Part 2: V1.5阶段（S06-S13）**

**S06: 现货网格策略 (Spot Grid Trading)**

**策略概述**

| 项目 | 内容 |
| :---- | :---- |
| 策略ID | S06 |
| 策略名称 | 现货网格 |
| 策略类型 | 自动化交易 |
| 适用场景 | 震荡行情，价格区间明确，想自动化高抛低吸 |
| 风险等级 | 中 |
| 实施阶段 | V1.5 (P1) |

**触发条件**

•   	市场环境：

◦   	资产处于震荡行情（布林带宽度 < 0.15）

◦   	近30天价格在±20%区间内波动

◦   	无明显单边趋势（价格围绕MA50上下波动）

•   	用户条件：

◦   	有足够资金支持网格（至少当前持仓价值的50%）

◦   	风险偏好≥平衡型

◦   	不需要短期流动性

 

**核心参数计算**

**1. 网格区间**

def calculate_grid_range(current_price, historical_data, user_risk_level):  
	"""  
	计算网格的价格区间（上限和下限）  
	"""  
	# 分析历史波动范围  
	high_30d = max(historical_data[-30:])  
	low_30d = min(historical_data[-30:])  
	volatility_range = (high_30d - low_30d) / current_price  
     
	# 根据风险偏好调整区间宽度  
	range_multiplier = {  
    	"conservative": 0.8,   # 更窄的区间  
    	"moderate": 1.0,  
    	"aggressive": 1.2  	# 更宽的区间  
	}  
     
	adjusted_range = volatility_range * range_multiplier[user_risk_level]  
     
	# 计算上下限  
	grid_lower = current_price * (1 - adjusted_range / 2)  
	grid_upper = current_price * (1 + adjusted_range / 2)  
     
	# 验证区间合理性  
	if (grid_upper - grid_lower) / current_price < 0.10:  
    	# 区间太窄，扩大到至少10%  
    	grid_lower = current_price * 0.95  
    	grid_upper = current_price * 1.05  
     
	return {  
    	"grid_lower": round(grid_lower, 2),  
    	"grid_upper": round(grid_upper, 2),  
    	"range_pct": f"{(grid_upper - grid_lower) / current_price * 100:.1f}%",  
    	"reason": f"基于近30天波动范围（{volatility_range*100:.1f}%）计算"  
	}

 

**2. 网格数量**

def calculate_grid_count(grid_range_pct, asset_volatility, available_funds, min_order_size):  
	"""  
	计算最优的网格数量  
	"""  
	# 基于波动率确定基础网格数  
	if asset_volatility > 0.06:  
    	base_grids = 15  # 高波动，更多网格  
	elif asset_volatility > 0.04:  
    	base_grids = 10  
	else:  
    	base_grids = 7   # 低波动，较少网格  
     
	# 基于区间宽度调整  
	if grid_range_pct > 0.30:  
    	base_grids = int(base_grids * 1.5)  
     
	# 验证资金是否足够  
	grid_spacing = grid_range_pct / base_grids  
	min_required_funds = min_order_size * base_grids  
     
	if available_funds < min_required_funds:  
    	# 资金不足，减少网格数  
    	base_grids = int(available_funds / min_order_size)  
     
	# 限制在5-30之间  
	return max(5, min(base_grids, 30))

 

**3. 每格投资金额**

def calculate_grid_order_size(available_funds, grid_count, current_holdings_value):  
	"""  
	计算每个网格的订单金额  
	"""  
	# 总投资金额 = 可用资金 + 当前持仓价值的一半（用于卖出网格）  
	total_investment = available_funds + current_holdings_value * 0.5  
     
	# 平均分配到每个网格  
	order_size = total_investment / grid_count  
     
	# 确保不低于最小订单金额（如$10）  
	min_order = 10  
	if order_size < min_order:  
    	order_size = min_order  
    	grid_count = int(total_investment / min_order)  
     
	return {  
    	"order_size_usd": round(order_size, 2),  
    	"total_investment": round(total_investment, 2),  
    	"adjusted_grid_count": grid_count  
	}

 

**4. 网格计划表**

def generate_grid_plan(grid_lower, grid_upper, grid_count, order_size, current_price):  
	"""  
	生成完整的网格买卖计划表  
	"""  
	grid_spacing = (grid_upper - grid_lower) / grid_count  
	plan = []  
     
	for i in range(grid_count + 1):  
    	grid_price = grid_lower + i * grid_spacing  
         
    	# 判断是买单还是卖单  
    	if grid_price < current_price:  
        	order_type = "买入"  
        	action = f"当价格跌至${grid_price:.2f}时买入"  
    	elif grid_price > current_price:  
        	order_type = "卖出"  
        	action = f"当价格涨至${grid_price:.2f}时卖出"  
    	else:  
        	order_type = "当前价"  
        	action = "参考价格"  
         
    	# 计算该网格的预期利润  
    	if i > 0:  
        	profit_per_grid = grid_spacing / grid_price * order_size  
    	else:  
        	profit_per_grid = 0  
         
    	plan.append({  
        	"grid_number": i + 1,  
        	"price": round(grid_price, 2),  
        	"order_type": order_type,  
        	"order_size_usd": order_size,  
        	"action": action,  
        	"profit_per_cycle": round(profit_per_grid, 2)  
    	})  
     
	return plan

 

示例输出：

 

| 网格# | 价格 | 类型 | 金额 | 操作 | 单次利润 |
| :---- | :---- | :---- | :---- | :---- | :---- |
| 1 | $2,850 | 买入 | $200 | 价格跌至此买入 | - |
| 2 | $2,925 | 买入 | $200 | 价格跌至此买入 | $5.26 |
| 3 | $3,000 | 当前价 | - | 参考价格 | - |
| 4 | $3,075 | 卖出 | $200 | 价格涨至此卖出 | $4.88 |
| 5 | $3,150 | 卖出 | $200 | 价格涨至此卖出 | $4.76 |

**回测验证**

def backtest_grid_strategy(historical_data, grid_plan, start_date, duration_days):  
	"""  
	回测网格策略  
	"""  
	executed_trades = []  
	total_profit = 0  
	active_orders = {grid["price"]: grid for grid in grid_plan}  
     
	for i, price in enumerate(historical_data[start_date:start_date+duration_days]):  
    	# 检查是否触发任何网格  
    	for grid_price, grid in list(active_orders.items()):  
        	if grid["order_type"] == "买入" and price <= grid_price:  
            	# 执行买入  
            	executed_trades.append({  
                	"date": start_date + i,  
                	"type": "buy",  
                	"price": price,  
                	"amount_usd": grid["order_size_usd"]  
            	})  
            	# 在上方网格设置卖单  
            	sell_price = grid_price + (grid_plan[1]["price"] - grid_plan[0]["price"])  
            	active_orders[sell_price] = {  
                	"order_type": "卖出",  
                	"order_size_usd": grid["order_size_usd"],  
                	"price": sell_price  
            	}  
                 
        	elif grid["order_type"] == "卖出" and price >= grid_price:  
            	# 执行卖出  
            	profit = grid["profit_per_cycle"]  
            	total_profit += profit  
            	executed_trades.append({  
                	"date": start_date + i,  
                	"type": "sell",  
                	"price": price,  
                	"amount_usd": grid["order_size_usd"],  
                	"profit": profit  
            	})  
            	# 在下方网格设置买单  
            	buy_price = grid_price - (grid_plan[1]["price"] - grid_plan[0]["price"])  
            	active_orders[buy_price] = {  
                	"order_type": "买入",  
                	"order_size_usd": grid["order_size_usd"],  
                	"price": buy_price  
            	}  
     
	# 计算收益率  
	total_investment = sum([grid["order_size_usd"] for grid in grid_plan])  
	roi = total_profit / total_investment * 100  
	annualized_roi = roi * (365 / duration_days)  
     
	return {  
    	"total_trades": len(executed_trades),  
    	"total_profit": round(total_profit, 2),  
    	"roi": f"{roi:.2f}%",  
    	"annualized_roi": f"{annualized_roi:.2f}%",  
    	"avg_profit_per_trade": round(total_profit / len(executed_trades), 2) if executed_trades else 0  
	}

 

 

 

**S07: 定时定额卖出策略 (Scheduled Selling)**

**策略概述**

| 项目 | 内容 |
| :---- | :---- |
| 策略ID | S07 |
| 策略名称 | 定时定额卖出 |
| 策略类型 | 自动化交易 |
| 适用场景 | 盈利退出，避免一次性卖出，分散市场风险 |
| 风险等级 | 低 |
| 实施阶段 | V1.5 (P1) |

**触发条件**

•   	推荐场景：

◦   	用户想退出某资产（盈利或止损）

◦   	持仓数量较大，一次性卖出可能影响市场

◦   	用户担心卖出后价格继续上涨（后悔风险）

 

**核心参数计算**

**1. 卖出金额和频率**

def calculate_selling_schedule(total_amount, target_duration_days, user_urgency):  
	"""  
	计算定时卖出的金额和频率  
	"""  
	# 根据紧急程度确定卖出周期  
	duration_config = {  
    	"urgent": 7,    	# 1周内卖完  
    	"moderate": 30, 	# 1个月内卖完  
    	"relaxed": 90   	# 3个月内卖完  
	}  
     
	duration = duration_config.get(user_urgency, target_duration_days)  
     
	# 确定卖出频率  
	if duration <= 7:  
    	frequency = "每天"  
    	intervals = duration  
	elif duration <= 30:  
    	frequency = "每3天"  
    	intervals = duration // 3  
	else:  
    	frequency = "每周"  
    	intervals = duration // 7  
     
	# 计算每次卖出金额  
	amount_per_sell = total_amount / intervals  
     
	return {  
    	"total_amount": total_amount,  
    	"duration_days": duration,  
    	"frequency": frequency,  
    	"intervals": intervals,  
    	"amount_per_sell": round(amount_per_sell, 4),  
    	"reason": f"分{intervals}次卖出，降低市场冲击和后悔风险"  
	}

 

**2. 卖出计划表**

def generate_selling_schedule(start_date, schedule_params):  
	"""  
	生成详细的卖出计划表  
	"""  
	plan = []  
	frequency_days = {  
    	"每天": 1,  
    	"每3天": 3,  
    	"每周": 7  
	}  
     
	days_interval = frequency_days[schedule_params["frequency"]]  
     
	for i in range(schedule_params["intervals"]):  
    	sell_date = start_date + timedelta(days=i * days_interval)  
    	plan.append({  
        	"period": i + 1,  
        	"date": sell_date.strftime("%Y-%m-%d"),  
        	"sell_amount": schedule_params["amount_per_sell"],  
        	"cumulative_sold": schedule_params["amount_per_sell"] * (i + 1),  
        	"remaining": schedule_params["total_amount"] - schedule_params["amount_per_sell"] * (i + 1)  
    	})  
     
	return plan

 

 

 

**S08: 价值平均策略 (Value Averaging)**

**策略概述**

| 项目 | 内容 |
| :---- | :---- |
| 策略ID | S08 |
| 策略名称 | 价值平均策略（VA） |
| 策略类型 | 自动化交易 |
| 适用场景 | 长期投资，比DCA更智能，根据市场表现动态调整投资额 |
| 风险等级 | 低 |
| 实施阶段 | V1.5 (P1) |

**核心逻辑**

价值平均策略（VA）是DCA的升级版。与DCA固定投资金额不同，VA设定一个目标价值增长路径，每期根据实际价值与目标价值的差距，动态调整投资金额。

 

示例：

•   	目标：每月组合价值增长$500

•   	第1个月：投资$500，价值$500

•   	第2个月：如果价值涨到$600，只需投资$400即可达到目标$1,000

•   	第3个月：如果价值跌到$900，需投资$600才能达到目标$1,500

 

**核心参数计算**

def calculate_va_strategy(target_monthly_growth, current_portfolio_value, months_elapsed, current_price, avg_cost):  
	"""  
	计算价值平均策略的本期投资金额  
	"""  
	# 计算目标价值  
	target_value = target_monthly_growth * (months_elapsed + 1)  
     
	# 计算实际价值与目标价值的差距  
	value_gap = target_value - current_portfolio_value  
     
	# 如果实际价值超过目标，可以考虑卖出部分  
	if value_gap < 0:  
    	action = "卖出"  
    	amount = abs(value_gap)  
    	reason = f"当前价值${current_portfolio_value:.0f}超过目标${target_value:.0f}，建议卖出${amount:.0f}"  
	elif value_gap > 0:  
    	action = "买入"  
    	amount = value_gap  
    	reason = f"当前价值${current_portfolio_value:.0f}低于目标${target_value:.0f}，建议买入${amount:.0f}"  
	else:  
    	action = "持有"  
    	amount = 0  
    	reason = "当前价值符合目标，无需操作"  
     
	return {  
    	"action": action,  
    	"amount_usd": round(amount, 2),  
    	"target_value": round(target_value, 2),  
    	"current_value": round(current_portfolio_value, 2),  
    	"reason": reason  
	}

 

 

 

**S09: 反马丁格尔加仓策略 (Anti-Martingale / Pyramiding)**

**策略概述**

| 项目 | 内容 |
| :---- | :---- |
| 策略ID | S09 |
| 策略名称 | 反马丁格尔加仓 |
| 策略类型 | 解套回本 |
| 适用场景 | 已盈利，趋势向上，想加仓扩大收益 |
| 风险等级 | 中 |
| 实施阶段 | V1.5 (P1) |

**核心逻辑**

与马丁格尔相反，反马丁格尔是"越涨越买"。当资产价格上涨并盈利时，逐步加仓，顺势而为。

 

**核心参数计算**

def calculate_anti_martingale_params(current_profit_pct, available_funds, user_risk_level):  
	"""  
	计算反马丁格尔加仓参数  
	"""  
	# 加仓触发条件：每盈利X%加仓一次  
	profit_steps = {  
    	"conservative": 0.15,   # 每盈利15%加仓  
    	"moderate": 0.10,   	# 每盈利10%加仓  
    	"aggressive": 0.05  	# 每盈利5%加仓  
	}  
     
	step = profit_steps[user_risk_level]  
     
	# 计算可以加仓几次  
	max_additions = int(current_profit_pct / step)  
     
	# 每次加仓金额递增  
	base_addition = available_funds * 0.1  
     
	plan = []  
	for i in range(1, max_additions + 1):  
    	trigger_profit = step * i  
    	addition_amount = base_addition * (1.2 ** (i - 1))  # 递增20%  
         
    	plan.append({  
        	"addition_number": i,  
        	"trigger_profit_pct": f"{trigger_profit*100:.0f}%",  
        	"addition_amount_usd": round(addition_amount, 2),  
        	"reason": f"盈利达到{trigger_profit*100:.0f}%，趋势向上，加仓扩大收益"  
    	})  
     
	return plan

 

 

 

**S10: 金字塔加仓策略 (Pyramid Adding)**

**策略概述**

| 项目 | 内容 |
| :---- | :---- |
| 策略ID | S10 |
| 策略名称 | 金字塔加仓 |
| 策略类型 | 解套回本 |
| 适用场景 | 趋势明确，想分批建仓，控制风险 |
| 风险等级 | 中 |
| 实施阶段 | V1.5 (P1) |

**核心逻辑**

金字塔加仓是一种"越涨越少买"的策略。首次建仓金额最大，后续每次加仓金额递减，形成金字塔形状。

 

def calculate_pyramid_adding(total_budget, num_layers, current_price):  
	"""  
	计算金字塔加仓方案  
	"""  
	# 金字塔比例：如3层为 50% : 30% : 20%  
	ratios = [0.5, 0.3, 0.2] if num_layers == 3 else [0.4, 0.3, 0.2, 0.1]  
     
	plan = []  
	for i, ratio in enumerate(ratios[:num_layers]):  
    	layer_amount = total_budget * ratio  
    	trigger_price = current_price * (1 + 0.05 * i)  # 每涨5%加一次  
         
    	plan.append({  
        	"layer": i + 1,  
        	"percentage": f"{ratio*100:.0f}%",  
        	"amount_usd": round(layer_amount, 2),  
        	"trigger_price": round(trigger_price, 2),  
        	"reason": f"第{i+1}层建仓，占总预算{ratio*100:.0f}%"  
    	})  
     
	return plan

 

 

 

**S11-S13 简要说明**

**S11: 区间交易策略**

在明确的支撑位买入，阻力位卖出，适合震荡行情。

 

**S12: 均值回归策略**

当价格偏离移动平均线超过一定幅度时，预期回归，进行反向交易。

 

**S13: 保本止损策略**

盈利后，将止损位上移至成本价，确保不亏本金。

 

 

 

**V1.5阶段策略总结**

| 策略 | 核心价值 | 技术复杂度 | 开发工作量 |
| :---- | :---- | :---- | :---- |
| S06 现货网格 | 自动化高抛低吸 | 高 | 6天 |
| S07 定时定额卖出 | 分散退出风险 | 低 | 2天 |
| S08 价值平均 | 智能调仓 | 中 | 3天 |
| S09 反马丁格尔 | 盈利加仓 | 中 | 3天 |
| S10 金字塔加仓 | 分批建仓 | 中 | 3天 |
| S11 区间交易 | 震荡套利 | 中 | 4天 |
| S12 均值回归 | 捕捉回归 | 中 | 3天 |
| S13 保本止损 | 保护本金 | 低 | 2天 |

总计：8个策略，26天开发工作量

**Money Coach策略库详细设计 - Part 3: V2.0 & V3.0阶段（S14-S25）**

**V2.0阶段：高级套利与技术策略（S14-S20）**

 

 

**S14: 跨交易所套利策略 (Cross-Exchange Arbitrage)**

**策略概述**

| 项目 | 内容 |
| :---- | :---- |
| 策略ID | S14 |
| 策略名称 | 跨交易所套利 |
| 策略类型 | 套利策略 |
| 适用场景 | 不同交易所间存在价差，有多个交易所账户 |
| 风险等级 | 低-中 |
| 实施阶段 | V2.0 (P2) |

**核心逻辑**

监控多个交易所的同一资产价格，当价差超过手续费和滑点成本时，在低价交易所买入，高价交易所卖出，锁定无风险利润。

 

**核心参数计算**

def detect_arbitrage_opportunity(exchanges_prices, trading_fees):  
	"""  
	检测跨交易所套利机会  
	"""  
	opportunities = []  
     
	# 遍历所有交易所对  
	for buy_exchange, buy_price in exchanges_prices.items():  
    	for sell_exchange, sell_price in exchanges_prices.items():  
        	if buy_exchange != sell_exchange:  
            	# 计算价差  
            	price_diff = sell_price - buy_price  
            	price_diff_pct = price_diff / buy_price  
                 
            	# 计算总成本  
            	buy_fee = buy_price * trading_fees[buy_exchange]  
            	sell_fee = sell_price * trading_fees[sell_exchange]  
            	withdrawal_fee = 0.0005 * buy_price  # 假设提现费0.05%  
            	total_cost = buy_fee + sell_fee + withdrawal_fee  
                 
            	# 净利润  
            	net_profit = price_diff - total_cost  
            	net_profit_pct = net_profit / buy_price  
                 
            	# 如果净利润>0.3%，认为是有效套利机会  
            	if net_profit_pct > 0.003:  
                	opportunities.append({  
                    	"buy_exchange": buy_exchange,  
                    	"sell_exchange": sell_exchange,  
                    	"buy_price": round(buy_price, 2),  
                    	"sell_price": round(sell_price, 2),  
                    	"price_diff_pct": f"{price_diff_pct*100:.2f}%",  
                    	"net_profit_pct": f"{net_profit_pct*100:.2f}%",  
                    	"recommended_amount": 1000,  # 建议套利金额  
                    	"estimated_profit": round(net_profit * 1000 / buy_price, 2)  
                	})  
     
	# 按净利润排序  
	opportunities.sort(key=lambda x: float(x["net_profit_pct"].strip("%")), reverse=True)  
     
	return opportunities

 

**风险提示**

def assess_arbitrage_risks(opportunity, exchange_liquidity):  
	"""  
	评估套利风险  
	"""  
	risks = []  
     
	# 流动性风险  
	if exchange_liquidity[opportunity["buy_exchange"]] < opportunity["recommended_amount"]:  
    	risks.append("买入交易所流动性不足，可能无法成交")  
     
	# 提现时间风险  
	withdrawal_time = get_withdrawal_time(opportunity["buy_exchange"])  
	if withdrawal_time > 30:  # 超过30分钟  
    	risks.append(f"提现时间较长（{withdrawal_time}分钟），价差可能消失")  
     
	# 价格波动风险  
	if opportunity["price_diff_pct"] < "1.0%":  
    	risks.append("价差较小，市场波动可能导致套利失败")  
     
	return risks

 

 

 

**S15: 三角套利策略 (Triangular Arbitrage)**

**策略概述**

| 项目 | 内容 |
| :---- | :---- |
| 策略ID | S15 |
| 策略名称 | 三角套利 |
| 策略类型 | 套利策略 |
| 适用场景 | 同一交易所内，三种货币间存在价差 |
| 风险等级 | 低 |
| 实施阶段 | V2.0 (P2) |

**核心逻辑**

在同一交易所内，通过三种货币的循环兑换，利用汇率不一致获利。

 

示例：USDT → BTC → ETH → USDT

 

def detect_triangular_arbitrage(exchange_rates):  
	"""  
	检测三角套利机会  
     
	exchange_rates = {  
    	"BTC/USDT": 95000,  
    	"ETH/USDT": 3000,  
    	"ETH/BTC": 0.0316  
	}  
	"""  
	# 计算循环汇率  
	# 路径：USDT → BTC → ETH → USDT  
	usdt_to_btc = 1 / exchange_rates["BTC/USDT"]  
	btc_to_eth = 1 / exchange_rates["ETH/BTC"]  
	eth_to_usdt = exchange_rates["ETH/USDT"]  
     
	final_amount = 1 * usdt_to_btc * btc_to_eth * eth_to_usdt  
     
	# 减去手续费（假设每次0.1%）  
	final_amount_after_fees = final_amount * (0.999 ** 3)  
     
	profit_pct = (final_amount_after_fees - 1) * 100  
     
	if profit_pct > 0.1:  # 利润>0.1%  
    	return {  
        	"opportunity": True,  
        	"path": "USDT → BTC → ETH → USDT",  
        	"profit_pct": f"{profit_pct:.2f}%",  
        	"recommended_amount": 1000,  
        	"estimated_profit": round(profit_pct * 10, 2)  
    	}  
	else:  
    	return {"opportunity": False}

 

 

 

**S16: 资金费率套利策略 (Funding Rate Arbitrage)**

**策略概述**

| 项目 | 内容 |
| :---- | :---- |
| 策略ID | S16 |
| 策略名称 | 资金费率套利 |
| 策略类型 | 套利策略 |
| 适用场景 | 永续合约资金费率较高，持有现货对冲 |
| 风险等级 | 中 |
| 实施阶段 | V2.0 (P2) |

**核心逻辑**

当永续合约的资金费率为正且较高时，做空永续合约，同时买入等量现货，赚取资金费率收益，价格波动被对冲。

 

def calculate_funding_rate_arbitrage(funding_rate, spot_price, contract_price, holding_period_hours=24):  
	"""  
	计算资金费率套利收益  
	"""  
	# 资金费率通常每8小时结算一次  
	funding_periods = holding_period_hours / 8  
     
	# 总资金费率收益  
	total_funding_income = funding_rate * funding_periods  
     
	# 计算基差（合约价格 - 现货价格）  
	basis = contract_price - spot_price  
	basis_pct = basis / spot_price  
     
	# 如果基差为负，套利更有利（MVP：使用 max(0, -basis_pct) 计算）  
	additional_profit = max(0, -basis_pct)  
     
	# 总收益 = 资金费率收益 + 基差收益 - 手续费  
	trading_fee = 0.001 * 2  # 开仓和平仓各0.1%  
	net_profit_pct = total_funding_income + additional_profit - trading_fee  
     
	if net_profit_pct > 0.005:  # 净利润>0.5%  
    	return {  
        	"opportunity": True,  
        	"funding_rate": f"{funding_rate*100:.3f}%",  
        	"holding_period": f"{holding_period_hours}小时",  
        	"net_profit_pct": f"{net_profit_pct*100:.2f}%",  
        	"annualized_return": f"{net_profit_pct * 365 / (holding_period_hours/24) * 100:.1f}%",  
        	"action": "买入现货 + 做空永续合约"  
    	}  
	else:  
    	return {"opportunity": False}

 

 

 

**S17: 期现套利策略 (Cash and Carry Arbitrage)**

**策略概述**

| 项目 | 内容 |
| :---- | :---- |
| 策略ID | S17 |
| 策略名称 | 期现套利 |
| 策略类型 | 套利策略 |
| 适用场景 | 期货价格高于现货，持有至交割锁定价差 |
| 风险等级 | 中 |
| 实施阶段 | V2.0 (P2) |

**核心逻辑**

当期货价格显著高于现货价格时，买入现货，卖出期货，持有至交割日，锁定价差收益。

 

def calculate_cash_carry_arbitrage(spot_price, futures_price, days_to_expiry, annual_risk_free_rate=0.05):  
	"""  
	计算期现套利收益  
	"""  
	# 计算基差  
	basis = futures_price - spot_price  
	basis_pct = basis / spot_price  
     
	# 计算理论期货价格（考虑无风险利率）  
	theoretical_futures = spot_price * (1 + annual_risk_free_rate * days_to_expiry / 365)  
     
	# 如果实际期货价格 > 理论价格，存在套利机会  
	arbitrage_profit = futures_price - theoretical_futures  
	arbitrage_profit_pct = arbitrage_profit / spot_price  
     
	# 减去交易成本  
	trading_cost = spot_price * 0.002  # 0.2%手续费  
	net_profit = arbitrage_profit - trading_cost  
	net_profit_pct = net_profit / spot_price  
     
	# 年化收益率  
	annualized_return = net_profit_pct * (365 / days_to_expiry)  
     
	if net_profit_pct > 0.01:  # 净利润>1%  
    	return {  
        	"opportunity": True,  
        	"spot_price": round(spot_price, 2),  
        	"futures_price": round(futures_price, 2),  
        	"basis_pct": f"{basis_pct*100:.2f}%",  
        	"days_to_expiry": days_to_expiry,  
        	"net_profit_pct": f"{net_profit_pct*100:.2f}%",  
        	"annualized_return": f"{annualized_return*100:.1f}%",  
        	"action": "买入现货 + 卖出期货，持有至交割"  
    	}  
	else:  
    	return {"opportunity": False}

 

 

 

**S18: 趋势跟随策略 (Trend Following)**

**策略概述**

| 项目 | 内容 |
| :---- | :---- |
| 策略ID | S18 |
| 策略名称 | 趋势跟随 |
| 策略类型 | 技术分析 |
| 适用场景 | 明确的上涨或下跌趋势，顺势交易 |
| 风险等级 | 中 |
| 实施阶段 | V2.0 (P2) |

**核心逻辑**

使用移动平均线等技术指标识别趋势，在上涨趋势中持有或加仓，在下跌趋势中减仓或做空（MVP 不支持做空，仅减仓）。

 

def detect_trend(historical_data, ma_short=20, ma_medium=50, ma_long=200):  
	"""  
	检测趋势方向和强度  
	"""  
	current_price = historical_data[-1]  
	ma_20 = calculate_ma(historical_data, ma_short)  
	ma_50 = calculate_ma(historical_data, ma_medium)  
	ma_200 = calculate_ma(historical_data, ma_long)  
     
	# 判断趋势  
	if current_price > ma_20 > ma_50 > ma_200:  
    	trend = "强上涨趋势"  
    	strength = "强"  
    	action = "持有或加仓"  
	elif current_price > ma_50 > ma_200:  
    	trend = "上涨趋势"  
    	strength = "中"  
    	action = "持有"  
	elif current_price < ma_20 < ma_50 < ma_200:  
    	trend = "强下跌趋势"  
    	strength = "强"  
    	action = "减仓或做空"  
	elif current_price < ma_50 < ma_200:  
    	trend = "下跌趋势"  
    	strength = "中"  
    	action = "减仓"  
	else:  
    	trend = "震荡"  
    	strength = "弱"  
    	action = "观望"  
     
	return {  
    	"trend": trend,  
    	"strength": strength,  
    	"action": action,  
    	"current_price": round(current_price, 2),  
    	"ma_20": round(ma_20, 2),  
    	"ma_50": round(ma_50, 2),  
    	"ma_200": round(ma_200, 2)  
	}

 

 

 

**S19: 突破交易策略 (Breakout Trading)**

**策略概述**

| 项目 | 内容 |
| :---- | :---- |
| 策略ID | S19 |
| 策略名称 | 突破交易 |
| 策略类型 | 技术分析 |
| 适用场景 | 价格突破关键阻力位或支撑位，捕捉新趋势 |
| 风险等级 | 中 |
| 实施阶段 | V2.0 (P2) |

**核心逻辑**

识别关键的支撑位和阻力位，当价格突破时入场，设置止损防止假突破。

 

def detect_breakout(current_price, resistance_level, support_level, volume_ratio):  
	"""  
	检测突破信号  
	"""  
	# 向上突破  
	if current_price > resistance_level * 1.02 and volume_ratio > 1.5:  
    	return {  
        	"breakout": True,  
        	"direction": "向上",  
        	"action": "买入",  
        	"entry_price": current_price,  
        	"stop_loss": resistance_level * 0.98,  
        	"target_price": current_price * 1.10,  
        	"reason": f"突破阻力位${resistance_level:.2f}，成交量放大{volume_ratio:.1f}倍"  
    	}  
     
	# 向下突破  
	elif current_price < support_level * 0.98 and volume_ratio > 1.5:  
    	return {  
        	"breakout": True,  
        	"direction": "向下",  
        	"action": "卖出或做空",  
        	"entry_price": current_price,  
        	"stop_loss": support_level * 1.02,  
        	"target_price": current_price * 0.90,  
        	"reason": f"跌破支撑位${support_level:.2f}，成交量放大{volume_ratio:.1f}倍"  
    	}  
     
	else:  
    	return {"breakout": False}

 

 

 

**S20: 合约对冲策略 (Hedging with Futures)**

**策略概述**

| 项目 | 内容 |
| :---- | :---- |
| 策略ID | S20 |
| 策略名称 | 合约对冲 |
| 策略类型 | 风险管理 |
| 适用场景 | 持有现货，担心短期下跌，用合约对冲 |
| 风险等级 | 中 |
| 实施阶段 | V2.0 (P2) |

**核心逻辑**

持有现货的同时，开空单合约，对冲价格下跌风险。

 

def calculate_hedge_ratio(spot_holdings, spot_price, user_hedge_percentage):  
	"""  
	计算对冲比例和合约数量  
	"""  
	# 计算需要对冲的价值  
	total_spot_value = spot_holdings * spot_price  
	hedge_value = total_spot_value * user_hedge_percentage  
     
	# 计算合约数量（假设1张合约=0.01 BTC）  
	contract_size = 0.01  
	contracts_needed = hedge_value / (spot_price * contract_size)  
     
	return {  
    	"spot_holdings": spot_holdings,  
    	"spot_value": round(total_spot_value, 2),  
    	"hedge_percentage": f"{user_hedge_percentage*100:.0f}%",  
    	"hedge_value": round(hedge_value, 2),  
    	"contracts_needed": int(contracts_needed),  
    	"action": f"开空{int(contracts_needed)}张合约",  
    	"protection": f"对冲{user_hedge_percentage*100:.0f}%的下跌风险"  
	}

 

 

 

**V3.0阶段：组合管理策略（S21-S25）**

 

 

**S21: 动态再平衡策略 (Dynamic Rebalancing)**

**策略概述**

| 项目 | 内容 |
| :---- | :---- |
| 策略ID | S21 |
| 策略名称 | 动态再平衡 |
| 策略类型 | 组合管理 |
| 适用场景 | 多资产组合，保持目标配置比例 |
| 风险等级 | 低 |
| 实施阶段 | V3.0 (P3) |

**核心逻辑**

定期（如每月）或当资产配置偏离目标超过一定阈值时，卖出涨幅大的资产，买入跌幅大的资产，恢复目标配置。

 

def calculate_rebalancing_actions(portfolio, target_allocation, rebalance_threshold=0.05):  
	"""  
	计算再平衡操作  
     
	portfolio = {  
    	"BTC": {"value": 50000, "target": 0.40},  
    	"ETH": {"value": 35000, "target": 0.30},  
    	"SOL": {"value": 15000, "target": 0.30}  
	}  
	"""  
	total_value = sum([asset["value"] for asset in portfolio.values()])  
	actions = []  
     
	for symbol, data in portfolio.items():  
    	current_allocation = data["value"] / total_value  
    	target = data["target"]  
    	deviation = current_allocation - target  
         
    	# 如果偏离超过阈值  
    	if abs(deviation) > rebalance_threshold:  
        	target_value = total_value * target  
        	adjustment_value = data["value"] - target_value  
             
        	if adjustment_value > 0:  
            	action = "卖出"  
        	else:  
            	action = "买入"  
             
        	actions.append({  
            	"asset": symbol,  
            	"current_allocation": f"{current_allocation*100:.1f}%",  
            	"target_allocation": f"{target*100:.1f}%",  
            	"deviation": f"{deviation*100:+.1f}%",  
            	"action": action,  
            	"amount_usd": round(abs(adjustment_value), 2)  
        	})  
     
	return actions

 

 

 

**S22: 风险平价策略 (Risk Parity)**

**策略概述**

| 项目 | 内容 |
| :---- | :---- |
| 策略ID | S22 |
| 策略名称 | 风险平价 |
| 策略类型 | 组合管理 |
| 适用场景 | 多资产组合，希望每个资产贡献相同的风险 |
| 风险等级 | 低 |
| 实施阶段 | V3.0 (P3) |

**核心逻辑**

根据每个资产的波动率，调整配置比例，使每个资产对组合总风险的贡献相等。

 

def calculate_risk_parity_allocation(assets_volatility):  
	"""  
	计算风险平价配置  
     
	assets_volatility = {  
    	"BTC": 0.04,  
    	"ETH": 0.06,  
    	"USDT": 0.001  
	}  
	"""  
	# 计算每个资产的风险权重（波动率的倒数）  
	risk_weights = {asset: 1/vol for asset, vol in assets_volatility.items()}  
	total_risk_weight = sum(risk_weights.values())  
     
	# 归一化为配置比例  
	allocations = {asset: weight/total_risk_weight for asset, weight in risk_weights.items()}  
     
	return {  
    	asset: {  
        	"allocation": f"{alloc*100:.1f}%",  
        	"volatility": f"{assets_volatility[asset]*100:.1f}%",  
        	"risk_contribution": "均等"  
    	}  
    	for asset, alloc in allocations.items()  
	}

 

 

 

**S23: 税收优化策略 (Tax Loss Harvesting)**

**策略概述**

| 项目 | 内容 |
| :---- | :---- |
| 策略ID | S23 |
| 策略名称 | 税收优化策略 |
| 策略类型 | 组合管理 |
| 适用场景 | 年底税务规划，通过实现亏损抵扣税收 |
| 风险等级 | 低 |
| 实施阶段 | V3.0 (P3) |

**核心逻辑**

在年底前，卖出亏损的资产实现亏损，用于抵扣其他资产的盈利，降低税负。

 

def identify_tax_loss_harvesting_opportunities(portfolio, tax_rate=0.20):  
	"""  
	识别税收优化机会  
	"""  
	opportunities = []  
     
	for asset, data in portfolio.items():  
    	if data["unrealized_pnl"] < 0:  
        	# 计算实现亏损后的税收节省  
        	loss_amount = abs(data["unrealized_pnl"])  
        	tax_savings = loss_amount * tax_rate  
             
        	opportunities.append({  
            	"asset": asset,  
            	"unrealized_loss": round(loss_amount, 2),  
            	"tax_savings": round(tax_savings, 2),  
            	"action": "卖出实现亏损",  
            	"note": "可在30天后重新买入（避免洗售规则）"  
        	})  
     
	return opportunities

 

 

 

**S24: 波动率套利策略 (Volatility Arbitrage)**

**策略概述**

| 项目 | 内容 |
| :---- | :---- |
| 策略ID | S24 |
| 策略名称 | 波动率套利 |
| 策略类型 | 解套回本 |
| 适用场景 | 高波动环境，使用期权策略 |
| 风险等级 | 高 |
| 实施阶段 | V3.0 (P3) |

注：此策略需要期权市场支持，实现复杂度高。

 

 

 

**S25: 事件驱动交易策略 (Event-Driven Trading)**

**策略概述**

| 项目 | 内容 |
| :---- | :---- |
| 策略ID | S25 |
| 策略名称 | 事件驱动交易 |
| 策略类型 | 技术分析 |
| 适用场景 | 重大事件前后（如减半、升级、监管），捕捉波动 |
| 风险等级 | 高 |
| 实施阶段 | V3.0 (P3) |

**核心逻辑**

监控重大事件日历，在事件前后根据历史规律和市场情绪进行交易。

 

def analyze_event_impact(event_type, days_to_event, historical_impact):  
	"""  
	分析事件对价格的潜在影响  
     
	event_type: "halving", "upgrade", "regulation"  
	historical_impact: {"avg_change_before": 0.15, "avg_change_after": -0.05}  
	"""  
	if days_to_event > 30:  
    	recommendation = "观望，事件影响尚未体现"  
	elif days_to_event > 7:  
    	if historical_impact["avg_change_before"] > 0.10:  
        	recommendation = "考虑买入，历史上事件前平均上涨"  
    	else:  
        	recommendation = "观望"  
	elif days_to_event > 0:  
    	recommendation = "谨慎，事件临近，波动加大"  
	else:  
    	if historical_impact["avg_change_after"] < -0.05:  
        	recommendation = "考虑卖出，历史上事件后平均下跌"  
    	else:  
        	recommendation = "观望"  
     
	return {  
    	"event": event_type,  
    	"days_to_event": days_to_event,  
    	"historical_avg_before": f"{historical_impact['avg_change_before']*100:+.1f}%",  
    	"historical_avg_after": f"{historical_impact['avg_change_after']*100:+.1f}%",  
    	"recommendation": recommendation  
	}

 

 

 

**V2.0 & V3.0阶段策略总结**

| 策略 | 核心价值 | 技术复杂度 | 开发工作量 |
| :---- | :---- | :---- | :---- |
| V2.0阶段 |   |   |   |
| S14 跨交易所套利 | 价差套利 | 高 | 7天 |
| S15 三角套利 | 汇率套利 | 中 | 4天 |
| S16 资金费率套利 | 费率收益 | 中 | 5天 |
| S17 期现套利 | 基差收益 | 中 | 5天 |
| S18 趋势跟随 | 顺势交易 | 中 | 4天 |
| S19 突破交易 | 捕捉新趋势 | 中 | 4天 |
| S20 合约对冲 | 风险对冲 | 中 | 5天 |
| V3.0阶段 |   |   |   |
| S21 动态再平衡 | 组合优化 | 中 | 5天 |
| S22 风险平价 | 风险均衡 | 高 | 6天 |
| S23 税收优化 | 税务规划 | 低 | 3天 |
| S24 波动率套利 | 期权策略 | 极高 | 10天 |
| S25 事件驱动 | 事件交易 | 高 | 6天 |

V2.0总计：7个策略，34天开发工作量  
V3.0总计：5个策略，30天开发工作量

 

 

 

**全策略库开发工作量汇总**

| 阶段 | 策略数量 | 开发工作量 | 累计策略数 |
| :---- | :---- | :---- | :---- |
| MVP | 5个 | 19天 | 5 |
| V1.5 | +8个 | 26天 | 13 |
| V2.0 | +7个 | 34天 | 20 |
| V3.0 | +5个 | 30天 | 25 |
| 总计 | 25个 | 109天 | 25 |

注：以上为纯开发工作量，不包括测试、优化和文档编写时间。
