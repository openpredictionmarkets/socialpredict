import React from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../../helpers/AuthContent';
import LoginModalButton from '../modals/login/LoginModalClick';

const Sidebar = () => {
    // useAuth hook to get auth state and logout function
    const { isLoggedIn, logout } = useAuth();

    const handleLogoutClick = () => {
        logout();
    };

return (
    <aside className="fixed top-0 left-0 z-sidebar w-sidebar h-screen transition-transform -translate-x-full sm:translate-x-0" aria-label="Sidebar">
    <div className="h-full px-3 py-4 dark:bg-custom-gray-dark">
        <ul className="space-y-2 font-medium">
            {isLoggedIn ? (
            // Logged In Sidebar Items
            <>
            <li>
                <Link to="/profile" className="sidebar-link">Profile</Link>
            </li>
            <li>
                <Link to="/markets" className="sidebar-link">Markets</Link>
            </li>
            <li>
                <Link to="/polls" className="sidebar-link">Polls</Link>
            </li>
            <li>
                <Link to="/notifications" className="sidebar-link">Notifications</Link>
            </li>
            <li>
                <Link to="/create" className="sidebar-link">Create</Link>
            </li>
            <li>
                <Link to="/about" className="sidebar-link">About</Link>
            </li>
            <li>
                <Link to="/" onClick={handleLogoutClick} className="sidebar-link">Logout</Link>
            </li>
            </>
            ) : (
            // Logged Out Sidebar Items
            <>
            <li>
                <LoginModalButton />
            </li>
            <li>
                <Link to="/markets" className="sidebar-link">Markets</Link>
            </li>
            <li>
                <Link to="/polls" className="sidebar-link">Polls</Link>
            </li>
            <li>
                <Link to="/about" className="sidebar-link">About</Link>
            </li>
            </>
        )}
        </ul>
    </div>
    </aside>
);
};

export default Sidebar;