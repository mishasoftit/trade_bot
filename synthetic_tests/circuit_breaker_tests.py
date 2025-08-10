import time
import requests
import logging

# Configuration
PROMETHEUS_URL = "http://prometheus:9090"
ALERTMANAGER_URL = "http://alertmanager:9093"
RISK_SERVICE_URL = "http://risk-controller:8080"
STRATEGY_SERVICE_URL = "http://strategy-service:8000"

# Test parameters
TEST_INTERVAL = 60  # seconds
CIRCUIT_BREAKER_TRIGGER_DURATION = 10  # seconds

def trigger_circuit_breaker(service):
    """Simulate conditions to trigger circuit breaker"""
    try:
        if service == "risk":
            # Simulate high risk exposure
            response = requests.post(
                f"{RISK_SERVICE_URL}/simulate/exposure",
                json={"exposure": 95, "duration": CIRCUIT_BREAKER_TRIGGER_DURATION}
            )
        elif service == "strategy":
            # Simulate high error rate
            response = requests.post(
                f"{STRATEGY_SERVICE_URL}/simulate/errors",
                json={"error_rate": 0.5, "duration": CIRCUIT_BREAKER_TRIGGER_DURATION}
            )
        return response.status_code == 200
    except Exception as e:
        logging.error(f"Error triggering circuit breaker: {e}")
        return False

def verify_alert(service):
    """Check if alert was generated"""
    try:
        # Check Alertmanager for active alerts
        alerts = requests.get(f"{ALERTMANAGER_URL}/api/v2/alerts").json()
        for alert in alerts:
            if alert["labels"]["alertname"] == "CircuitBreakerTriggered" and \
               alert["labels"]["service"] == service and \
               alert["status"]["state"] == "active":
                return True
        
        # Check Prometheus for firing alerts
        prom_query = f'ALERTS{{alertname="CircuitBreakerTriggered", service="{service}"}}'
        response = requests.get(
            f"{PROMETHEUS_URL}/api/v1/query",
            params={"query": prom_query}
        )
        results = response.json()["data"]["result"]
        return len(results) > 0
    except Exception as e:
        logging.error(f"Error verifying alert: {e}")
        return False

def run_tests():
    logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
    
    while True:
        logging.info("Starting circuit breaker tests...")
        
        # Test risk controller circuit breaker
        if trigger_circuit_breaker("risk"):
            time.sleep(5)  # Wait for alert propagation
            if verify_alert("risk-controller"):
                logging.info("Risk controller circuit breaker test PASSED")
            else:
                logging.error("Risk controller circuit breaker test FAILED - alert not received")
        else:
            logging.error("Failed to trigger risk controller circuit breaker")
        
        # Test strategy service circuit breaker
        if trigger_circuit_breaker("strategy"):
            time.sleep(5)  # Wait for alert propagation
            if verify_alert("strategy-service"):
                logging.info("Strategy service circuit breaker test PASSED")
            else:
                logging.error("Strategy service circuit breaker test FAILED - alert not received")
        else:
            logging.error("Failed to trigger strategy service circuit breaker")
        
        logging.info(f"Waiting {TEST_INTERVAL} seconds before next test...")
        time.sleep(TEST_INTERVAL)

if __name__ == "__main__":
    run_tests()