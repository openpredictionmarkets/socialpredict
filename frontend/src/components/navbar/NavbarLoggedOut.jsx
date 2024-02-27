import { useState, useEffect, useRef } from 'react';
import Navlink from './Navlink';

function NavbarLoggedOut() {
  const [bgColor, setBgColor] = useState('bg-transparent');
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const menuRef = useRef();

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

  const toggleMenu = () => {
    setIsMenuOpen(!isMenuOpen);
  };

  const closeMenu = () => {
    setIsMenuOpen(false);
  };

  const linkStyle =
    'transition-colors px-2 rounded-lg duration-150 ease-in-out hover:text-gray-200 block py-2';

  return (
    <nav
      className={`flex items-center justify-center md:justify-between w-full ${bgColor} fixed top-0 z-50 text-xl transition-transform duration-150 ease-in-out h-10`}
      ref={menuRef}
    >
      <button
        onClick={toggleMenu}
        className='absolute z-50 px-4 mt-3 mb-2 md:hidden right-5'
      >
        {isMenuOpen ? (
          <svg
            className='w-6 h-6'
            fill='none'
            stroke='currentColor'
            viewBox='0 0 24 24'
            xmlns='http://www.w3.org/2000/svg'
          >
            <path
              strokeLinecap='round'
              strokeLinejoin='round'
              strokeWidth='2'
              d='M6 18L18 6M6 6l12 12'
            />
          </svg>
        ) : (
          <svg
            className='w-6 h-6'
            fill='none'
            stroke='currentColor'
            viewBox='0 0 24 24'
            xmlns='http://www.w3.org/2000/svg'
          >
            <path
              strokeLinecap='round'
              strokeLinejoin='round'
              strokeWidth='2'
              d='M4 6h16M4 12h16m-7 6h7'
            />
          </svg>
        )}
      </button>

      <div
        className={`absolute w-[90%] md:w-full mt-2 rounded-md p-[1.4rem] top-0 bg-[#1f2251] md:relative md:top-auto md:flex ${
          isMenuOpen ? 'flex' : 'hidden'
        } flex-col z-10 md:flex-row transition-transform duration-150 ease-in-out`}
      >
        <Navlink
          to='/markets'
          className={`${linkStyle} md:inline-block`}
          text='Markets'
          onClick={closeMenu}
        />
        <Navlink
          to='/polls'
          className={`${linkStyle} md:inline-block`}
          text='Polls'
          onClick={closeMenu}
        />
        <Navlink
          to='/about'
          className={`${linkStyle} md:inline-block`}
          text='About'
          onClick={closeMenu}
        />
      </div>
    </nav>
  );
}

export default NavbarLoggedOut;
