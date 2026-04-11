# douyin

This repository contains the combined frontend, backend, and nginx configuration for the Douyin/TikTok clone project.

## Structure

- `front/`: Vue 3 + Vite frontend
- `backend/`: Go + Gin backend
- `nginx/`: nginx reverse proxy configuration

## Notes

- Sensitive local credentials have been replaced with placeholder values before publishing.
- Large local runtime artifacts and caches are excluded from version control.
- If you want to run the project locally, update the backend config files under `backend/config/` with your own database, Redis, and object storage settings.
