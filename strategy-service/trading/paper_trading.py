import asyncio
from ..config import EXCHANGES, RISK_PARAMS
from ..arbitrage.detector import ArbitrageDetector

class PaperTradingSimulator:
    def __init__(self, initial_balance=10000):
        self.balance = initial_balance
        self.positions = {}
        self.detector = ArbitrageDetector(EXCHANGES)
        
    async def execute_trade(self, opportunity):
        symbol = opportunity['symbol']
        bid_exchange = opportunity['bid_exchange']
        ask_exchange = opportunity['ask_exchange']
        bid_price = opportunity['bid_price']
        ask_price = opportunity['ask_price']
        
        # Calculate maximum trade size based on risk parameters
        max_trade_value = min(RISK_PARAMS['max_trade_value'], 
                             self.balance * RISK_PARAMS['max_trade_percentage'])
        trade_size = max_trade_value / ask_price
        
        # Simulate buying on ask exchange and selling on bid exchange
        self.balance -= ask_price * trade_size
        self.balance += bid_price * trade_size
        
        # Log the trade
        profit = (bid_price - ask_price) * trade_size
        print(f"Executed arbitrage trade: Bought {trade_size} {symbol} on {ask_exchange} at {ask_price}, Sold on {bid_exchange} at {bid_price}")
        print(f"Profit: ${profit:.2f}, New Balance: ${self.balance:.2f}")
        
        return profit
        
    async def run(self):
        try:
            while True:
                opportunity = await self.detector.detect_opportunities()
                if opportunity:
                    await self.execute_trade(opportunity)
                await asyncio.sleep(10)  # Check every 10 seconds
        except KeyboardInterrupt:
            pass
        finally:
            await self.detector.close()
            
if __name__ == "__main__":
    simulator = PaperTradingSimulator()
    asyncio.run(simulator.run())