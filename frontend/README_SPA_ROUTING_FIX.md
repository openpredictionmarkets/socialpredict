# SPA Routing Fix for Production

## Problem
Direct navigation to routes like `/markets` or `/user/{username}` in production resulted in 404 errors, while navigation from the home page worked fine.

## Root Cause
The production setup was missing proper Single Page Application (SPA) fallback configuration in nginx. When users directly visited routes like `/markets`, the server tried to find a file at that path instead of serving the React app's `index.html` and letting React Router handle the routing.

## Solution
Added a custom nginx configuration (`nginx.conf`) to the frontend container with the essential `try_files` directive:

```nginx
location / {
    try_files $uri $uri/ /index.html;
}
```

This ensures that:
1. nginx first tries to serve the requested URI as a file
2. If that fails, tries to serve it as a directory
3. If that also fails, falls back to serving `/index.html`
4. React Router then handles the client-side routing

## Files Modified
- `frontend/nginx.conf` - New custom nginx configuration for SPA routing
- `frontend/Dockerfile.prod` - Updated to copy the custom nginx config

## How It Works
- Static assets (JS, CSS, images) are served directly with caching headers
- All other requests fall back to `index.html`, allowing React Router to handle routing
- API requests are still handled by the main nginx proxy configuration

## Testing
After rebuilding and deploying:
- Direct navigation to `/markets` should work
- Direct navigation to `/user/{username}` should work  
- Clicking on market creator links should work
- All existing functionality should continue to work

## Related Files
- `data/nginx/vhosts/prod/app.conf.template` - Main nginx proxy configuration
- `scripts/prod/build_prod.sh` - Production build script
