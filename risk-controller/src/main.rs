use actix_web::{web, App, HttpResponse, HttpServer, Responder};
use serde::Deserialize;
use anyhow::Result;
use std::sync::{Arc, Mutex};

mod risk_management;
mod redis_state;
mod circuit_breaker;
mod telegram_alerts;

use risk_management::RiskParameters;
use redis_state::RedisState;
use circuit_breaker::CircuitBreaker;
use telegram_alerts::TelegramAlert;

struct AppState {
    risk_params: Arc<Mutex<RiskParameters>>,
    circuit_breaker: Arc<Mutex<CircuitBreaker>>,
    telegram_alert: Arc<Mutex<TelegramAlert>>,
}

async fn health_check() -> impl Responder {
    HttpResponse::Ok().body("Risk Controller Operational")
}

#[derive(Deserialize)]
struct TradeRequest {
    position_size: f64,
    daily_loss: f64,
    monthly_drawdown: f64,
}

#[derive(Deserialize)]
struct RiskParamsUpdate {
    account_balance: Option<f64>,
    position_sizing_pct: Option<f64>,
    daily_loss_limit_pct: Option<f64>,
    monthly_drawdown_pct: Option<f64>,
}

async fn validate_trade(
    data: web::Data<AppState>,
    trade: web::Json<TradeRequest>,
) -> impl Responder {
    let risk_params = data.risk_params.lock().unwrap().clone();
    let mut circuit_breaker = data.circuit_breaker.lock().unwrap();
    
    // Check circuit breaker first
    if circuit_breaker.is_activated().unwrap_or(true) {
        return HttpResponse::Forbidden().body("Circuit breaker activated - trading halted");
    }
    
    // Check position size
    if let Err(e) = risk_params.validate_position_size(trade.position_size) {
        return HttpResponse::BadRequest().body(e);
    }
    
    // Check risk limits
    if let Ok(breached) = circuit_breaker.check_limits(trade.daily_loss, trade.monthly_drawdown) {
        if breached {
            // Send Telegram alert
            let telegram_alert = data.telegram_alert.lock().unwrap();
            if let Err(e) = telegram_alert.circuit_breaker_activated().await {
                log::error!("Failed to send Telegram alert: {}", e);
            }
            return HttpResponse::Forbidden().body("Risk limit breached - circuit breaker activated");
        }
    }
    
    HttpResponse::Ok().body("Trade validated")
}

async fn update_risk_params(
    data: web::Data<AppState>,
    update: web::Json<RiskParamsUpdate>,
) -> impl Responder {
    let mut risk_params = data.risk_params.lock().unwrap();
    
    if let Some(balance) = update.account_balance {
        risk_params.account_balance = balance;
    }
    if let Some(pct) = update.position_sizing_pct {
        risk_params.position_sizing_pct = pct;
    }
    if let Some(pct) = update.daily_loss_limit_pct {
        risk_params.daily_loss_limit_pct = pct;
    }
    if let Some(pct) = update.monthly_drawdown_pct {
        risk_params.monthly_drawdown_pct = pct;
    }
    
    HttpResponse::Ok().body("Risk parameters updated")
}

async fn get_circuit_breaker_status(
    data: web::Data<AppState>,
) -> impl Responder {
    let circuit_breaker = data.circuit_breaker.lock().unwrap();
    match circuit_breaker.is_activated() {
        Ok(status) => HttpResponse::Ok().json(json!({ "activated": status })),
        Err(e) => HttpResponse::InternalServerError().body(e.to_string()),
    }
}

async fn activate_circuit_breaker(
    data: web::Data<AppState>,
) -> impl Responder {
    let circuit_breaker = data.circuit_breaker.lock().unwrap();
    match circuit_breaker.activate() {
        Ok(_) => {
            let telegram_alert = data.telegram_alert.lock().unwrap();
            if let Err(e) = telegram_alert.circuit_breaker_activated().await {
                log::error!("Failed to send Telegram alert: {}", e);
            }
            HttpResponse::Ok().body("Circuit breaker activated")
        },
        Err(e) => HttpResponse::InternalServerError().body(e.to_string()),
    }
}

async fn deactivate_circuit_breaker(
    data: web::Data<AppState>,
) -> impl Responder {
    let circuit_breaker = data.circuit_breaker.lock().unwrap();
    match circuit_breaker.deactivate() {
        Ok(_) => {
            let telegram_alert = data.telegram_alert.lock().unwrap();
            if let Err(e) = telegram_alert.circuit_breaker_deactivated().await {
                log::error!("Failed to send Telegram alert: {}", e);
            }
            HttpResponse::Ok().body("Circuit breaker deactivated")
        },
        Err(e) => HttpResponse::InternalServerError().body(e.to_string()),
    }
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    dotenv::dotenv().ok();
    env_logger::init();

    let account_balance = 2000.0; // Default balance, should come from config
    let risk_params = Arc::new(Mutex::new(RiskParameters::new(account_balance)));
    let redis_state = RedisState::new().expect("Failed to connect to Redis");
    let circuit_breaker = Arc::new(Mutex::new(CircuitBreaker::new(
        risk_params.lock().unwrap().clone(),
        redis_state,
    )));
    let telegram_alert = Arc::new(Mutex::new(
        TelegramAlert::new().expect("Failed to initialize Telegram alerts")
    ));

    let app_state = web::Data::new(AppState {
        risk_params: Arc::clone(&risk_params),
        circuit_breaker: Arc::clone(&circuit_breaker),
        telegram_alert: Arc::clone(&telegram_alert),
    });

    HttpServer::new(move || {
        App::new()
            .app_data(app_state.clone())
            .route("/health", web::get().to(health_check))
            .route("/validate", web::post().to(validate_trade))
            .route("/risk_params", web::put().to(update_risk_params))
            .route("/circuit_breaker", web::get().to(get_circuit_breaker_status))
            .route("/circuit_breaker/activate", web::post().to(activate_circuit_breaker))
            .route("/circuit_breaker/deactivate", web::post().to(deactivate_circuit_breaker))
    })
    .bind("127.0.0.1:8080")?
    .run()
    .await
}