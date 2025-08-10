# Trading Platform Monitoring System

## Overview
This monitoring stack provides real-time visibility into trading operations, including:
- Real-time P&L tracking
- System health monitoring
- Risk exposure analysis
- Latency heatmaps
- Alerting via Telegram

## Components
- **Prometheus**: Metrics collection and alerting
- **Grafana**: Dashboards and visualization
- **Alertmanager**: Alert routing and notification
- **Loki**: Log aggregation
- **Promtail**: Log collection
- **Custom Exporters**: Trading-specific metrics
- **Synthetic Tests**: Circuit breaker monitoring

## Setup Instructions

### Prerequisites
- Docker and Docker Compose
- Telegram bot token and chat ID

### Configuration
1. Set Telegram credentials in `.env`:
```bash
TELEGRAM_BOT_TOKEN=your_bot_token
TELEGRAM_CHAT_ID=your_chat_id
```

2. Build and start the stack:
```bash
docker-compose up -d --build
```

### Services
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000
- Alertmanager: http://localhost:9093
- Loki: http://localhost:3100

## Dashboards
Pre-configured dashboards:
1. **Real-time P&L Tracking**: Shows cumulative and hourly P&L by strategy
2. **Latency Heatmaps**: Visualizes API and processing latencies
3. **Risk Exposure Analysis**: Displays risk metrics and circuit breaker status
4. **System Health**: Shows service uptime and resource usage

## Alert Management
Alerts are configured in `prometheus/alerts.yml` and include:
- Circuit breaker triggers
- High risk exposure
- Low order fill rates
- API latency warnings
- P&L anomalies

To reload Prometheus configuration:
```bash
curl -X POST http://localhost:9090/-/reload
```

## Synthetic Monitoring
The circuit breaker test system runs continuously:
```bash
docker-compose -f synthetic-tests.yml up
```

## Maintenance
### Persistent Storage
Volumes are configured for:
- Prometheus data
- Grafana data
- Loki logs

### Upgrading
To upgrade components:
1. Update image versions in `docker-compose.yml`
2. Rebuild containers:
```bash
docker-compose up -d --build
```

## Troubleshooting
### Common Issues
1. **Alerts not firing**:
   - Check Alertmanager logs: `docker logs alertmanager`
   - Verify Telegram credentials
   - Check Prometheus rule evaluation: http://localhost:9090/rules

2. **Metrics missing**:
   - Check service discovery: http://localhost:9090/targets
   - Verify exporters are running

3. **Logs not appearing**:
   - Check Promtail configuration
   - Verify log paths match service configurations

### Log Investigation
Use Grafana's Explore tab with Loki to query logs:
```logql
{job="trading"} |= "error"
```

## Support
For additional assistance, contact:
- Trading Operations Team: operations@trading-platform.com
- On-call Engineer: +1-555-123-4567