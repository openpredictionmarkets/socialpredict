import React from 'react';
import { Link, useHistory } from 'react-router-dom';
import './Navbar.css'; // Make sure to create this CSS file


function Navbar({onLogout}) {

  const history = useHistory();

  const handleLogoutClick = () => {
    if (onLogout) {
      onLogout(); // Call the onLogout function
    }
    history.push('/markets'); // Redirect after logout
  };


  return (
    <nav className="navbar">
      <Link to="/profile" className="nav-link">Profile</Link>
      <Link to="/markets" className="nav-link">Markets</Link>
      <Link to="/polls" className="nav-link">Polls</Link>
      <Link to="/notifications" className="nav-link">Notifications</Link>
      <Link to="/create" className="nav-link nav-link-callout">Create</Link>
      <Link to="/about" className="nav-link">About</Link>
      <Link to="/" onClick={handleLogoutClick} className="nav-link">Logout</Link>

    </nav>
  );
}

export default Navbar;
