# Non-secret production/model-office rate-limit overlay for OpenPredictionMarkets.
# Keep this conservative unless fresh capacity evidence supports a change.
RATE_LIMIT_PROFILE=env-file
RATE_LIMIT_LOGIN_RATE_PER_SECOND=0.1
RATE_LIMIT_LOGIN_BURST=3
RATE_LIMIT_GENERAL_RATE_PER_SECOND=1
RATE_LIMIT_GENERAL_BURST=10
RATE_LIMIT_CLEANUP_INTERVAL=5m
