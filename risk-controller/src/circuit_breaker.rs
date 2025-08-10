use crate::risk_management::RiskParameters;
use crate::redis_state::RedisState;
use anyhow::Result;

pub struct CircuitBreaker {
    risk_params: RiskParameters,
    redis_state: RedisState,
}

impl CircuitBreaker {
    pub fn new(risk_params: RiskParameters, redis_state: RedisState) -> Self {
        CircuitBreaker {
            risk_params,
            redis_state,
        }
    }

    /// Check all risk limits and activate circuit breaker if any are breached
    pub fn check_limits(&mut self, daily_loss: f64, monthly_drawdown: f64) -> Result<bool> {
        // Check daily loss limit
        if let Err(e) = self.risk_params.check_daily_loss(daily_loss) {
            self.redis_state.set_circuit_breaker(true)?;
            log::error!("Daily loss limit breached: {}", e);
            return Ok(true);
        }

        // Check monthly drawdown limit
        if let Err(e) = self.risk_params.check_monthly_drawdown(monthly_drawdown) {
            self.redis_state.set_circuit_breaker(true)?;
            log::error!("Monthly drawdown limit breached: {}", e);
            return Ok(true);
        }

        // If all checks pass, ensure circuit breaker is off
        self.redis_state.set_circuit_breaker(false)?;
        Ok(false)
    }

    /// Get current circuit breaker state
    pub fn is_activated(&self) -> Result<bool> {
        self.redis_state.get_circuit_breaker()
    }

    /// Manual override to activate circuit breaker
    pub fn activate(&self) -> Result<()> {
        self.redis_state.set_circuit_breaker(true)
    }

    /// Manual override to deactivate circuit breaker
    pub fn deactivate(&self) -> Result<()> {
        self.redis_state.set_circuit_breaker(false)
    }
}