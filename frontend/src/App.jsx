import { API_URL } from './config';
import React, { useState, useEffect } from 'react';
import {
  BrowserRouter as Router,
  Route,
  Switch,
  Redirect,
} from 'react-router-dom';
import { AuthProvider } from './helpers//AuthContent';
import Login from './components/login/Login';
import Navbar from './components/navbar/Navbar';
import NavbarLoggedOut from './components/navbar/NavbarLoggedOut';
import Profile from './pages/profile/Profile';
import Markets from './pages/markets/Markets';
import Polls from './pages/polls/Polls';
import Notifications from './pages/notifications/Notifications';
import Create from './pages/create/Create';
import About from './pages/about/About';
import { UserProvider } from './helpers/UserContext';
import MarketDetails from './pages/marketDetails/MarketDetails';
import User from './pages/user/User';
import Footer from './components/footer/Footer';
import Style from './pages/style/Style';
import '../index.css';


function App() {
  // state variables
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [username, setUsername] = useState(null);

  // remove token if logged out
  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('userId');
    localStorage.removeItem('username');
    setIsLoggedIn(false);
    setUsername(null);
  };

  // login function
  const handleLogin = async (username, password) => {
    try {
      const response = await fetch(`${API_URL}/api/v0/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username, password }),
      });

      const responseData = await response.json();
      if (response.ok) {
        const token = responseData.token;
        console.log('JWT Key:', token); // Log the JWT key
        localStorage.setItem('token', token);
        setIsLoggedIn(true);
        setUsername(username); // Set the username
        // Redirect to markets page
      } else {
        // Handle login error
        // You can return error message to show in Login component
      }
    } catch (error) {
      console.error('Login error:', error);
      // Handle network or other errors
    }
  };

  useEffect(() => {
    const token = localStorage.getItem('token');
    setIsLoggedIn(!!token);

    console.log(`Request URL: ${API_URL}/api/v0/home`);
    fetch(`${API_URL}/api/v0/home`)
      .then((response) => response.json())
      .then((data) => {
        console.log('Data Received:', data);
      })
      .catch((error) => console.error('Error fetching data:', error));
  }, []);

  return (
    <AuthProvider>
    <UserProvider value={{ username, setUsername, isLoggedIn }}>
      <Router>
      <div className='App bg-primary-background text-white sm:pl-sidebar sm:pr-sidebar h-screen'>
          <header className='App-header'>
            {isLoggedIn && <Navbar onLogout={handleLogout} />}{' '}
            {/* Render Navbar if Logged In */}
            {!isLoggedIn && <NavbarLoggedOut />} {/* Render if Logged Out */}
            {/* Define Our Router */}
            <Switch>
              <Route exact path='/'>
                {!isLoggedIn ? (
                  <Login onLogin={handleLogin} />
                ) : (
                  <Redirect to='/markets' />
                )}
              </Route>
              <Route path='/profile'>
                {/* Render Profile only if not logged in, else redirect to "/" */}
                {isLoggedIn ? <Profile /> : <Redirect to='/' />}
              </Route>
              {/* Render MarketDetails for individual marketId if either logged in or not*/}
              <Route path='/markets/:marketId' component={MarketDetails} />
              <Route path='/markets'>
                {/* Render Login only if not logged in */}
                {!isLoggedIn && <Login onLogin={handleLogin} />}
                {/* Add the Markets route */}
                <Markets />
              </Route>
              <Route path='/polls'>
                {/* Render Login only if not logged in */}
                {!isLoggedIn && <Login onLogin={handleLogin} />}
                {/* Add the Polls route */}
                <Polls />
              </Route>
              <Route path='/user/:username' component={User} />
              <Route path='/notifications'>
                {/* Render Notifications only if not logged in, else redirect to "/" */}
                {isLoggedIn ? <Notifications /> : <Redirect to='/' />}
              </Route>
              <Route path='/create'>
                {/* Render Create only if not logged in, else redirect to "/" */}
                {isLoggedIn ? <Create /> : <Redirect to='/' />}
              </Route>
              <Route path='/about'>
                {/* Render Login only if not logged in */}
                {!isLoggedIn && <Login onLogin={handleLogin} />}
                <About />
              </Route>
              {/* Define other routes as needed */}
              <Route path="/style">
                <Style />
              </Route>
            </Switch>
          </header>
          <Footer />
        </div>
      </Router>
    </UserProvider>
    </AuthProvider>
  );
}

export default App;
