import React, { useState, useEffect } from 'react';
import { NavLink } from 'react-router-dom';
import { Navbar, Container, Nav } from 'reactstrap';

// Assuming you have your custom CSS
import './Navbar.css';

const NavbarLoggedOut = () => {
  const [color, setColor] = useState('navbar-transparent');

  useEffect(() => {
    const changeColor = () => {
      if (
        document.documentElement.scrollTop > 99 ||
        document.body.scrollTop > 99
      ) {
        setColor('bg-info');
      } else if (
        document.documentElement.scrollTop < 100 ||
        document.body.scrollTop < 100
      ) {
        setColor('navbar-transparent');
      }
    };
    window.addEventListener('scroll', changeColor);
    return function cleanup() {
      window.removeEventListener('scroll', changeColor);
    };
  }, []);

  return (
    <Navbar className={`fixed-top ${color}`} expand='lg'>
      <Container>
        <Nav navbar>
          <NavLink to='/markets' className='nav-link' activeClassName='active'>
            Markets
          </NavLink>
          <NavLink to='/polls' className='nav-link' activeClassName='active'>
            Polls
          </NavLink>
          <NavLink to='/about' className='nav-link' activeClassName='active'>
            About
          </NavLink>
        </Nav>
      </Container>
    </Navbar>
  );
};

export default NavbarLoggedOut;
