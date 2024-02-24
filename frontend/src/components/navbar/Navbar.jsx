import React, { useState, useEffect } from 'react';
import { Link, useHistory } from 'react-router-dom';

function Navbar({ onLogout }) {
  const [bgColor, setBgColor] = useState('bg-transparent');
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

    return () => {
      window.removeEventListener('scroll', changeNavbarColor);
    };
  }, []);

  const handleLogoutClick = () => {
    if (onLogout) {
      onLogout();
    }
    history.push('/markets');
  };

  return (
    <div
      className={`flex items-center w-full h-10 ${bgColor} fixed top-0 z-50 text-xl transition-colors duration-300 ease-in-out px-4`}
    >
      {/* Left Section */}
      <div className='flex flex-grow items-center justify-start gap-8'>
        <Link to='/profile'>Profile</Link>
        <Link to='/markets'>Markets</Link>
        <Link to='/polls'>Polls</Link>{' '}
        {/* Moved Polls to the left as requested */}
      </div>

      {/* Center Section - Ensure Create is centered by flex-grow on both sides */}
      <div className='flex justify-center flex-grow-0'>
        <Link to='/create' className='text-center'>
          Create
        </Link>
      </div>

      {/* Right Section */}
      <div className='flex flex-grow items-center justify-end gap-8'>
        <Link to='/notifications'>Notifications</Link>
        <Link to='/about'>About</Link>
        <Link to='/' onClick={handleLogoutClick}>
          Logout
        </Link>
      </div>
    </div>
  );
}

export default Navbar;
