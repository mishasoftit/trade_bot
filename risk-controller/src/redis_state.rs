use redis::{Client, Commands, Connection, RedisResult};
use redis::r2d2::{Pool, PooledConnection, ConnectionManager};
use std::env;
use anyhow::Result;

pub struct RedisState {
    pool: Pool<ConnectionManager>,
}

impl RedisState {
    pub fn new() -> Result<Self> {
        let redis_url = env::var("REDIS_URL")
            .unwrap_or_else(|_| "redis://127.0.0.1:6379".to_string());
        
        let client = Client::open(redis_url)?;
        let manager = ConnectionManager::new(client);
        let pool = Pool::builder().build(manager)?;

        Ok(RedisState { pool })
    }

    fn get_conn(&self) -> Result<PooledConnection<ConnectionManager>> {
        self.pool.get().map_err(|e| e.into())
    }

    pub fn set_daily_loss(&self, loss: f64) -> Result<()> {
        let mut conn = self.get_conn()?;
        conn.set("risk:daily_loss", loss.to_string())?;
        Ok(())
    }

    pub fn get_daily_loss(&self) -> Result<f64> {
        let mut conn = self.get_conn()?;
        let loss_str: String = conn.get("risk:daily_loss")?;
        loss_str.parse::<f64>().map_err(|e| e.into())
    }

    pub fn set_monthly_drawdown(&self, drawdown: f64) -> Result<()> {
        let mut conn = self.get_conn()?;
        conn.set("risk:monthly_drawdown", drawdown.to_string())?;
        Ok(())
    }

    pub fn get_monthly_drawdown(&self) -> Result<f64> {
        let mut conn = self.get_conn()?;
        let drawdown_str: String = conn.get("risk:monthly_drawdown")?;
        drawdown_str.parse::<f64>().map_err(|e| e.into())
    }

    pub fn set_circuit_breaker(&self, state: bool) -> Result<()> {
        let mut conn = self.get_conn()?;
        conn.set("risk:circuit_breaker", state.to_string())?;
        Ok(())
    }

    pub fn get_circuit_breaker(&self) -> Result<bool> {
        let mut conn = self.get_conn()?;
        let state_str: String = conn.get("risk:circuit_breaker")?;
        match state_str.as_str() {
            "true" => Ok(true),
            "false" => Ok(false),
            _ => Err(anyhow::anyhow!("Invalid circuit breaker state")),
        }
    }
}