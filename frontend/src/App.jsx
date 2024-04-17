import React, { useState } from 'react';
import { BrowserRouter as Router } from 'react-router-dom';
import { AuthProvider } from './helpers/AuthContent';
import Navbar from './components/navbar/Navbar';
import NavbarLoggedOut from './components/navbar/NavbarLoggedOut';
import Footer from './components/footer/Footer';
import AppRoutes from './helpers/AppRoutes';

function App() {
  const [isLoggedIn, setIsLoggedIn] = useState(false);

  return (
    <AuthProvider>
      <Router>
        <div className='App bg-primary-background text-white sm:pl-sidebar sm:pr-sidebar h-screen'>
          <header className='App-header'>
            {isLoggedIn ? <Navbar onLogout={handleLogout} /> : <NavbarLoggedOut />}
            <AppRoutes isLoggedIn={isLoggedIn} onLogin={handleLogin} onLogout={handleLogout} />
          </header>
          <Footer />
        </div>
      </Router>
    </AuthProvider>
  );
}

export default App;