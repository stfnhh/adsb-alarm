# ✈️ ADS-B Inbound Aircraft Monitor

A tiny Go service (~1.7 MB container) that polls the ADSB.fi Open Data API and triggers a webhook when configured aircraft categories are inbound toward your location.

Built for home automation — e.g., play a sound when a helicopter approaches.

## ✅ Features

* Polls ADSB.fi REST API on a schedule
* Filters aircraft by category (configurable)
* Detects inbound vs outbound using track + dir
* Optional distance-trend check
* Temporary blacklist prevents repeat alerts
* Minimal CPU/memory usage
* Runs in Docker, K3s, Kubernetes, or bare metal
* Uses UniFi Protect webhook (or any HTTP endpoint)

## 🚨 How Alerting Works

An alert fires when:

* Aircraft category matches ALERT_CATEGORIES
* Aircraft is not currently blacklisted
* Aircraft heading (track) is within 90° of dir → inbound
* Optional: distance decreasing → approaching

Webhook triggered once  
Aircraft hex added to blacklist for BLACKLIST_TTL

## License

MIT — do whatever you want 