# Script para adicionar testadores via gcloud CLI

# Instalar gcloud CLI (se não tiver):
# https://cloud.google.com/sdk/docs/install

# 1. Fazer login
gcloud auth login

# 2. Definir projeto
gcloud config set project 645799287936

# 3. Adicionar testador (SUBSTITUA SEU EMAIL)
gcloud alpha iap oauth-clients add-iam-policy-binding OAUTH_CLIENT_ID \
    --member="user:SEU_EMAIL@gmail.com" \
    --role="roles/iap.httpsResourceAccessor"

# Alternativa: Editar diretamente via API
# (Requer configuração adicional)
