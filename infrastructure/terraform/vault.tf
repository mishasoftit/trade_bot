# Vault secrets management configuration
resource "aws_instance" "vault" {
  count         = terraform.workspace == "prod" ? 3 : 1
  ami           = var.ami_id
  instance_type = "t3.medium"
  subnet_id     = element(var.subnet_ids, count.index)
  security_groups = [aws_security_group.vault.id]
  iam_instance_profile = aws_iam_instance_profile.vault.name
  user_data     = templatefile("${path.module}/vault_init.tpl", {
    cluster_name = "vault-${terraform.workspace}"
    node_count   = terraform.workspace == "prod" ? 3 : 1
  })

  tags = {
    Name        = "vault-${terraform.workspace}-${count.index}"
    Environment = terraform.workspace
    Component   = "vault"
  }
}

resource "aws_security_group" "vault" {
  name        = "vault-sg-${terraform.workspace}"
  description = "Vault security group"

  ingress {
    from_port   = 8200
    to_port     = 8200
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  ingress {
    from_port   = 8201
    to_port     = 8201
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

resource "aws_iam_instance_profile" "vault" {
  name = "vault-instance-profile-${terraform.workspace}"
  role = aws_iam_role.vault.name
}

resource "aws_iam_role" "vault" {
  name = "vault-role-${terraform.workspace}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "vault_kms" {
  role       = aws_iam_role.vault.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonKMSFullAccess"
}

resource "aws_iam_role_policy_attachment" "vault_ssm" {
  role       = aws_iam_role.vault.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}