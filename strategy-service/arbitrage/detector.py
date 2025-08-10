import asyncio
import ccxt.async_support as ccxt
import csv
import os
from datetime import datetime
from ..config import RISK_PARAMS, DATA_STORAGE_PATH

class ArbitrageDetector:
    def __init__(self, exchanges):
        self.exchanges = {exch: getattr(ccxt, exch)(config) for exch, config in exchanges.items()}
        self.symbols = self._get_common_symbols()
        self.data_path = DATA_STORAGE_PATH
        
    def _get_common_symbols(self):
        # Get common symbols across all exchanges
        symbols = set()
        for exchange in self.exchanges.values():
            markets = exchange.load_markets()
            symbols.update(markets.keys())
        return list(symbols)
    
    async def fetch_order_books(self, symbol):
        tasks = []
        for exchange in self.exchanges.values():
            tasks.append(exchange.fetch_order_book(symbol))
        return await asyncio.gather(*tasks, return_exceptions=True)
    
    def calculate_arbitrage(self, order_books):
        best_bid = 0
        best_ask = float('inf')
        bid_exchange = None
        ask_exchange = None
        
        for i, ob in enumerate(order_books):
            if isinstance(ob, Exception) or not ob['bids'] or not ob['asks']:
                continue
                
            exchange_name = list(self.exchanges.keys())[i]
            bid = ob['bids'][0][0]
            ask = ob['asks'][0][0]
            
            if bid > best_bid:
                best_bid = bid
                bid_exchange = exchange_name
                
            if ask < best_ask:
                best_ask = ask
                ask_exchange = exchange_name
                
        spread = best_bid - best_ask
        return {
            'symbol': symbol,
            'bid_exchange': bid_exchange,
            'bid_price': best_bid,
            'ask_exchange': ask_exchange,
            'ask_price': best_ask,
            'spread': spread
        }
    
    def check_risk(self, opportunity):
        # Implement risk management rules
        max_trade_value = RISK_PARAMS['max_trade_value']
        max_trade_percentage = RISK_PARAMS['max_trade_percentage']
        
        # Placeholder - actual implementation would use account balance
        account_balance = 2000  # Example balance
        max_trade = min(max_trade_value, account_balance * max_trade_percentage)
        
        return opportunity['spread'] > 0 and max_trade >= 10
    
    def log_opportunity(self, opportunity):
        os.makedirs(self.data_path, exist_ok=True)
        filename = os.path.join(self.data_path, f"{datetime.now().strftime('%Y%m%d')}_arbitrage.csv")
        file_exists = os.path.isfile(filename)
        
        with open(filename, 'a', newline='') as f:
            writer = csv.DictWriter(f, fieldnames=opportunity.keys())
            if not file_exists:
                writer.writeheader()
            writer.writerow(opportunity)
    
    async def detect_opportunities(self):
        for symbol in self.symbols:
            order_books = await self.fetch_order_books(symbol)
            opportunity = self.calculate_arbitrage(order_books)
            
            if opportunity and self.check_risk(opportunity):
                self.log_opportunity(opportunity)
                return opportunity
        return None
    
    async def close(self):
        for exchange in self.exchanges.values():
            await exchange.close()