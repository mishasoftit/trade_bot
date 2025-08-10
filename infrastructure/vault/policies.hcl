# API Key Encryption Policy
path "transit/encrypt/api-keys" {
  capabilities = ["update"]
}

path "transit/decrypt/api-keys" {
  capabilities = ["update"]
}

# Database Dynamic Secrets Policy
path "database/creds/*" {
  capabilities = ["read"]
}

# Certificate Management Policy
path "pki/issue/*" {
  capabilities = ["create", "update"]
}

# Secrets Management Policy
path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}