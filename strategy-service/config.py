import os

EXCHANGES = {
    'binance': {
        'apiKey': os.getenv('BINANCE_API_KEY'),
        'secret': os.getenv('BINANCE_API_SECRET'),
        'enableRateLimit': True
    },
    'kraken': {
        'apiKey': os.getenv('KRAKEN_API_KEY'),
        'secret': os.getenv('KRAKEN_API_SECRET'),
        'enableRateLimit': True
    },
    'coinbase': {
        'apiKey': os.getenv('COINBASE_API_KEY'),
        'secret': os.getenv('COINBASE_API_SECRET'),
        'enableRateLimit': True
    }
}

RISK_PARAMS = {
    'max_trade_percentage': 0.005,  # 0.5%
    'max_trade_value': 10.0  # $10
}

DATA_STORAGE_PATH = './data/historical/'