import asyncio
import argparse
from arbitrage.detector import ArbitrageDetector
from trading.paper_trading import PaperTradingSimulator
from backtesting.freqtrade_integration import run_backtest
from config import EXCHANGES

def main():
    parser = argparse.ArgumentParser(description='Strategy Service MVP')
    parser.add_argument('--mode', choices=['detect', 'trade', 'backtest'], 
                        default='detect', help='Operation mode')
    parser.add_argument('--symbol', default='BTC/USDT', help='Trading symbol')
    parser.add_argument('--config', default='./freqtrade_config.json', 
                        help='Freqtrade config file for backtesting')
    args = parser.parse_args()

    if args.mode == 'detect':
        detector = ArbitrageDetector(EXCHANGES)
        asyncio.run(detector.detect_opportunities())
        asyncio.run(detector.close())
    elif args.mode == 'trade':
        simulator = PaperTradingSimulator()
        asyncio.run(simulator.run())
    elif args.mode == 'backtest':
        results = run_backtest(args.config)
        print(f"Backtest results: {results}")

if __name__ == "__main__":
    main()