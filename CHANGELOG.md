# 更新日志

团队成员在此记录各自的改动，便于双方同步进度。

---

## 2026-04-05 15:00:00

**提交人：Ennio**

- 新增「每日资产简报 Daily Briefing」完整功能
- 后端：
  - 新增 Gemini prompt（system-prompts/daily-briefing.txt），支持多语言和经验等级适配
  - 新增 Briefing 数据模型和自动迁移
  - 新增每日调度器（briefing_scheduler.go），每天早8点按用户时区触发，Redis 分布式锁+去重
  - 新增任务处理（briefing_processing.go），调用 Gemini 生成5条简报后存库
  - 新增推送（briefing_push.go），选 priority 最高2条通过 APNS/FCM 推送
  - 新增 API：GET /v1/briefings/today
  - defaultNotificationPrefs 新增 daily_briefing 默认开启
- 前端：
  - Assets Tab 顶部新增每日简报卡片，支持下拉刷新
  - 新增 Push 通知引导页（sc08），登录后首次未授权时展示
  - 新增 briefings API service
  - 5种语言（en/zh-CN/zh-TW/ja/ko）完整翻译

---

## 2026-04-03 16:00:00

**提交人：Ennio**

- 创建 CHANGELOG.md，用于团队协作记录改动
