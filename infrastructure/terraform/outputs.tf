output "timescaledb_endpoint" {
  description = "TimescaleDB connection endpoint"
  value       = aws_db_instance.timescaledb.endpoint
}

output "redis_endpoint" {
  description = "Redis connection endpoint"
  value       = aws_elasticache_cluster.redis.cache_nodes[0].address
}

output "kafka_bootstrap_brokers" {
  description = "Kafka bootstrap brokers"
  value       = aws_msk_cluster.kafka.bootstrap_brokers
}

output "vault_endpoints" {
  description = "Vault instance endpoints"
  value       = aws_instance.vault[*].private_dns
}

output "vault_lb_endpoint" {
  description = "Vault load balancer endpoint"
  value       = aws_lb.vault.dns_name
}