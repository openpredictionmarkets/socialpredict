import React, { useState, useEffect, useRef } from 'react';
import { useHistory } from 'react-router-dom';
import Navlink from './Navlink';

function Navbar({ onLogout }) {
  const [bgColor, setBgColor] = useState('bg-transparent');
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const menuRef = useRef();
  const history = useHistory();

  useEffect(() => {
    const changeNavbarColor = () => {
      if (window.scrollY > 40) {
        setBgColor('bg-[#1d8cf8]');
      } else {
        setBgColor('bg-transparent');
      }
    };

    window.addEventListener('scroll', changeNavbarColor);

    const handleClickOutside = (event) => {
      if (menuRef.current && !menuRef.current.contains(event.target)) {
        setIsMenuOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);

    return () => {
      window.removeEventListener('scroll', changeNavbarColor);
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  const handleLogoutClick = () => {
    if (onLogout) {
      onLogout();
    }
    history.push('/markets');
    setIsMenuOpen(false); // Ensure the menu is closed on logout
  };

  const toggleMenu = () => {
    setIsMenuOpen(!isMenuOpen);
  };

  const closeMenu = () => {
    setIsMenuOpen(false);
  };

  const linkStyle =
    'transition-colors px-2 rounded-lg duration-150 ease-in-out hover:text-gray-200';

  return (
    <div
      className={`flex items-center justify-between w-full ${bgColor} fixed top-0 z-50 text-xl transition-colors duration-300 ease-in-out px-4`}
      ref={menuRef}
    >
      <button onClick={toggleMenu} className='z-50 px-4 md:hidden'></button>

      <div className='flex items-center justify-start flex-grow gap-8'>
        <Navlink
          to='/profile'
          className={linkStyle}
          text='Profile'
          onClick={closeMenu}
        />
        <Navlink
          to='/markets'
          className={linkStyle}
          text='Markets'
          onClick={closeMenu}
        />
        <Navlink
          to='/polls'
          className={linkStyle}
          text='Polls'
          onClick={closeMenu}
        />
      </div>

      <div className='flex justify-center flex-grow-0'>
        <Navlink
          to='/create'
          className={linkStyle}
          text='Create'
          onClick={closeMenu}
        />
      </div>

      <div className='flex items-center justify-end flex-grow gap-8'>
        <Navlink
          to='/notifications'
          className={linkStyle}
          text='Notifications'
          onClick={closeMenu}
        />
        <Navlink
          to='/about'
          className={linkStyle}
          text='About'
          onClick={closeMenu}
        />
        <Navlink
          to='/'
          className={linkStyle}
          text='Logout'
          onClick={handleLogoutClick}
        />
      </div>
    </div>
  );
}

export default Navbar;
