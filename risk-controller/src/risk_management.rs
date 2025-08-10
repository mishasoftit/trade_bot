use serde::Deserialize;

#[derive(Debug, Deserialize, Clone)]
pub struct RiskParameters {
    pub position_sizing_pct: f64,      // Max 0.5% per trade
    pub daily_loss_limit_pct: f64,     // 2% = $40
    pub monthly_drawdown_pct: f64,     // 15% = $300
    pub account_balance: f64,          // Current account balance
}

impl RiskParameters {
    pub fn new(account_balance: f64) -> Self {
        RiskParameters {
            position_sizing_pct: 0.005,   // 0.5%
            daily_loss_limit_pct: 0.02,    // 2%
            monthly_drawdown_pct: 0.15,    // 15%
            account_balance,
        }
    }

    /// Calculate maximum position size for a trade
    pub fn max_position_size(&self) -> f64 {
        self.account_balance * self.position_sizing_pct
    }

    /// Calculate daily loss limit in absolute terms
    pub fn daily_loss_limit(&self) -> f64 {
        self.account_balance * self.daily_loss_limit_pct
    }

    /// Calculate monthly drawdown limit in absolute terms
    pub fn monthly_drawdown_limit(&self) -> f64 {
        self.account_balance * self.monthly_drawdown_pct
    }

    /// Check if trade exceeds position sizing limits
    pub fn validate_position_size(&self, position_size: f64) -> Result<(), String> {
        if position_size <= self.max_position_size() {
            Ok(())
        } else {
            Err(format!(
                "Position size ${:.2} exceeds max ${:.2} (0.5% of account)",
                position_size,
                self.max_position_size()
            ))
        }
    }

    /// Check if daily loss exceeds limit
    pub fn check_daily_loss(&self, daily_loss: f64) -> Result<(), String> {
        if daily_loss <= self.daily_loss_limit() {
            Ok(())
        } else {
            Err(format!(
                "Daily loss ${:.2} exceeds limit ${:.2} (2% of account)",
                daily_loss,
                self.daily_loss_limit()
            ))
        }
    }

    /// Check if monthly drawdown exceeds limit
    pub fn check_monthly_drawdown(&self, monthly_drawdown: f64) -> Result<(), String> {
        if monthly_drawdown <= self.monthly_drawdown_limit() {
            Ok(())
        } else {
            Err(format!(
                "Monthly drawdown ${:.2} exceeds limit ${:.2} (15% of account)",
                monthly_drawdown,
                self.monthly_drawdown_limit()
            ))
        }
    }
}