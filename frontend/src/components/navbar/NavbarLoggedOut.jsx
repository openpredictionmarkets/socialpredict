import React from 'react';
import { Link } from 'react-router-dom';
import './Navbar.css'; // Make sure to create this CSS file

function NavbarLoggedOut() {
  return (
    <nav className="navbar">
      <Link to="/markets" className="nav-link">Markets</Link>
      <Link to="/polls" className="nav-link">Polls</Link>
      <Link to="/about" className="nav-link">About</Link>
    </nav>
  );
}

export default NavbarLoggedOut;
