# Strategy Service MVP

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

This project provides tools for detecting arbitrage opportunities, simulating trades, and backtesting strategies across multiple cryptocurrency exchanges.

## Features

- **Arbitrage Detection**: Find price discrepancies across exchanges
- **Paper Trading**: Simulate trades without risking real funds
- **Backtesting**: Test strategies using historical data
- **Multi-exchange Support**: Binance, Kraken, Coinbase

## Installation

1. Clone the repository:
```bash
git clone https://github.com/mishasoftit/trade_bot
cd strategy-service
```

2. Install dependencies:
```bash
pip install -r requirements.txt
```

3. Set up environment variables:
```bash
cp .env.example .env
```
Edit `.env` with your exchange API keys

## Usage

### Detect arbitrage opportunities
```bash
python main.py --mode detect --symbol BTC/USDT
```

### Run paper trading simulation
```bash
python main.py --mode trade
```

### Run backtesting
```bash
python main.py --mode backtest --config ./freqtrade_config.json
```

## Configuration
Edit `config.py` to adjust risk parameters:
```python
RISK_PARAMS = {
    'max_trade_percentage': 0.005,  # 0.5% of portfolio
    'max_trade_value': 10.0  # $10 max per trade
}
```

## Contributing
Contributions are welcome! Please open an issue or submit a pull request.

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
