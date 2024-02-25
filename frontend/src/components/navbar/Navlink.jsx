import React from 'react';
import { Link } from 'react-router-dom';

const NavLink = ({ to, className, text, onClick }) => {
  return (
    <Link to={to} className={className} onClick={onClick}>
      {text}
    </Link>
  );
};

export default NavLink;
