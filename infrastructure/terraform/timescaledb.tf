# TimescaleDB cluster configuration
resource "aws_db_instance" "timescaledb" {
  identifier             = "timescaledb-${terraform.workspace}"
  allocated_storage      = 100
  storage_type           = "gp3"
  engine                 = "timescaledb"
  engine_version         = "2.4.1"
  instance_class         = "db.m5.large"
  db_name                = "timeseries"
  username               = var.db_username
  password               = var.db_password
  parameter_group_name   = aws_db_parameter_group.timescaledb.name
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.timescaledb.id]
  publicly_accessible    = false
  multi_az               = terraform.workspace == "prod" ? true : false

  tags = {
    Environment = terraform.workspace
    Component   = "timescaledb"
  }
}

resource "aws_db_parameter_group" "timescaledb" {
  name   = "timescaledb-params-${terraform.workspace}"
  family = "timescaledb2.4"

  parameter {
    name  = "timescaledb.telemetry_level"
    value = "off"
  }
}

resource "aws_security_group" "timescaledb" {
  name        = "timescaledb-sg-${terraform.workspace}"
  description = "TimescaleDB security group"

  ingress {
    from_port   = 5432
    to_port     = 5432
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