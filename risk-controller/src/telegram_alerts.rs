use teloxide::prelude::*;
use std::env;
use anyhow::Result;

pub struct TelegramAlert {
    bot: Bot,
    chat_id: ChatId,
}

impl TelegramAlert {
    pub fn new() -> Result<Self> {
        let bot_token = env::var("TELEGRAM_BOT_TOKEN")
            .expect("TELEGRAM_BOT_TOKEN environment variable not set");
        
        let chat_id_str = env::var("TELEGRAM_CHAT_ID")
            .expect("TELEGRAM_CHAT_ID environment variable not set");
        
        let chat_id = chat_id_str.parse::<i64>()?;
        
        Ok(TelegramAlert {
            bot: Bot::new(bot_token),
            chat_id: ChatId(chat_id),
        })
    }

    pub async fn send_alert(&self, message: &str) -> Result<()> {
        self.bot.send_message(self.chat_id, message).await?;
        Ok(())
    }

    pub async fn risk_limit_breached(&self, limit_type: &str, current: f64, limit: f64) -> Result<()> {
        let message = format!(
            "🚨 RISK LIMIT BREACHED 🚨\n\n{}: ${:.2}\nLimit: ${:.2}\n\nAll trading halted.",
            limit_type, current, limit
        );
        self.send_alert(&message).await
    }

    pub async fn circuit_breaker_activated(&self) -> Result<()> {
        let message = "🔴 CIRCUIT BREAKER ACTIVATED 🔴\n\nTrading halted due to risk limit breach.";
        self.send_alert(message).await
    }

    pub async fn circuit_breaker_deactivated(&self) -> Result<()> {
        let message = "🟢 CIRCUIT BREAKER DEACTIVATED 🟢\n\nTrading resumed.";
        self.send_alert(message).await
    }
}