import React from 'react';
import { Link } from 'react-router-dom';
import Logo from '../logo/Logo.jsx';
import LoginModalButton from '../modals/login/LoginModalClick.jsx';

// This function returns menu items, we can use it in both Sidebar and Header components
const MenuItems = ({ isLoggedIn, onLogout }) => {
    const handleLogoutClick = () => {
        if (onLogout) {
            onLogout();
        }
    };

    return (
        <ul className="flex space-x-4">
            {isLoggedIn ? (
                // Logged In Menu Items
                <>
                    <li>
                        <Logo />
                    </li>
                    <li>
                        <Link to="/profile" className="header-link">Profile</Link>
                    </li>
                    {/* ... other logged in menu items */}
                    <li>
                        <Link to="/" onClick={handleLogoutClick} className="header-link">Logout</Link>
                    </li>
                </>
            ) : (
                // Logged Out Menu Items
                <>
                    <li>
                        <Logo />
                    </li>
                    <li>
                        <LoginModalButton />
                    </li>
                    <li>
                        <Link to="/markets" className="header-link">Markets</Link>
                    </li>
                    <li>
                        <Link to="/polls" className="header-link">Polls</Link>
                    </li>
                    <li>
                        <Link to="/about" className="header-link">About</Link>
                    </li>
                </>
            )}
        </ul>
    );
};

const Header = ({ isLoggedIn, onLogout }) => {
    return (
        <nav className="flex items-center justify-between flex-wrap dark:bg-custom-gray-dark p-6">
            <div className="flex items-center flex-shrink-0 dark:bg-custom-gray-dark mr-6">
                {/* ... Logo and branding */}
            </div>
            <div className="w-full block flex-grow lg:flex lg:items-center lg:w-auto">
                <div className="space-y-2 lg:flex-grow text-white font-medium">
                    {/* Here we reuse the MenuItems function */}
                    <MenuItems isLoggedIn={isLoggedIn} onLogout={onLogout} />
                </div>
                {/* ... other header elements */}
            </div>
        </nav>
    );
};

export default Header;