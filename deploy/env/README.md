# Deploy Env Overlays

These files are non-secret environment overlays used by the OpenPredictionMarkets Ansible deployment before running `./SocialPredict install`.

They are not replacements for the generated host `.env`. The install command reads the overlay values, validates/writes the rate-limit settings, and the host continues to run from `/opt/socialpredict/.env`.

Files:

- `.env.staging`: high per-IP rate limits for temporary single-source staging load tests from a Mac or one load-generator host.
- `.env.mo`: conservative model-office/production rate limits.

Do not add secrets to these files. Secrets belong in GitHub Actions secrets, HostOps local config, or the generated host `.env`.
