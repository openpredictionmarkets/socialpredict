import React from 'react';
import { BrowserRouter as Router } from 'react-router-dom';
import { AuthProvider } from './helpers/AuthContent';
import Footer from './components/footer/Footer';
import AppRoutes from './helpers/AppRoutes';
import '../index.css';
import Sidebar from './components/sidebar/Sidebar';

function App() {
  return (
    <AuthProvider>
      <Router>

        <div className='App bg-primary-background text-white flex h-[calc(100vh-96px)]'>
          <Sidebar />
          <div className='flex flex-col flex-grow'>
            <main className='flex-grow p-4'>
              <AppRoutes />
            </main>
            <Footer />
          </div>
        </div>
      </Router>
    </AuthProvider>
  );
}

export default App;
