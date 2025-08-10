# Kafka message bus configuration
resource "aws_msk_cluster" "kafka" {
  cluster_name           = "kafka-${terraform.workspace}"
  kafka_version          = "2.8.1"
  number_of_broker_nodes = terraform.workspace == "prod" ? 3 : 1

  broker_node_group_info {
    instance_type  = "kafka.m5.large"
    ebs_volume_size = 100
    client_subnets = var.subnet_ids
    security_groups = [aws_security_group.kafka.id]
  }

  encryption_info {
    encryption_in_transit {
      client_broker = "TLS_PLAINTEXT"
    }
  }

  tags = {
    Environment = terraform.workspace
    Component   = "kafka"
  }
}

resource "aws_security_group" "kafka" {
  name        = "kafka-sg-${terraform.workspace}"
  description = "Kafka security group"

  ingress {
    from_port   = 9092
    to_port     = 9092
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