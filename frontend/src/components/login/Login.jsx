import React, { useState, useContext } from 'react';
import { useHistory } from 'react-router-dom';
import './Login.css';
import UserContext from '../../helpers/UserContext';

function Login({ onLogin }) {
  const history = useHistory();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');

  const { setUsername: setContextUsername } = useContext(UserContext);

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError('');
    if (onLogin) {
      try {
        await onLogin(username, password); // Let App.js handle login
        setContextUsername(username); // Set username in context
        console.log('Logged in as:', username); // Log out the username
        history.push('/'); // Redirect after successful login
      } catch (loginError) {
        console.error('Login error:', loginError);
        setError('Incorrect Login'); // Set error message
      }
    } else {
      console.error('onLogin prop is not a function');
    }
  };

  return (
    <form onSubmit={handleSubmit} className='login-form'>
      <div>
        <label>
          Username:
          <input
            type='text'
            value={username}
            onChange={(e) => setUsername(e.target.value)}
          />
        </label>
      </div>
      <div>
        <label>
          Password:
          <input
            type='password'
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </label>
      </div>
      {error && <div className='error-message'>{error}</div>}{' '}
      {/* Display error message */}
      <button type='submit'>Login</button>
    </form>
  );
}

export default Login;
