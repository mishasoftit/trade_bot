from freqtrade.strategy import IStrategy
from freqtrade.resolvers import StrategyResolver
from ..arbitrage.detector import ArbitrageDetector
import pandas as pd
import numpy as np

class ArbitrageStrategy(IStrategy):
    timeframe = '1m'
    minimal_roi = {"0": 0.01}
    stoploss = -0.01
    trailing_stop = False
    process_only_new_candles = True
    use_sell_signal = True
    sell_profit_only = False
    ignore_roi_if_buy_signal = False
    
    def __init__(self, config: dict) -> None:
        super().__init__(config)
        self.detector = ArbitrageDetector(config.get('exchanges', {}))
        self.symbol = config.get('symbol', 'BTC/USDT')
        
    def populate_indicators(self, dataframe: pd.DataFrame, metadata: dict) -> pd.DataFrame:
        # Not needed for arbitrage detection
        return dataframe
        
    def populate_buy_trend(self, dataframe: pd.DataFrame, metadata: dict) -> pd.DataFrame:
        # Detect arbitrage opportunities
        opportunity = asyncio.run(self.detector.detect_opportunities())
        
        # Create buy signal when opportunity exists
        dataframe['buy'] = 0
        if opportunity and opportunity['symbol'] == metadata['pair']:
            dataframe.loc[dataframe.index[-1], 'buy'] = 1
            
        return dataframe
        
    def populate_sell_trend(self, dataframe: pd.DataFrame, metadata: dict) -> pd.DataFrame:
        # Sell immediately after buy (arbitrage is instantaneous)
        dataframe['sell'] = dataframe['buy'].shift(1).fillna(0)
        return dataframe

def run_backtest(config_path: str):
    # Load Freqtrade configuration
    from freqtrade.configuration import Configuration
    from freqtrade.data.history import load_pair_history
    from freqtrade.optimize.backtesting import Backtesting
    
    config = Configuration.from_files([config_path]).get_config()
    
    # Initialize strategy
    strategy = StrategyResolver.load_strategy(config)
    
    # Load historical data
    data = load_pair_history(
        datadir=config['datadir'],
        timeframe=config['timeframe'],
        pair=config['pairs'][0],
        timerange=config.get('timerange', None)
    )
    
    # Run backtesting
    backtesting = Backtesting(config)
    results = backtesting.backtest(
        data={config['pairs'][0]: data},
        strategy=strategy
    )
    
    return results