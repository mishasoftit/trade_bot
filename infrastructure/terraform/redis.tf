# Redis instance configuration
resource "aws_elasticache_cluster" "redis" {
  cluster_id           = "redis-${terraform.workspace}"
  engine               = "redis"
  node_type            = "cache.m5.large"
  num_cache_nodes      = terraform.workspace == "prod" ? 2 : 1
  parameter_group_name = aws_elasticache_parameter_group.redis.name
  engine_version       = "6.x"
  port                 = 6379
  security_group_ids   = [aws_security_group.redis.id]
  subnet_group_name    = aws_elasticache_subnet_group.redis.name

  tags = {
    Environment = terraform.workspace
    Component   = "redis"
  }
}

resource "aws_elasticache_parameter_group" "redis" {
  name   = "redis-params-${terraform.workspace}"
  family = "redis6.x"

  parameter {
    name  = "maxmemory-policy"
    value = "allkeys-lru"
  }
}

resource "aws_elasticache_subnet_group" "redis" {
  name       = "redis-subnet-${terraform.workspace}"
  subnet_ids = var.subnet_ids
}

resource "aws_security_group" "redis" {
  name        = "redis-sg-${terraform.workspace}"
  description = "Redis security group"

  ingress {
    from_port   = 6379
    to_port     = 6379
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}