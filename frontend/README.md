# SocialPredict Frontend

This is the frontend application for SocialPredict, built with React and Vite.

This gets you started with running only the frontend using Docker.

#### Prerequisites

- **Docker**: Version 20.10 or higher
- **Docker Compose**: Version 2.0 or higher

#### Running with Docker

1. **Build the frontend Docker image:**

   ```bash
   cd /path/to/socialpredict
   cd frontend
   docker build -t socialpredict .
   ```

2. **Run the frontend container:**

   ```bash
   docker run -d -p 5173:5173 socialpredict
   ```

   This will start the Vite development server with the following features:

   - **Port**: Typically runs on `http://localhost:5173` (Vite's default port)
   - **Host**: Accessible from other devices on your network
   - **Hot reload**: Automatic reloading when files change
   - **Fast builds**: Vite's lightning-fast development experience

3. **Access the application:**
   Navigate to `http://localhost:5173` on your browser to view the application.

4. **When ready, stop the container:**
   ```bash
   docker stop $(docker ps -n 1 -a -q)
   ```

### Mobile Responsiveness

The application is fully responsive and includes:

- Mobile-optimized navigation with hamburger menu
- Touch-friendly interface
- Responsive design for all screen sizes

### Tech Stack

- **React 18**: UI framework
- **Vite**: Build tool and development server
- **Tailwind CSS**: Styling
- **React Router**: Client-side routing
- **Chart.js & Recharts**: Data visualization
