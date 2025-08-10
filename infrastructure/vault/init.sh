#!/bin/bash

# Initialize Vault
vault operator init -key-shares=1 -key-threshold=1 > init.txt
export VAULT_TOKEN=$(grep 'Initial Root Token' init.txt | awk '{print $4}')

# Enable secrets engines
vault secrets enable transit
vault secrets enable pki
vault secrets enable database

# Configure API key encryption
vault write transit/keys/api-keys type="aes256-gcm96"

# Configure PKI for certificate management
vault secrets tune -max-lease-ttl=87600h pki
vault write pki/root/generate/internal common_name="example.com" ttl=87600h
vault write pki/roles/example-dot-com allowed_domains="example.com" allow_subdomains=true max_ttl=72h

# Configure database secrets engine for TimescaleDB
vault write database/config/timescaledb \
  plugin_name=postgresql-database-plugin \
  allowed_roles="readonly,write" \
  connection_url="postgresql://{{username}}:{{password}}@timescaledb:5432/timeseries" \
  username="vault_admin" \
  password="vault_password"

# Create database roles
vault write database/roles/readonly \
  db_name=timescaledb \
  creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; \
      GRANT SELECT ON ALL TABLES IN SCHEMA public TO \"{{name}}\";" \
  default_ttl="1h" \
  max_ttl="24h"

vault write database/roles/write \
  db_name=timescaledb \
  creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; \
      GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO \"{{name}}\";" \
  default_ttl="1h" \
  max_ttl="24h"

# Apply policies
vault policy write trading-policy policies.hcl

# Enable AppRole authentication
vault auth enable approle
vault write auth/approle/role/trading-role \
  token_policies="trading-policy" \
  token_ttl="1h" \
  token_max_ttl="4h"

# Output role ID and secret ID
vault read auth/approle/role/trading-role/role-id
vault write -f auth/approle/role/trading-role/secret-id